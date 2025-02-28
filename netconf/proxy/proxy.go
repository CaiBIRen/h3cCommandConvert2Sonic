package proxy

import (
	"bytes"
	"encoding/xml"
	"sonic-unis-framework/configuration"
	h3cmodel "sonic-unis-framework/model/h3c"
	"strings"

	"errors"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/coreos/pkg/capnslog"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

var glog = capnslog.NewPackageLogger("sonic-unis-framework", "NETCONF_PROXY")

const NETCONF_DELIM string = "]]>]]>"
const NETCONF_HELLO string = `
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
    <capabilities>
        <capability>urn:ietf:params:netconf:base:1.0</capability>
        <capability>urn:ietf:params:netconf:base:1.1</capability>
        <capability>urn:ietf:params:netconf:capability:writable-running:1.0</capability>
        <capability>urn:ietf:params:netconf:capability:candidate:1.0</capability>
        <capability>urn:ietf:params:netconf:capability:confirmed-commit:1.0</capability>
        <capability>urn:ietf:params:netconf:capability:rollback-on-error:1.0</capability>
        <capability>urn:ietf:params:netconf:capability:startup:1.0</capability>
        <capability>urn:ietf:params:netconf:capability:url:1.0?scheme=http,ftp,file,https,sftp</capability>
        <capability>urn:ietf:params:netconf:capability:validate:1.0</capability>
        <capability>urn:ietf:params:netconf:capability:xpath:1.0</capability>
        <capability>urn:ietf:params:netconf:capability:notification:1.0</capability>
        <capability>urn:liberouter:params:netconf:capability:power-control:1.0</capability>
        <capability>urn:ietf:params:netconf:capability:interleave:1.0</capability>
        <capability>urn:ietf:params:netconf:capability:with-defaults:1.0</capability>
    </capabilities>
</hello>
`

type ProxyResults struct {
	Proxyresults []ProxyResult
}

type ProxyResult struct {
	Hostname string
	Output   string
	Success  bool
}

type Ncclient struct {
	username string
	password string
	hostname string
	key      string
	port     int
	timeout  time.Duration

	sshClient     *ssh.Client
	session       *ssh.Session
	sessionStdin  io.WriteCloser
	sessionStdout io.Reader
}

func (n Ncclient) Hostname() string {
	return n.hostname
}

func (n Ncclient) Close() {
	n.session.Close()
	n.sshClient.Close()
}

func (n Ncclient) SendHello() (io.Reader, error) {
	reader, err := n.Write(NETCONF_HELLO)
	return reader, err
}

// TODO: use the xml module to add/remove rpc related tags
func (n Ncclient) WriteRPC(line string) (io.Reader, error) {
	line = fmt.Sprintf("<rpc>%s</rpc>", line)
	return n.Write(line)
}

func (n Ncclient) Write(line string) (result io.Reader, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = errors.New(r.(string))
		}
	}()

	if _, err := n.sessionStdin.Write([]byte(line + NETCONF_DELIM)); err != nil {
		panic(err)
	}
	finished := make(chan *bytes.Buffer, 1)
	go func() {
		buf1 := bytes.NewBuffer([]byte{})
		buf2 := make([]byte, 1024)
		for {
			n, err := n.sessionStdout.Read(buf2)
			if err != nil && err != io.EOF {
				glog.Errorf("networkproxy  read error %v", err)
				break
			}
			if err == io.EOF {
				glog.Errorf("networkproxy read over %v", err)
				break
			}
			if n > 0 && n <= len(buf2) {
				glog.Infof("networkproxy read msg %s", string(buf2[:n]))
				buf1.Write(buf2[:n])
				if bytes.HasSuffix(buf1.Bytes(), []byte(NETCONF_DELIM)) {
					buf1 = bytes.NewBuffer(bytes.TrimSuffix(buf1.Bytes(), []byte(NETCONF_DELIM)))
					finished <- buf1
					break
				}
				buf2 = make([]byte, 1024)
			}
		}
	}()

	select {
	case result := <-finished:
		return result, err
	case <-time.After(n.timeout):
		panic("Timed out waiting for NETCONF Reply!")
	}
}

