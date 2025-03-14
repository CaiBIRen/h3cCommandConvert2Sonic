//
// Software Name: sonic-netconf-server
// SPDX-FileCopyrightText: Copyright (c) Orange SA
// SPDX-License-Identifier: Apache 2.0
//
// This software is distributed under the Apache 2.0 licence,
// the text of which is available at https://opensource.org/license/apache-2-0/
// or see the "LICENSE" file for more details.
//
// Authors: hossam4.hassan@orange.com, abdelmuhaimen.seaudi@orange.com
// Software description: RFC compliant NETCONF server implementation for SONiC
//

package server

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"sonic-unis-framework/device"
	"sonic-unis-framework/netconf/proxy"
	sonichandlers "sonic-unis-framework/sonichandlers"
	"sonic-unis-framework/tcontext"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/coreos/pkg/capnslog"
	"github.com/gliderlabs/ssh"
)

var sessionID = 0
var glog = capnslog.NewPackageLogger("sonic-unis-framework", "NETCONF_HANDLER")

const (
	delimeter   = "]]>]]>"
	declaration = "<?xml version=\"1.0\" encoding=\"utf-8\"?>"
)

type SessionRequest struct {
	xml           string
	authenticator Authenticator
	session       ssh.Session
}

func SessionHandler(s ssh.Session) {
	scanner := bufio.NewScanner(s)
	scanner.Split(SplitAt)

	glog.Info("session connected, starting main loop")
	for scanner.Scan() {
		requestStr := scanner.Text()
		glog.Infof("\nReceving request <<< %s >>> \n %s \n\n", time.Now().Local().String(), requestStr)
		request := SessionRequest{
			xml: requestStr,
			//authenticator: s.Context().Value("auth").(Authenticator),
			session: s,
		}
		response := process(request)
		glog.Infof("\nSending response <<< %s >>> \n %s \n\n", time.Now().Local().String(), response)
		writeResponse(s, response)
	}
}

func capabilitesXML() []byte {

	var serverHello Hello

	sessionID += 1 // TODO: handle session id out of bounds
	serverHello.SessionID = sessionID
	serverHello.Capabilities = append(serverHello.Capabilities, CapNetconf10)
	//serverHello.Capabilities = append(serverHello.Capabilities, CapNetconf11)

	serverHello.Capabilities = append(serverHello.Capabilities, CapWritableRunning)
	serverHello.Capabilities = append(serverHello.Capabilities, CapXPath)
	serverHello.Capabilities = append(serverHello.Capabilities, CapMonitoring)
	serverHello.Capabilities = append(serverHello.Capabilities, CapStartup)
	serverHello.Capabilities = append(serverHello.Capabilities, CapNameSpace)

	output, _ := xml.Marshal(serverHello)

	return output
}

func process(request SessionRequest) string {

	defer doRecover(request.session, request.xml)
	rpcNode, err := xmlquery.Parse(strings.NewReader(request.xml))

	if err != nil {
		return createErrorResponse(extractMessageId(request.xml), errors.New("[Malformed XML] Unable to parser request string"))
	}

	rootNode := xmlquery.FindOne(rpcNode, "*")

	if rootNode == nil {
		return createErrorResponse(extractMessageId(request.xml), errors.New("[Malformed XML] Root node not found"))
	}

	if rootNode.Data == "hello" {
		response, err := HelloRequestHandler(rpcNode)
		if err != nil {
			return createErrorResponse("0", err)
		}
		return response

	} else {
		//fw只做转发,故在此请求vfw获取response后直接返回,不会再往下面进行
		if device.Devicehdl.Role() == "Firewall" {
			response, err := proxy.Proxyvfw(request.xml)
			if err != nil {
				return createErrorResponse(extractMessageId(request.xml), err)
			}
			return CreateResponse(extractMessageId(request.xml), []byte(response))
		}

		messageId := rootNode.SelectAttr("message-id")
		if messageId == "" {
			return createErrorResponse(extractMessageId(request.xml), errors.New("[Missing data] Unable to read message-id in rpc"))
		}

		//顶层直接处理ospfv3 netconf报文: [1]SONIC对ospfv3支持有限 [2]H3C ospfv3下发格式特殊
		if strings.Contains(request.xml, "<v:network-instances") {
			return CreateResponse(messageId, []byte("ok"))
		}

		response, err := handleRequest(request, rpcNode)
		if err != nil {
			return createErrorResponse(messageId, err)
		}
		return CreateResponse(messageId, []byte(response))
	}
}

func handleRequest(request SessionRequest, rpcXML *xmlquery.Node) (string, error) {

	var response string
	var err error

	typeNode := xmlquery.FindOne(rpcXML, "//*[local-name() = 'rpc']/*") // Get request type

	switch typeNode.Data {
	case "get":
		response, err = GetRequestHandler(rpcXML)
	// case "get-schema":
	//	response, err = GetSchemaHandler(rpcXML)
	case "edit-config":
		response, err = ConfigRequestHandler(rpcXML)
	case "action":
		response, err = ActionRequestHandler(rpcXML)
	case "close-session":
		time.AfterFunc(1*time.Second, func() { request.session.Close() }) // probably a better way to do this ?
		return "ok", nil
	case "save":
		return "ok", nil
	default:
		return "", errors.New("unsupported command")
	}

	if err != nil {
		return "", err
	}

	return response, nil
}

func CreateResponseFromNode(request *xmlquery.Node, responsePayload []byte) string {
	messageId := request.SelectAttr("message-id")
	return CreateResponse(messageId, responsePayload)
}

