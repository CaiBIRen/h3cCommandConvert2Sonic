package sonichandlers

import (
	"errors"
	"fmt"
	"sonic-unis-framework/basic"
	"sonic-unis-framework/httpclient"
	sonicmodel "sonic-unis-framework/model/sonic"
	"sonic-unis-framework/tcontext"

	"github.com/mitchellh/mapstructure"
	"github.com/vishvananda/netlink"
)

func GetSONICDevice(t *tcontext.Tcontext) error {
	var devicenode sonicmodel.Device
	devicenode.Base.HostName = "SONIC"
	devicenode.Base.HostDescription = "Peace and love"
	devicenode.PhysicalEntities.Entity.SoftwareRev = "SONIC VS Software System"
	t.SonicConfig[basic.SONICDEVICE] = devicenode
	return nil
}

func GetSONICLLDP(t *tcontext.Tcontext) error {
	var lldproot sonicmodel.OpenConfigLLDP
	rsp := httpclient.SONICCLENT.SendSonicRequest("get", "/restconf/data/openconfig-lldp:lldp/interfaces/interface", nil)
	err := GetHandlerResolve(rsp)
	if err != nil {
		return err
	}
	if rsp.Responese != nil {
		err := mapstructure.Decode(rsp.Responese, &lldproot)
		if err != nil {
			return err
		}
		t.SonicConfig[basic.SONICLLDP] = lldproot
	}
	return nil
}

func GetSONICPort(t *tcontext.Tcontext) error {
	var ports sonicmodel.PortTable
	rsp := httpclient.SONICCLENT.SendSonicRequest("get", "/restconf/data/sonic-port:sonic-port/PORT_TABLE/PORT_TABLE_LIST", nil)
	err := GetHandlerResolve(rsp)
	if err != nil {
		return err
	}
	if rsp.Responese != nil {
		err := mapstructure.Decode(rsp.Responese, &ports)

		if err != nil {
			return err
		}

		for k, v := range ports.PortTableList {
			link, err := netlink.LinkByName(v.Ifname)
			if err != nil {
				glog.Errorf("[GetSONICPort] %s netlink error %v", v.Ifname, err)
				continue
				//return err
			}
			ports.PortTableList[k].MAC = link.Attrs().HardwareAddr.String()
		}
		t.SonicConfig[basic.SONICPORT] = ports
		//fmt.Println(t.SonicConfig[basic.SONICINTERFACE])
	}
	return nil
}

func GetSONICPortChannelList(t *tcontext.Tcontext) error {
	var portchannelroot sonicmodel.PortChannelList
	rsp := httpclient.SONICCLENT.SendSonicRequest("get", "/restconf/data/sonic-portchannel:sonic-portchannel/LAG_TABLE/LAG_TABLE_LIST", nil)
	err := GetHandlerResolve(rsp)
	if err != nil {
		glog.Errorf("[sonichandlers] portchannellist err %v", err)
		return err
	}
	if rsp.Responese != nil {
		err := mapstructure.Decode(rsp.Responese, &portchannelroot)
		if err != nil {
			return err
		}

		for k, v := range portchannelroot.LAGTableList {
			link, err := netlink.LinkByName(v.Name)
			if err != nil {
				glog.Errorf("[GetSONICPort] %s netlink error %v", v.Name, err)
				continue
				//return err
			}
			portchannelroot.LAGTableList[k].MAC = link.Attrs().HardwareAddr.String()
		}
		t.SonicConfig[basic.SONICPORTCHANNEL] = portchannelroot
	}
	return nil
}

func GetSONICPortChannelMembers(t *tcontext.Tcontext) error {
	var portmembersroot sonicmodel.PortChannelMembers
	rsp := httpclient.SONICCLENT.SendSonicRequest("get", "/restconf/data/sonic-portchannel:sonic-portchannel/PORTCHANNEL_MEMBER/PORTCHANNEL_MEMBER_LIST", nil)
	err := GetHandlerResolve(rsp)
	if err != nil {
		glog.Errorf("[sonichandlers] PortChannelMembers err %v", err)
		return err
	}
	if rsp.Responese != nil {
		err := mapstructure.Decode(rsp.Responese, &portmembersroot)
		if err != nil {
			return err
		}
		t.SonicConfig[basic.SONICPORTCHANNELMEMBERS] = portmembersroot
	}
	return nil
}

func GetSONICBridgeMAC(t *tcontext.Tcontext) error {
	var metadataroot sonicmodel.SonicDeviceMetadata
	link, err := netlink.LinkByName("eth0")
	if err != nil {
		glog.Errorf("[GetSONICBridgeMAC] %s netlink error %v", "eth0", err)
		return err
	}
	metadataroot.MAC = link.Attrs().HardwareAddr.String()
	t.SonicConfig[basic.SONICSYSTEMID] = metadataroot
	return nil
}

func GetSONICBGPInstance(t *tcontext.Tcontext) error {
	var asnroot sonicmodel.BGPGlobalConfigASN
	rsp := httpclient.SONICCLENT.SendSonicRequest("get", "/restconf/data/sonic-bgp-global:sonic-bgp-global/BGP_GLOBALS/BGP_GLOBALS_LIST=default/local_asn", nil)
	err := GetHandlerResolve(rsp)
	if err != nil {
		return err
	}
	if rsp.Responese != nil {
		err := mapstructure.Decode(rsp.Responese, &asnroot)
		if err != nil {
			return err
		}
		t.SonicConfig[basic.SONICBGPKEY] = asnroot
	}
	return nil
}

