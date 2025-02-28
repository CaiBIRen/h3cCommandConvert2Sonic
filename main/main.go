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

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sonic-unis-framework/configuration"
	"sonic-unis-framework/device"
	"sonic-unis-framework/httpclient"
	"sonic-unis-framework/netconf/server"
	"sonic-unis-framework/sshserver"
	"time"

	"github.com/coreos/pkg/capnslog"
	gliderssh "github.com/gliderlabs/ssh"
	rotatelogs "github.com/iproj/file-rotatelogs"
)

// Command line parameters
var (
	logpath        string
	port           int // Server port
	privateKeyPath = "/etc/sonic/netconf-key"
)

func init() {
	// Parse command line
	flag.StringVar(&logpath, "logpath", "/var/log/sonic-unis-framework", "log dir")
	loggerSet(logpath)
	// flag.StringVar(&clientAuth, "client_auth", "none", "Client auth mode - none|user")
	flag.Parse()
	// Suppress warning messages related to logging before flag parse
	flag.CommandLine.Parse([]string{})
	ConfigfileInit()
}

func main() {
	configuration.ConfigViper()
	device.NewDevice()
	httpclient.NewClient()
	//下配置的时候再开放redis功能
	//redisclient.NewClient()
	// MakeSSHKeyPair(publicKeyPath, privateKeyPath)
	// startip := basic.FindAManagementIP()
	// sshsrvaddr := fmt.Sprintf(startip,)
	sshsrv := &gliderssh.Server{Addr: ":22", Handler: sshserver.SessionHandler}
	srv1 := &gliderssh.Server{Addr: ":830", Handler: server.DefaultHandler}
	srv1.SubsystemHandlers = map[string]gliderssh.SubsystemHandler{}

	sshsrv.SetOption(gliderssh.HostKeyFile(privateKeyPath))
	sshsrv.SetOption(gliderssh.NoPty())
	sshsrv.SetOption(gliderssh.PasswordAuth(authenticate))

	srv1.SetOption(gliderssh.HostKeyFile(privateKeyPath))
	srv1.SetOption(gliderssh.NoPty())
	srv1.SetOption(gliderssh.PasswordAuth(authenticate))

	srv1.SubsystemHandlers["netconf"] = server.SessionHandler
	go sshsrv.ListenAndServe()
	if err := srv1.ListenAndServe(); err != nil {
		fmt.Printf("main error happened %v", err)
	}
}

func authenticate(ctx gliderssh.Context, password string) bool {

	// pamAuthenticator := server.NewPAMAuthenticator(ctx.User(), password)

	// if !pamAuthenticator.Authenticate() {
	// 	glog.Errorf("[PAM] Authentication failed user:(%s)", ctx.User())
	// 	return false
	// }

	// ctx.SetValue("auth-type", "local")

	// ctx.SetValue("auth", pamAuthenticator)

	// ctx.SetValue("uuid", uuid.New().String())

	// glog.Infof("Authentication success user:(%s)", ctx.User())

	return true
}

func loggerSet(dirpath string) {
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		os.MkdirAll(dirpath, os.ModePerm)
	}
	writer := GetWriter(dirpath + "/sonic-agent.log")
	capnslog.SetFormatter(capnslog.NewPrettyFormatter(io.MultiWriter(writer), false))
}

func GetWriter(filename string) io.Writer {
	writer, _ := rotatelogs.New(
		filename+".%Y-%m-%d",
		rotatelogs.WithLinkName(filename),         // 生成软链，指向最新日志文
		rotatelogs.WithMaxAge(90*24*time.Hour),    // 文件最大保存时间
		rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
	)
	return writer
}

func ConfigfileInit() {
	err := os.MkdirAll("/etc/sonic-unis-framework", 0644)
	if err != nil {
		fmt.Printf("make dir failed %v", err)
		panic("make config dir failed")
	}
	_, err = os.Stat("/etc/sonic-unis-framework/config.json")
	if err == nil {
		return
	} else if os.IsNotExist(err) {
		fd, err := os.OpenFile("/etc/sonic-unis-framework/config.json", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("make config file failed %v", err)
			panic("make config file failed")
		}
		fd.Write([]byte("{\"Service\":\"sonic-unis-framework\",\"Company\":\"H3C\"}"))
	} else {
		fmt.Printf("unexpected error %v", err)
		panic("unexpected error")
	}
}
