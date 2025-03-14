package sshserver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sonic-unis-framework/configuration"
	"strings"
	"time"

	"github.com/coreos/pkg/capnslog"
	"github.com/creack/pty"
	sshsrv "github.com/gliderlabs/ssh"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

var glog = capnslog.NewPackageLogger("sonic-unis-framework", "SSH_SERVER")
var UserView, SystemView = ">", "]"

func SessionHandler(s sshsrv.Session) {
	defer DosshRecover(s)
	bashc := exec.Command("bash")
	pty, err := pty.Start(bashc)
	if err != nil {
		glog.Errorf("Failed to start command: %v\n", err)
		return
	}
	defer pty.Close()
	scanner := bufio.NewScanner(s)
	glog.Info("ssh server session connected, starting main loop")

	var ClientRequestChan = make(chan []byte, 512)
	var CloseVfwSubChan = make(chan struct{}, 1)
	//0 - disconnected 1 - connected 2 - beclosed
	var ConnectStatusChan = make(chan int, 3)
	var connectedvfw bool
	for scanner.Scan() {
		requestStr := scanner.Text()
		glog.Infof("\nReceving request <<< %s >>> \n %s \n\n", time.Now().Local().String(), requestStr)
		//对return命令甄别, return命令之前的exit命令  对本程序来说可能会导致vfw connect中断;对于控制器来说会涉及切换实墙、虚墙视图
		if strings.Contains(requestStr, "return") {
			select {
			case <-ConnectStatusChan:
				connectedvfw = false
				// 清理 vfw SSH connection
				CloseVfwSubChan <- struct{}{}
				glog.Infof("connection vfw be closed")
			case <-time.After(1 * time.Second): // 等待 ConnectStatusChan 消息，最多等待1秒
				// 如果5秒内没有收到 ConnectStatusChan 的消息，则继续执行
				glog.Infof("connection vfw ok")
			}
		}

		//未连接vfw
		if !connectedvfw {
			if strings.HasPrefix(requestStr, "sys") {
				Doresponse(s, SystemView)
			} else if strings.Contains(requestStr, "CORE-DRIVER") {
				Doresponse(s, requestStr+UserView)
			} else if strings.Contains(requestStr, "exit") {
				Doresponse(s, SystemView)
			} else if strings.Contains(requestStr, "return") {
				Doresponse(s, UserView)
			} else if strings.HasPrefix(requestStr, "switchto context") {
				vfwname := strings.Split(requestStr, " ")[2]
				info, err := FindSetVfwInfo(vfwname)
				if err != nil {
					Doresponse(s, "no vfw can be used for export")
				} else {
					go ConnectVfw(info, s, ClientRequestChan, ConnectStatusChan, CloseVfwSubChan)
					//阻塞等待返回连接结果
					select {
					case status := <-ConnectStatusChan:
						switch status {
						case 0:
							Doresponse(s, "connecting vfw error")
						case 1:
							connectedvfw = true
							glog.Infof("connected vfw %s", vfwname)
						}
					case <-time.After(30 * time.Second):
						Doresponse(s, "connecting vfw time out")
					}
				}
				//TODO:GNS模拟器中vfw删除流程思考
			} else if strings.Contains(requestStr, "undo context") {
				vfwname := strings.Split(requestStr, " ")[2]
				info, err := FindSetVfwInfo(vfwname)
				if err != nil {
					Doresponse(s, "export vfw not found")
				} else {
					go ConnectVfw(info, s, ClientRequestChan, ConnectStatusChan, CloseVfwSubChan)
					//阻塞等待返回连接结果
					select {
					case status := <-ConnectStatusChan:
						switch status {
						case 0:
							Doresponse(s, "connecting vfw error")
						case 1:
							connectedvfw = true
							glog.Infof("connected vfw %s", vfwname)
						}
					case <-time.After(30 * time.Second):
						Doresponse(s, "connecting vfw time out")
					}
				}
				if connectedvfw {
					FindUnsetVfwInfo(vfwname)
					ResetVfwConfigAndReboot(ClientRequestChan)
					Doresponse(s, SystemView)
				}
			} else if requestStr == "Y" {
				Doresponse(s, SystemView)
			}
		} else if connectedvfw {
			ClientRequestChan <- []byte(requestStr + "\n")
		} else {
			glog.Infof("receive command unexpected %s", requestStr)
			s.Write([]byte(fmt.Sprintf("receive command unexpected %s", requestStr)))
		}
	}
	//关闭连接清理connect协程
	close(CloseVfwSubChan)

	glog.Infof("client ssh session close %s", s.RemoteAddr().String())
}

// 后续考虑更优实现
func ResetVfwConfigAndReboot(rc chan []byte) {
	rc <- []byte("startup saved-configuration init.cfg" + "\n")
	time.Sleep(3 * time.Second)
	rc <- []byte("reboot" + "\n")
	time.Sleep(3 * time.Second)
	rc <- []byte("N" + "\n")
	time.Sleep(1 * time.Second)
	rc <- []byte("Y" + "\n")
	time.Sleep(1 * time.Second)
}