func CreateResponse(messageId string, responsePayload []byte) string {
	reply := string(responsePayload)
	switch reply {
	case "{}":
		reply = `<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="` + messageId + `"></rpc-reply>`
	case "ok":
		reply = `<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="` + messageId + `"><ok/></rpc-reply>`
	default:
		reply = `<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="` + messageId + `">` + reply + "</rpc-reply>"
		reply = strings.ReplaceAll(reply, "&amp;", "&")
	}

	return reply
}

func writeResponse(session ssh.Session, message string) {
	rspmsg := declaration + message + delimeter
	session.Write([]byte(rspmsg))
}

func writeOkResponse(session ssh.Session, id string) {
	writeResponse(session, CreateResponse(id, []byte("ok")))
}

func createErrorXML(err error) string {
	return fmt.Sprintf("<rpc-error><error-type>rpc</error-type><error-severity>error</error-severity><error-message xml:lang=\"en\">%s</error-message></rpc-error>", err.Error())
}

func createErrorResponse(messageId string, err error) string {
	return CreateResponse(messageId, []byte(createErrorXML(err)))
}

func DefaultHandler(s ssh.Session) {
	fmt.Println("Default ssh is disabled, closing connection")
	s.Close()
}

func HelloRequestHandler(rootNode *xmlquery.Node) (string, error) {
	glog.Infof("Extracted requests %+v", rootNode.Data)
	err := ParseHelloRequest(rootNode)

	if err != nil {
		return "", err
	}
	capabilities := string(capabilitesXML())

	return capabilities, nil
}

// 将请求解析为map[string]struct的格式
func GetRequestHandler(rootNode *xmlquery.Node) (string, error) {
	getc := tcontext.NewTcontext()
	messageid := rootNode.SelectAttr("message-id")
	getc.Messageid = messageid
	getc.Operation = "get"
	featuremap, err := ParseGetRequest(rootNode)
	if err != nil {
		return "", err
	}

	////将本次请求转义并且请求要获取的数据
	err = device.Devicehdl.EncodeGet(featuremap, &getc)
	if err != nil {
		return "", err
	}

	rsp, err := device.Devicehdl.IntegrationReply(&getc)
	if err != nil {
		return "", err
	}
	return rsp, nil
}

func ConfigRequestHandler(rootNode *xmlquery.Node) (string, error) {
	glog.Info("start parse config request")
	addc, delc, err := ParseConfigRequest(rootNode)
	if err != nil {
		glog.Errorf("[parse config request error] %v", err)
		return "", err
	}
	var addrsp, delrsp bool = true, true
	if delc.Err == nil {
		err = device.Devicehdl.EncodeRemove(delc)
		if err != nil {
			glog.Errorf("[encoding remove config error] %v", err)
			return "", err
		}

		delrsp, err = sonichandlers.SonicRemoveConfigHandlers(delc)
		if err != nil {
			glog.Errorf("[sonic delete config error] %v", err)
			return "", err
		}
	}
	if addc.Err == nil {
		err = device.Devicehdl.EncodeMerge(addc)
		if err != nil {
			glog.Errorf("[encoding merge config error] %v", err)
			return "", err
		}

		addrsp, err = sonichandlers.SonicAddConfigHandlers(addc)
		if err != nil {
			glog.Errorf("[sonic merge config error] %v", err)
			return "", err
		}
	}

	if delrsp && addrsp {
		return "ok", nil
	}
	return "", err
}

func ActionRequestHandler(rootNode *xmlquery.Node) (string, error) {
	actionc, err := ParseActionRequest(rootNode)
	if err != nil {
		return "", err
	}
	err = device.Devicehdl.EncodeAction(actionc)
	if err != nil {
		return "", err
	}
	return "ok", nil
}

func SplitAt(data []byte, atEOF bool) (advance int, token []byte, err error) {

	if atEOF && len(data) == 0 || len(trimInput(string(data))) == 0 {
		return 0, nil, nil
	}

	// Find the index of the input of the separator substring
	if i := strings.Index(string(data), RPCDelimiter); i >= 0 {
		return i + len(RPCDelimiter), data[0:i], nil
	}

	if i := strings.Index(string(data), ChunkDelimiter); i >= 0 {
		return i + len(ChunkDelimiter), data[0:i], nil
	}

	// If at end of file with data return the data
	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

func trimInput(input string) string {
	trimmed := strings.Trim(string(input), "\n")
	trimmed = strings.Trim(string(trimmed), "\r")
	trimmed = strings.Trim(string(trimmed), " ")
	return trimmed
}

func doRecover(session ssh.Session, inputStr string) {
	if err := recover(); err != nil {

		buf := make([]byte, 64<<10)
		buf = buf[:runtime.Stack(buf, false)]

		glog.Errorf("Runtime error: panic serving NETCONF request (%s)", inputStr)
		glog.Errorf("Panic data: %v \n\n %s \n\n //Trace end", err, buf)

		errorXML := createErrorXML(errors.New("unknown panic recover"))
		writeResponse(session, CreateResponse(extractMessageId(inputStr), []byte(errorXML)))
		//to peer release connection
		session.Close()
	}
}

func extractMessageId(xmlStr string) string {
	r := regexp.MustCompile("message-id=\"(\\S+)\"")
	matches := r.FindStringSubmatch(xmlStr)
	if len(matches) == 0 {
		return "1"
	}
	return matches[1]
}