func MakeSshClient(username string, password string, hostname string, key string, port int) (*ssh.Client, *ssh.Session, io.WriteCloser, io.Reader) {

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", hostname, strconv.Itoa(port)), config)
	if err != nil {
		panic("Failed to dial:" + hostname + err.Error())
	}

	session, err := client.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		panic(err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		panic(err)
	}
	return client, session, stdin, stdout
}

func (n *Ncclient) Connect() (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = errors.New(r.(string))
		}
	}()
	sshClient, sshSession, sessionStdin, sessionStdout := MakeSshClient(n.username, n.password, n.hostname, n.key, n.port)

	if err := sshSession.RequestSubsystem("netconf"); err != nil {
		// TODO: the command `xml-mode netconf need-trailer` can be executed
		// as a  backup if the netconf subsystem is not available, try that if we fail
		sshClient.Close()
		sshSession.Close()
		panic("Failed to make subsystem request: " + err.Error())
	}
	n.sshClient = sshClient
	n.session = sshSession
	n.sessionStdin = sessionStdin
	n.sessionStdout = sessionStdout
	return err
}

func MakeClient(username string, password string, hostname string, key string, port int) Ncclient {
	nc := new(Ncclient)
	nc.username = username
	nc.password = password
	nc.hostname = hostname
	nc.key = key
	nc.port = port
	nc.timeout = time.Second * 5
	return *nc
}

type NetconfRequest struct {
	Hosts    []string
	Username string
	Password string
	Key      string
	Port     int
	Request  string
}

type NetconfResult struct {
	success bool
	output  io.Reader
	client  *Ncclient
}

func NetconfWork(request string, client *Ncclient) (result *NetconfResult) {
	// Initialize our NetconfResult
	result = new(NetconfResult)
	result.client = client
	// Make sure we can connect
	if err := client.Connect(); err != nil {
		glog.Errorf("NETCONF connect error %v", err)
		result.output = bytes.NewBufferString(err.Error())
		result.success = false
	} else {
		// Ensure we always close our client connections! Then start the NETCONF protocol.
		defer client.Close()
		client.SendHello()
		// Make sure our request gets written to the
		// client.
		if output, err := client.Write(request); err != nil {
			result.output = bytes.NewBufferString(err.Error())
			result.success = false
		} else {
			result.output = output
			result.success = true
		}
	}
	return result
}

func retrieveResults(res *ProxyResults, results chan *NetconfResult, resultCount int) {
	for i := 0; i < resultCount; i++ {
		result := <-results
		buf := new(bytes.Buffer)
		buf.ReadFrom(result.output)
		output := buf.String()
		var resp ProxyResult
		resp.Hostname = result.client.Hostname()
		resp.Output = output
		resp.Success = result.success
		res.Proxyresults = append(res.Proxyresults, resp)
	}
}

func newNetconfRequest(requestbody string) *NetconfRequest {

	request := new(NetconfRequest)

	configuration.ServiceConfiguration.Configmux.RLock()
	defer configuration.ServiceConfiguration.Configmux.RUnlock()

	for _, v := range configuration.ServiceConfiguration.Vfws {
		request.Hosts = append(request.Hosts, v.IP)
		request.Username = v.Username
		request.Password = v.Password
	}
	request.Request = requestbody

	if request.Request == "" {
		panic(errors.New("received an empty request!"))
	}

	request.Port = 830
	return request
}

func errRecovery() {
	if err := recover(); err != nil {
		fmt.Printf("dsadasdasdasd %v", err)
	}
}

func NetconfWorker(request string, client *Ncclient) *NetconfResult {
	return NetconfWork(request, client)
}