func ConnectVfw(info configuration.Vfwinfo, cs sshsrv.Session, rc chan []byte, csc chan int, clc chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &ssh.ClientConfig{
		User: info.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(info.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:22", info.IP)
	vfwclient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		glog.Errorf("failed to dial: %v", err)
		csc <- 0
		return
	}
	vfwsession, err := vfwclient.NewSession()
	if err != nil {
		glog.Errorf("connect vfw new ssh session error %v", err)
		csc <- 0
		return
	}
	fd := int(os.Stdin.Fd())
	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		glog.Errorf("term get size error %v", err)
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := vfwsession.RequestPty("xterm", termHeight, termWidth, modes); err != nil {
		glog.Errorf("RequestPty error %v", err)
		vfwclient.Close()
		vfwsession.Close()
		csc <- 0
		return
	}
	vfwstdin, _ := vfwsession.StdinPipe()
	vfwstdout, _ := vfwsession.StdoutPipe()
	// 启动远程shell
	if err = vfwsession.Shell(); err != nil {
		glog.Errorf("start shell error: %s", err.Error())
		vfwsession.Close()
		vfwclient.Close()
		csc <- 0
		return
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				glog.Info("ReadChannelG exit")
				return
			default:
				buf := make([]byte, 1024)
				n, err := vfwstdout.Read(buf)
				if err != nil && err != io.EOF {
					glog.Errorf("ReadChannelG read error %v", err)
				}
				if err == io.EOF {
					glog.Errorf("ReadChannelG read error %v", err)
					//此处可以检测到vfw connection closed
					csc <- 2
					cs.Write([]byte(SystemView))
					return
				}
				if n > 0 && n <= len(buf) {
					glog.Debugf("ReadChannelG read msg %s", string(buf[:n]))
					cs.Write(buf[:n])
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case v, ok := <-rc:
				if !ok {
					glog.Errorf("ClientRequestChan close!!")
					return
				}
				_, err := vfwstdin.Write(v)
				if err != nil {
					glog.Errorf("WriteChannelG write error %v", err)
				}
			case <-ctx.Done():
				glog.Info("WriteChannelG exit")
				return
			}
		}
	}()
	csc <- 1
	<-clc
	cancel()
	if err = vfwsession.Wait(); err != nil {
		glog.Errorf("return error: %s", err.Error())
	}
	vfwsession.Close()
	vfwclient.Close()
}

// func ConnectionClean() {
// 	if vfwsession != nil {
// 		vfwsession.Close()
// 	}
// 	if vfwclient != nil {
// 		vfwclient.Close()
// 	}
// 	vfwstdin = nil
// 	vfwstdout = nil
// }

func Doresponse(W io.Writer, w string) {
	_, err := W.Write([]byte(w))
	if err != nil {
		glog.Errorf("\nSending response error <<< %s >>> \n %s \n\n", time.Now().Local().String(), err.Error())
	}
	glog.Infof("\nSending response <<< %s >>> \n %s \n\n", time.Now().Local().String(), w)
}

func DosshRecover(s sshsrv.Session) {
	if err := recover(); err != nil {
		glog.Errorf("ssh server handler panic :  %v\n", err)
		Doresponse(s, "unable to handle command")
	}
}

func FindSetVfwInfo(name string) (configuration.Vfwinfo, error) {
	var Writeconfig bool
	var vinfo configuration.Vfwinfo
	configuration.ServiceConfiguration.Configmux.RLock()
	defer configuration.ServiceConfiguration.Configmux.RUnlock()
	for k, v := range configuration.ServiceConfiguration.Vfws {
		//Find a vfw of the corresponding exit or unused vfirewall info
		if v.Name == name {
			vinfo = v
			break
		}
		if v.Name == "" {
			Writeconfig = true
			configuration.ServiceConfiguration.Vfws[k].Name = name
			vinfo = configuration.ServiceConfiguration.Vfws[k]
			break
		}
	}
	if Writeconfig {
		viper.Set("Vfws", configuration.ServiceConfiguration.Vfws)
		viper.WriteConfig()
		glog.Info("config vfws info changed")
	}
	if vinfo.IP != "" {
		return vinfo, nil
	} else {
		return vinfo, errors.New("vfw not exist")
	}
}

// 清理资源
// 将vfwinfo与export信息解绑
func FindUnsetVfwInfo(name string) {
	var Writeconfig bool
	configuration.ServiceConfiguration.Configmux.RLock()
	defer configuration.ServiceConfiguration.Configmux.RUnlock()
	for k, v := range configuration.ServiceConfiguration.Vfws {
		//Find a vfw of the corresponding exit or unused vfirewall info
		if v.Name == name {
			Writeconfig = true
			configuration.ServiceConfiguration.Vfws[k].Name = ""
			break
		}
	}
	if Writeconfig {
		viper.Set("Vfws", configuration.ServiceConfiguration.Vfws)
		viper.WriteConfig()
		glog.Info("config vfws info changed")
	}
}

// func ReadChannelG(ctx context.Context, r io.Reader, cs sshsrv.Session, wg sync.WaitGroup) {

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			glog.Info("WriteChannelG exit")
// 			return
// 		default:
// 			buf := make([]byte, 1024)
// 			n, err := r.Read(buf)
// 			if err != nil && err != io.EOF {
// 				break
// 			}
// 			if err == io.EOF || n == 0 {
// 				break
// 			}
// 			if n > 0 && n <= len(buf) {
// 				cs.Write(buf[:n])
// 			}
// 		}
// 	}
// }

// func WriteChannelG(ctx context.Context, w io.WriteCloser, wg sync.WaitGroup) {

// 	for {
// 		select {
// 		case v, ok := <-ClientRequestChan:
// 			if !ok {
// 				glog.Errorf("ClientRequestChan close!!")
// 				return
// 			}
// 			_, err := w.Write(v)
// 			if err != nil {
// 				glog.Errorf("WriteChannelG write error %v", err)
// 			}
// 		case <-ctx.Done():
// 			glog.Info("WriteChannelG exit")
// 			return
// 		}
// 	}
// }