func GetSONICVlanInterface(t *tcontext.Tcontext) error {
	var vlaninterfaces sonicmodel.Get_VLANInterface
	rsp := httpclient.SONICCLENT.SendSonicRequest("get", "/restconf/data/sonic-vlan-interface:sonic-vlan-interface/VLAN_INTERFACE", nil)
	err := GetHandlerResolve(rsp)
	if err != nil {
		return err
	}
	if rsp.Responese != nil {
		err := mapstructure.Decode(rsp.Responese, &vlaninterfaces)
		if err != nil {
			return err
		}
		t.SonicConfig[basic.SONICVLANINTERFACEKEY] = vlaninterfaces
	}
	return nil
}

func GetSONICVlanInterfaceIPByName(VLANName string) (string, error) {
	var vlaninterfaceips sonicmodel.Get_VLANInterfaceListIPs
	urlsuffix := "/restconf/data/sonic-vlan-interface:sonic-vlan-interface/VLAN_INTERFACE/VLAN_INTERFACE_IPADDR_LIST"
	rsp := httpclient.SONICCLENT.SendSonicRequest("get", urlsuffix, nil)
	err := GetHandlerResolve(rsp)
	if err != nil {
		return "", err
	}
	if rsp.Responese != nil {
		err := mapstructure.Decode(rsp.Responese, &vlaninterfaceips)
		if err != nil {
			return "", err
		}
		for _, v := range vlaninterfaceips.VLAN_INTERFACE_LIST_IP {
			if v.VlanName == VLANName {
				return v.IPPrefix, nil
			}
		}
	}
	return "", errors.New(basic.RESOURCENOTFOUND)
}

func GetOSPFInstancesByDescription(description string) (string, error) {
	var ospfrouters sonicmodel.Get_SonicOspfv2Router
	urlsuffix := "/restconf/data/sonic-ospfv2:sonic-ospfv2/OSPFV2_ROUTER"
	rsp := httpclient.SONICCLENT.SendSonicRequest("get", urlsuffix, nil)
	err := GetHandlerResolve(rsp)
	if err != nil {
		return "", err
	}
	if rsp.Responese != nil {
		err := mapstructure.Decode(rsp.Responese, &ospfrouters)
		if err != nil {
			return "", err
		}
		for _, v := range ospfrouters.OSPFv2Router.OSPFv2RouterList {
			if v.Description == description {
				return v.VrfName, nil
			}
		}
	}
	return "", errors.New(basic.RESOURCENOTFOUND)
}

func GetSONICVlanInterfaceByName(Name string) error {
	urlsuffix := fmt.Sprintf("/restconf/data/sonic-vlan-interface:sonic-vlan-interface/VLAN_INTERFACE/VLAN_INTERFACE_LIST=%s", Name)
	rsp := httpclient.SONICCLENT.SendSonicRequest("get", urlsuffix, nil)
	err := GetHandlerResolve(rsp)
	if err != nil {
		return err
	}
	return nil
}

func GetSONICLoopbackInterfaceByName(Name string) error {
	urlsuffix := fmt.Sprintf("/restconf/data/sonic-loopback-interface:sonic-loopback-interface/LOOPBACK_INTERFACE/LOOPBACK_INTERFACE_LIST=%s", Name)
	rsp := httpclient.SONICCLENT.SendSonicRequest("get", urlsuffix, nil)
	err := GetHandlerResolve(rsp)
	if err != nil {
		return err
	}
	return nil
}

func GetSONICRoutepolicySetPrefixList(t *tcontext.Tcontext) error {
	// var asnroot sonicmodel.BGPGlobalConfigASN
	// rsp := httpclient.SONICCLENT.SendSonicRequest("get", "/restconf/data/sonic-routing-policy-sets:sonic-routing-policy-sets/PREFIX/PREFIX_LIST", nil)
	// err := GetHandlerResolve(rsp)
	// if err != nil {
	// 	return err
	// }
	// if rsp.Responese != nil {
	// 	err := mapstructure.Decode(rsp.Responese, &asnroot)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fmt.Println("++", asnroot)
	// 	t.SonicConfig[basic.SONICBGPKEY] = asnroot
	// }
	return nil
}

func GetSONICVRFByName(Name string) error {
	urlsuffix := fmt.Sprintf("/restconf/data/sonic-vrf:sonic-vrf/VRF/VRF_LIST=%s", Name)
	rsp := httpclient.SONICCLENT.SendSonicRequest("get", urlsuffix, nil)
	err := GetHandlerResolve(rsp)
	if err != nil {
		return err
	}
	return nil
}

func GetHandlerResolve(rsp *httpclient.SonicResp) error {
	code := rsp.Code
	if code > basic.DefaultSuccess {
		if len(rsp.ErrorMessage.SErrors.ErrorList) <= 0 {
			return errors.New("opreation failed,sonic resp body nil,unknown error")
		} else {
			rsperr := rsp.ErrorMessage.SErrors.ErrorList[0].ErrorMessage
			if rsperr == "" {
				rsperr = rsp.ErrorMessage.SErrors.ErrorList[0].ErrorTag
			}
			errmsg := fmt.Sprintf("GET failed:%s", rsperr)
			glog.Error(errmsg)
			return errors.New(errmsg)
		}
	} else {
		return nil
	}
}