func NetconfHandler(requestbody string) (res *ProxyResults) {
	defer errRecovery()
	n := newNetconfRequest(requestbody)
	glog.Infof("Received a request to run '%s' on %d hosts", n.Request, len(n.Hosts))

	results := make(chan *NetconfResult, len(n.Hosts))

	for _, host := range n.Hosts {
		client := MakeClient(n.Username, n.Password, host, n.Key, n.Port)
		go func() {
			results <- NetconfWorker(n.Request, &client)
		}()
	}
	// Block while read in results, and write them out
	// to our client.
	res = &ProxyResults{Proxyresults: make([]ProxyResult, 0)}
	retrieveResults(res, results, len(n.Hosts))
	return
}

func Proxyvfw(request string) (string, error) {
	var replydataprefix string = "<data>"
	var replytopprefix string = "<top xmlns=\"http://www.h3c.com/netconf/data:1.0\">"
	var replytopsuffidx string = "</top>"
	var replydatasuffix string = "</data>"
	var dataxml string
	if strings.Contains(request, "ContextInformations") && strings.Contains(request, "config") {
		err := ContextIDSave(request)
		if err != nil {
			return "", errors.New("parse contextinformation ID error")
		}
		return "ok", nil
	}
	if strings.Contains(request, "ContextInformations") && strings.Contains(request, "filter") {
		reply, err := GetContextID(request)
		if err != nil {
			return "", errors.New("get contextinformation ID error")
		}
		return replydataprefix + replytopprefix + reply + replytopsuffidx + replydatasuffix, nil
	}
	if strings.Contains(request, "Device") && strings.Contains(request, "filter") && strings.Contains(request, "LAGG") {
		vfwrsp := NetconfHandler(request)
		if len(vfwrsp.Proxyresults) > 0 {
			v := vfwrsp.Proxyresults[0]
			if !v.Success {
				return "", errors.New("vfw reply false to sonic-fw")
			}
			rpcNode, err := xmlquery.Parse(strings.NewReader(v.Output))
			if err != nil {
				glog.Errorf("xmlparse error %v", err)
				return "", errors.New(fmt.Sprintf("[Proxyvfw] sonic-fw internal error1"))
			}
			bodynode := xmlquery.FindOne(rpcNode, "//data/*")
			if bodynode == nil {
				glog.Error("//data/* xmlparse nil")
				return replydataprefix + replydatasuffix, nil
			}
			featurenode := xmlquery.Find(bodynode, "./*")
			var reply string
			for _, node := range featurenode {
				reply += node.OutputXML(true)
			}
			return replydataprefix + replytopprefix + string(reply) + replytopsuffidx + replydatasuffix, nil
		} else {
			return "", errors.New("vfw reply false to sonic-fw")
		}
	}

	//对vfw防火墙的save
	if strings.Contains(request, "save") && strings.Contains(request, "file") {
		NetconfHandler(request)
		return "ok", nil
	}

	//除去虚墙相关的netconf配置，其余[查询]都转发到虚墙上然后汇总结果,默认当前实墙不下任何配置
	if strings.Contains(request, "filter") && strings.Contains(request, "get") {
		featureNode := DistinguishNode(request)
		if featureNode == nil {
			glog.Error("netconf request error,please check request xml")
			return "", errors.New("illegal netconf request message")
		}
		tableNode := xmlquery.Find(featureNode, "./*")
		if len(tableNode) == 0 {
			glog.Error("netconf request error,please check request xml")
			return "", errors.New("illegal netconf request message")
		}
		//request to vfw
		rsp := NetconfHandler(request)
		replymap := make(map[string]*xmlquery.Node, 0)
		for _, v := range rsp.Proxyresults {
			if !v.Success {
				return "", errors.New("vfw reply false to sonic-fw")
			}
			rpcNode, err := xmlquery.Parse(strings.NewReader(v.Output))
			if err != nil {
				glog.Errorf("xmlparse error %v", err)
				return "", errors.New(fmt.Sprintf("[Proxyvfw] sonic-fw internal error1"))
			}
			errornode := xmlquery.FindOne(rpcNode, "//rpc-error")
			if errornode != nil {
				glog.Error("vfw return rpc error")
				return "", errors.New("[Proxyvfw] vfw reply rpc-error")
			}
			datanode := xmlquery.FindOne(rpcNode, "//data")
			if datanode == nil {
				glog.Error("xmlparse error datanode")
				return "", errors.New("[Proxyvfw] sonic-fw parse datanode failed")
			}
			bodynode := xmlquery.FindOne(rpcNode, "//data/*")
			if bodynode == nil {
				glog.Error("//data/* xmlparse nil")
				continue
			}
			ParseingRPCReplyXML(replymap, bodynode)
		}
		dataxml = Replymap2XML(featureNode.Data, replymap)
		if dataxml == "" {
			return replydataprefix + replydatasuffix, nil
		}
		return replydataprefix + replytopprefix + dataxml + replytopsuffidx + replydatasuffix, nil

	}
	return "", errors.New("firewall not support this operation yet")
}

