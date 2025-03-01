package httpclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sonic-unis-framework/basic"
	sonicmodel "sonic-unis-framework/model/sonic"

	"github.com/coreos/pkg/capnslog"
)

var glog = capnslog.NewPackageLogger("sonic-unis-framework", "HTTPCLIENT")

type SonicHttpClient struct {
	C         *http.Client
	Urlprefix string
	Username  string
	Password  string
}

type SError struct {
	ErrorType    string `json:"error-type"`
	ErrorTag     string `json:"error-tag"`
	ErrorAppTag  string `json:"error-app-tag"`
	ErrorMessage string `json:"error-message"`
}

// 定义包含错误列表的结构体
type SErrors struct {
	ErrorList []SError `json:"error"`
}

// 定义顶层的结构体
type IetfRestconfErrors struct {
	SErrors SErrors `json:"ietf-restconf:errors"`
}

type SonicResp struct {
	Code         int
	ErrorMessage IetfRestconfErrors
	Responese    interface{}
}

var SONICCLENT *SonicHttpClient

func NewClient() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}
	// basicurl := "https://10.246.144.12:8443"
	basicurl := "https://localhost:443"

	SONICCLENT = &SonicHttpClient{
		C:         client,
		Urlprefix: basicurl,
		Username:  basic.DefaultUser,
		Password:  basic.DefaultPassword,
	}
}

func (shc *SonicHttpClient) SendSonicRequest(Operation string, Urlsuffix string, bf *bytes.Buffer) *SonicResp {
	url := shc.Urlprefix + Urlsuffix
	Method := OperationToMethod(Operation)
	if Method == "" {
		sresp := EncapsolateErrorStruct(basic.DefaultHttpErrorCode, "Unsupported operation type")
		return sresp
	}
	var req *http.Request
	if bf != nil {
		req, _ = http.NewRequest(Method, url, bf)
	} else {
		req, _ = http.NewRequest(Method, url, nil)
	}
	req.Header.Set("Content-Type", "application/yang-data+json")
	req.SetBasicAuth(shc.Username, shc.Password)
	resp, err := shc.C.Do(req)
	if err != nil {
		glog.Errorf("send http request error:%v", err)
		sresp := EncapsolateErrorStruct(basic.DefaultHttpErrorCode, err.Error())
		return sresp
	}
	if resp.StatusCode > basic.DefaultSuccess {
		glog.Errorf("request sonic server failed")
		var sresp SonicResp
		sresp.Code = resp.StatusCode
		body, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &(sresp.ErrorMessage))
		return &sresp
	}
	//操作成功
	var sresp SonicResp
	sresp.Code = resp.StatusCode
	if resp.ContentLength == 0 {
		sresp.Responese = nil
	} else {
		body, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &(sresp.Responese))
	}

	return &sresp
}

func EncapsolateErrorStruct(code int, errormsg string) *SonicResp {
	resp := &SonicResp{
		Code: code,
		ErrorMessage: IetfRestconfErrors{
			SErrors: SErrors{
				ErrorList: []SError{
					{ErrorMessage: errormsg},
				},
			},
		},
	}
	return resp

}

func OperationToMethod(operation string) string {
	switch operation {
	case basic.OPERMERGE:
		return "PATCH"
	case basic.OPERREMOVE:
		return "DELETE"
	case basic.OPERGET:
		return "GET"
	case basic.OPERACTION:
		return "PATCH"
	}
	return ""
}

func (shc *SonicHttpClient) GetVlanFromSONIC(mapname string) string {
	queryurl := fmt.Sprintf("/restconf/data/sonic-vxlan:sonic-vxlan/VXLAN_TUNNEL_MAP/VXLAN_TUNNEL_MAP_LIST=%s,%s/vlan", basic.TUNNELNAME, mapname)
	url := shc.Urlprefix + queryurl
	req, err := http.NewRequest("Get", url, nil)
	if err != nil {
		glog.Errorf("Error creating request:%v", err)
		return ""
	}
	req.SetBasicAuth(shc.Username, shc.Password)
	resp, err := shc.C.Do(req)
	if err != nil {
		glog.Errorf("GetVlanFromSONIC error:%v", err)
		return ""
	}
	if resp.StatusCode > basic.DefaultSuccess {
		glog.Errorf("GetVlanFromSONIC back error code %d", resp.StatusCode)
		return ""
	}
	//操作成功
	if resp.ContentLength == 0 {
		glog.Errorf("GetVlanFromSONIC back ContentLength 0")
		return ""
	} else {
		var vxlanvlan sonicmodel.VxlanVlan
		body, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &vxlanvlan)
		return vxlanvlan.Vlan
	}
}