func GetContextID(request string) (string, error) {
	xmlNode, err := xmlquery.Parse(strings.NewReader(request))
	if err != nil {
		return "", errors.New("[Proxyvfw] parsr config xml error")
	}
	contextNode := xmlquery.FindOne(xmlNode, "//Context")
	if contextNode == nil {
		return "", errors.New("[Proxyvfw] get context name error")
	}

	nameNode := xmlquery.FindOne(xmlNode, "//Name")
	if nameNode == nil {
		return "", errors.New("[Proxyvfw] get context name error")
	}

	configuration.ServiceConfiguration.Configmux.RLock()
	contextid := viper.GetString(nameNode.Data)
	configuration.ServiceConfiguration.Configmux.RUnlock()

	var context h3cmodel.Context
	err = xml.Unmarshal([]byte(contextNode.OutputXML(true)), &context)
	if err != nil {
		return "", errors.New("[GetContextID] context xml parse error")
	}

	context.ContextInformations.Name = nameNode.InnerText()
	context.ContextInformations.ContextID = contextid
	newXmlData, err := xml.MarshalIndent(context, "", "  ")
	if err != nil {
		return "", errors.New("[GetContextID] context xml make error")
	}
	return string(newXmlData), nil

}
func ContextIDSave(request string) error {
	xmlNode, err := xmlquery.Parse(strings.NewReader(request))
	if err != nil {
		return errors.New("[Proxyvfw] parsr config xml error")
	}
	contextIDNode := xmlquery.FindOne(xmlNode, "//ContextID")
	if contextIDNode == nil {
		return errors.New("[Proxyvfw] get contextid error ")
	}
	nameNode := xmlquery.FindOne(xmlNode, "//Name")
	if nameNode == nil {
		return errors.New("[Proxyvfw] get context name error")
	}

	configuration.ViperMutexWriteConfig(nameNode.InnerText(), contextIDNode.InnerText())
	return nil
}

func DistinguishNode(request string) *xmlquery.Node {
	rpcNode, err := xmlquery.Parse(strings.NewReader(request))
	if err != nil {
		glog.Errorf("xmlparse error %v", err)
		return nil
	}
	topNode := xmlquery.FindOne(rpcNode, "//top")
	if topNode == nil {
		glog.Errorf("topNode not found %v", err)
		return nil
	}
	featureNode := xmlquery.FindOne(topNode, "./*")
	return featureNode
}

// 对xml报文进行整理,
func ParseingRPCReplyXML(xmlmap map[string]*xmlquery.Node, replynode *xmlquery.Node) {
	feature := xmlquery.FindOne(replynode, "./*")
	tables := xmlquery.Find(feature, "./*")
	for _, tablenode := range tables {
		if _, ok := xmlmap[tablenode.Data]; !ok {
			xmlmap[tablenode.Data] = tablenode
			continue
		}
		rows := xmlquery.Find(tablenode, "./*")
		for _, f := range rows {
			xmlquery.AddChild(xmlmap[tablenode.Data], f)
		}
	}
}

func Replymap2XML(featurename string, xmlmap map[string]*xmlquery.Node) string {
	var data string
	if len(xmlmap) == 0 {
		return ""
	}
	for _, v := range xmlmap {
		data += v.OutputXML(true)
	}
	return "<" + featurename + ">" + data + "</" + featurename + ">"
}
