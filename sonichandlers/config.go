package sonichandlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sonic-unis-framework/basic"
	"sonic-unis-framework/httpclient"
	"sonic-unis-framework/model"
	sonicmodel "sonic-unis-framework/model/sonic"
	"sonic-unis-framework/redisclient"
	"sonic-unis-framework/tcontext"
	"strconv"
	"strings"

	"github.com/coreos/pkg/capnslog"
	"github.com/mitchellh/mapstructure"
	"github.com/vishvananda/netlink"
)

var glog = capnslog.NewPackageLogger("sonic-unis-framework", "SONIC_CONFIG")

//按照sonic的配置顺序注册的handlers
/*
{ vxlantunnel  evpn_nvo bgp } 基础配置

| SONiC                | H3C                |
|----------------------|--------------------|
| Acl                  | acl                |
| Vlan                 | vsi                |
| map vlan vni         | vsi                |
| VRF                  | vrf l3vpn          |
| Vlan-interface       | vsi-interface      |
| vrf_vni_map          | l3vsi-interface    |
| frr-BGP              | vrf l3vpn          |
| route                | StaticRoute        |

直接取context中的报文调用接口即可,翻译工作不放在这里
*/

var MERGE_FEATURE_ORDER_LIST = []string{
	basic.SONICVLANKEY,
	basic.SONICVXLANKEY,
	basic.SONICVRFKEY,
	basic.SONICVLANINTERFACEKEY,
	basic.SONICVLANINTERFACEIPADDRKEY,
	basic.SONICLOOPBACKKEY,
	basic.SONICLOOPBACKINTERFACEIPADDRKEY,
	basic.SONICROUTEMAPSETKEY,
	basic.SONICROUTEMAPKEY,
	basic.SONICBGPKEY,
	basic.SONICOSPFKEY,
	basic.SONICROUTECOMMONKEY,
	basic.SONICSTATICROUTEKEY,
	basic.SONICINTERFACEMAC,
}

var REMOVE_FEATURE_ORDER_LIST = []string{
	basic.SONICROUTECOMMONKEY,
	basic.SONICROUTEMAPKEY,
	basic.SONICROUTEMAPSETKEY,
	basic.SONICSTATICROUTEKEY,
	basic.SONICVLANINTERFACEIPADDRKEY,
	basic.SONICVLANINTERFACEKEY,
	basic.SONICLOOPBACKINTERFACEIPADDRKEY,
	basic.SONICLOOPBACKKEY,
	basic.SONICOSPFKEY,
	basic.SONICBGPKEY,
	basic.SONICVXLANKEY,
	basic.SONICVLANKEY,
	basic.SONICVRFKEY,
}

type Handler func(t *tcontext.Tcontext) error

type chain_node map[string]Handler

type config_chain struct {
	merge_chain  chain_node
	remove_chain chain_node
}

// one feature one handler
func (s config_chain) SONICChainRegister(opreation string, feature string, f func(*tcontext.Tcontext) error) {
	glog.Infof("[chain register]operation %v, feature %v", opreation, feature)
	if opreation == basic.OPERMERGE {
		s.merge_chain[feature] = f
	} else {
		s.remove_chain[feature] = f
	}
}

var Config_chain config_chain = config_chain{
	merge_chain:  make(chain_node),
	remove_chain: make(chain_node),
}

func init() {
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICVLANKEY, ConfigSONICVlan)
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICVXLANKEY, ConfigSONICVxlan)
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICVRFKEY, ConfigSONICVrf)
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICVLANINTERFACEKEY, ConfigSONICVlanInterface)
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICLOOPBACKKEY, ConfigSONICLoopbackInterface)
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICVLANINTERFACEIPADDRKEY, ConfigSONICVlanInterfaceIPAddr)
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICLOOPBACKINTERFACEIPADDRKEY, ConfigSONICLoopbackInterfaceIPAddr)
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICOSPFKEY, ConfigSONICOSPFv2)
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICBGPKEY, ConfigSONICBGP)
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICROUTECOMMONKEY, ConfigSONICRouteCommon)
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICSTATICROUTEKEY, ConfigSONICStaticRoute)
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICROUTEMAPSETKEY, ConfigSONICRoutemapSet)
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICROUTEMAPKEY, ConfigSONICRoutemap)
	Config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICINTERFACEMAC, ConfigSONICInterfaceMacInKernel)
	//sonic_config_chain.SONICChainRegister(basic.OPERMERGE, basic.SONICINDEX, SetIndexOfResouce)

	//删除IP部分待实现
	Config_chain.SONICChainRegister(basic.OPERREMOVE, basic.SONICVLANKEY, RemoveSONICVlan)
	Config_chain.SONICChainRegister(basic.OPERREMOVE, basic.SONICVXLANKEY, RemoveSONICVxlan)
	Config_chain.SONICChainRegister(basic.OPERREMOVE, basic.SONICVRFKEY, RemoveSONICVrf)
	Config_chain.SONICChainRegister(basic.OPERREMOVE, basic.SONICVLANINTERFACEKEY, RemoveSONICVlanInterface)
	Config_chain.SONICChainRegister(basic.OPERREMOVE, basic.SONICLOOPBACKKEY, RemoveSONICLoopbackInterface)
	Config_chain.SONICChainRegister(basic.OPERREMOVE, basic.SONICVLANINTERFACEIPADDRKEY, RemoveSONICVlanInterfaceIPAddr)
	Config_chain.SONICChainRegister(basic.OPERREMOVE, basic.SONICOSPFKEY, RemoveSONICOSPFv2)
	Config_chain.SONICChainRegister(basic.OPERREMOVE, basic.SONICBGPKEY, RemoveSONICBGP)
	Config_chain.SONICChainRegister(basic.OPERREMOVE, basic.SONICROUTECOMMONKEY, RemoveSONICRouteCommon)
	Config_chain.SONICChainRegister(basic.OPERREMOVE, basic.SONICSTATICROUTEKEY, RemoveSONICStaticRoute)
	Config_chain.SONICChainRegister(basic.OPERREMOVE, basic.SONICROUTEMAPSETKEY, RemoveSONICRoutemapSet)
	Config_chain.SONICChainRegister(basic.OPERREMOVE, basic.SONICROUTEMAPKEY, RemoveSONICRoutemap)
}

func SonicAddConfigHandlers(t *tcontext.Tcontext) (bool, error) {
	for _, k := range MERGE_FEATURE_ORDER_LIST {
		if _, ok := t.SonicConfig[k]; ok {
			err := (Config_chain.merge_chain)[k](t)
			if err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func SonicRemoveConfigHandlers(t *tcontext.Tcontext) (bool, error) {
	for _, k := range REMOVE_FEATURE_ORDER_LIST {
		if _, ok := t.SonicConfig[k]; ok {
			err := (Config_chain.remove_chain)[k](t)
			if err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

// 暂不设置索引
// func SetIndexOfResouce(t *tcontext.Tcontext) error {
// 	data := t.SonicConfig[basic.SONICINDEX].(map[string]int)
// 	for k, v := range data {
// 		redisclient.IndexSet(k, v)
// 	}
// 	return nil
// }

func ConfigSONICInterfaceMacInKernel(t *tcontext.Tcontext) error {
	macdata := t.SonicConfig[basic.SONICINTERFACEMAC].(model.Mac_interface_list)
	for _, v := range macdata.Mac_interfaces {
		link, err := netlink.LinkByName(v.Ifname)
		if err != nil {
			return err
		}
		mac, err := net.ParseMAC(v.Mac)
		if err != nil {
			return err
		}
		err = netlink.LinkSetHardwareAddr(link, mac)
		if err != nil {
			glog.Errorf("netlink set interface %s mac failed,err:%s", v.Ifname, err.Error())
			return err
		}
	}
	return nil
}

func ConfigSONICVxlan(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-vxlan:sonic-vxlan/VXLAN_TUNNEL_MAP"
	vxlandata := t.SonicConfig[basic.SONICVXLANKEY].(sonicmodel.Vxlanroot)

	sonicvxlan, err := json.Marshal(vxlandata)
	if err != nil {
		glog.Errorf("vxlan root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicvxlan)
	glog.Info("vxlan config is sending")
	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, b)
	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Info("vxlan config has completed")
	//SaveIndexToRedis(t, "VLANMapping")
	return nil
}

func RemoveSONICVxlan(t *tcontext.Tcontext) error {
	vxlandata := t.SonicConfig[basic.SONICVXLANKEY].(sonicmodel.Vxlanroot)
	for _, v := range vxlandata.SonicVxlan.VXLAN_TUNNEL_MAP_LIST {
		urlsuffix := fmt.Sprintf("/restconf/data/sonic-vxlan:sonic-vxlan/VXLAN_TUNNEL/VXLAN_TUNNEL_LIST=%s", v.Mapname)
		glog.Infof("vxlan mapping {%s} is deleting", v.Mapname)
		rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
		err := DeleteHandlerResolve(rsp)
		if err != nil {
			return err
		}
		glog.Infof("vxlan mapping {%s} has deleted", v.Mapname)
	}
	return nil
}

func ConfigSONICVlan(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-vlan:sonic-vlan"
	vlandata := t.SonicConfig[basic.SONICVLANKEY].(sonicmodel.Vlanroot)
	sonicvlan, err := json.Marshal(vlandata)
	if err != nil {
		glog.Errorf("vlan root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicvlan)
	glog.Info("vlan config is sending")
	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, b)
	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Info("vlan config send completed")
	return nil
}

func RemoveSONICVlan(t *tcontext.Tcontext) error {
	vlandata := t.SonicConfig[basic.SONICVLANKEY].(sonicmodel.Vlanroot)
	for _, v := range vlandata.SonicVLAN.VLAN.VLANList {
		urlsuffix := fmt.Sprintf("/restconf/data/openconfig-interfaces:interfaces/interface=%s", v.Name)
		glog.Infof("vlan all config {%s} is deleting", v.Name)
		rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
		err := DeleteHandlerResolve(rsp)
		if err != nil {
			return err
		}
		glog.Infof("vlan all config {%s} has deleted", v.Name)
	}
	return nil
}

func ConfigSONICVlanInterface(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-vlan-interface:sonic-vlan-interface"
	vlaninterfacedata := t.SonicConfig[basic.SONICVLANINTERFACEKEY].(sonicmodel.VlanInterfaceroot)
	sonicvlaninterface, err := json.Marshal(vlaninterfacedata)
	if err != nil {
		glog.Errorf("vlan root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicvlaninterface)
	glog.Info("vlan interface config is sending")
	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, b)

	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Info("vlan interface config completed")
	return nil
}

func ConfigSONICLoopbackInterface(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-loopback-interface:sonic-loopback-interface/LOOPBACK_INTERFACE/LOOPBACK_INTERFACE_LIST"
	loopbackinterfacedata := t.SonicConfig[basic.SONICLOOPBACKKEY].(sonicmodel.LoopbackInterfacesroot)
	sonicloopbackinterface, err := json.Marshal(loopbackinterfacedata)
	if err != nil {
		glog.Errorf("loopback root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicloopbackinterface)
	glog.Infof("loopback interface config is sending,data: %s", b.String())
	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, b)

	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Info("loopback interface config completed")
	return nil
}

func RemoveSONICLoopbackInterface(t *tcontext.Tcontext) error {
	loopbackinterfacedata := t.SonicConfig[basic.SONICLOOPBACKKEY].(sonicmodel.LoopbackInterfacesroot)
	for _, v := range loopbackinterfacedata.LoopbackInterfaceList {
		urlsuffix := fmt.Sprintf("/restconf/data/openconfig-interfaces:interfaces/interface=%s", v.LoIfName)
		glog.Infof("loopback interface %s is deleting", v.LoIfName)
		rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
		err := DeleteHandlerResolve(rsp)
		if err != nil {
			return err
		}
		glog.Infof("loopback interface %s has deleted", v.LoIfName)
	}
	return nil
}

func RemoveSONICVlanInterface(t *tcontext.Tcontext) error {
	vlaninterfacedata := t.SonicConfig[basic.SONICVLANINTERFACEKEY].(sonicmodel.VlanInterfaceroot)
	for _, v := range vlaninterfacedata.SonicVLANInterface.VLAN_INTERFACE.VLAN_INTERFACE_LIST {
		urlsuffix := fmt.Sprintf("/restconf/data/openconfig-interfaces:interfaces/interface=%s", v.VlanName)
		glog.Infof("vlan interface %s is deleting", v.VlanName)
		rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
		err := DeleteHandlerResolve(rsp)
		if err != nil {
			return err
		}
		glog.Infof("vlan interface %s has deleted", v.VlanName)
	}
	return nil
}

func ConfigSONICVlanInterfaceIPAddr(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-vlan-interface:sonic-vlan-interface/VLAN_INTERFACE/VLAN_INTERFACE_IPADDR_LIST"
	vlaninterfaceipdata := t.SonicConfig[basic.SONICVLANINTERFACEIPADDRKEY].(sonicmodel.VLANInterfaceIPAddrList)
	sonicvlaninterface, err := json.Marshal(vlaninterfaceipdata)
	if err != nil {
		glog.Errorf("vlan root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicvlaninterface)
	glog.Infof("vlan interface ip is sending,data %s", b.String())
	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, b)
	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Info("vlan interface ip has completed")
	return nil
}

func RemoveSONICVlanInterfaceIPAddr(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-vlan-interface:sonic-vlan-interface/VLAN_INTERFACE/VLAN_INTERFACE_IPADDR_LIST"
	vlaninterfaceipdata := t.SonicConfig[basic.SONICVLANINTERFACEIPADDRKEY].(sonicmodel.VlanInterfaceroot)
	for _, v := range vlaninterfaceipdata.SonicVLANInterface.VLAN_INTERFACE.VLAN_INTERFACE_LIST {
		urlsuffix = fmt.Sprintf("/restconf/data/openconfig-interfaces:interfaces/interface=%s/openconfig-vlan:routed-vlan/openconfig-if-ip:ipv4/addresses", v.VlanName)
		glog.Infof("vlan interface addr {%s} is deleting", v.VlanName)
		rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
		err := DeleteHandlerResolve(rsp)
		if err != nil {
			return err
		}
		glog.Infof("vlan interface addr {%s} has deleted", v.VlanName)
	}
	return nil
}

func ConfigSONICLoopbackInterfaceIPAddr(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-loopback-interface:sonic-loopback-interface/LOOPBACK_INTERFACE/LOOPBACK_INTERFACE_IPADDR_LIST"
	loopbackinterfaceipdata := t.SonicConfig[basic.SONICLOOPBACKINTERFACEIPADDRKEY].(sonicmodel.LoopbackInterfacesIPAddrList)
	sonicloopbackinterfaceips, err := json.Marshal(loopbackinterfaceipdata)
	if err != nil {
		glog.Errorf("LoopbackInterfacesIPAddrLists root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicloopbackinterfaceips)
	glog.Info("loopback interface ip is sending")
	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, b)
	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Info("loopback interface ip has completed")
	return nil
}

func RemoveSONICLoopbackInterfaceIPAddr(t *tcontext.Tcontext) error {
	// urlsuffix := "/restconf/data/sonic-vlan-interface:sonic-vlan-interface/VLAN_INTERFACE/VLAN_INTERFACE_IPADDR_LIST"
	// vlaninterfaceipdata := t.SonicConfig[basic.SONICVLANINTERFACEIPADDRKEY].(sonicmodel.VlanInterfaceroot)
	// for _, v := range vlaninterfaceipdata.SonicVLANInterface.VLAN_INTERFACE.VLAN_INTERFACE_LIST {
	// 	urlsuffix = fmt.Sprintf("/restconf/data/openconfig-interfaces:interfaces/interface=%s/openconfig-vlan:routed-vlan/openconfig-if-ip:ipv4/addresses", v.VlanName)
	// 	glog.Infof("vlan interface addr {%s} is deleting", v.VlanName)
	// 	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
	// 	err := DeleteHandlerResolve(rsp)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	glog.Infof("vlan interface addr {%s} has deleted", v.VlanName)
	// }
	return nil
}

func ConfigSONICVrf(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-vrf:sonic-vrf"
	vrfdata := t.SonicConfig[basic.SONICVRFKEY].(sonicmodel.Vrfroot)
	sonicvrf, err := json.Marshal(vrfdata)
	if err != nil {
		glog.Errorf("vrf root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicvrf)
	glog.Infof("vrf config data is sending")
	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, b)
	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Infof("vrf config data completed")
	//TODO:
	//SaveIndexToRedis(t, "VRF")
	return nil
}

func RemoveSONICVrf(t *tcontext.Tcontext) error {
	vrfdata := t.SonicConfig[basic.SONICVRFKEY].(sonicmodel.Vrfroot)
	for _, v := range vrfdata.SonicVrf.VRF.VRF_LIST {
		urlsuffix := fmt.Sprintf("/restconf/data/sonic-vrf:sonic-vrf/VRF/VRF_LIST=%s", v.VrfName)
		glog.Infof("vrf {%s} is deleting", v.VrfName)
		rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
		err := DeleteHandlerResolve(rsp)
		if err != nil {
			return err
		}
		glog.Infof("vrf {%s} has deleted", v.VrfName)
	}
	return nil
}

func ConfigSONICBGP(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-bgp-global:sonic-bgp-global"
	bgpdata := t.SonicConfig[basic.SONICBGPKEY].(sonicmodel.BGProot)
	sonicbgp, err := json.Marshal(bgpdata)
	if err != nil {
		glog.Errorf("bgp root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicbgp)
	glog.Info("bgp config data sending")
	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, b)
	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Info("bgp config data has completed")
	return nil
}

func RemoveSONICBGP(t *tcontext.Tcontext) error {
	bgprootdata := t.SonicConfig[basic.SONICBGPKEY].(sonicmodel.BGProot)
	for _, v := range bgprootdata.Sonicbgpglobal.BGP_GLOBALS.BGP_GLOBALS_LIST {
		urlsuffix := fmt.Sprintf("/restconf/data/sonic-bgp-global:sonic-bgp-global/BGP_GLOBALS/BGP_GLOBALS_LIST=%s", v.VrfName)
		rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
		err := DeleteHandlerResolve(rsp)
		if err != nil {
			return err
		}
	}
	for _, v := range bgprootdata.Sonicbgpglobal.BGP_GLOBALS_AF.BGP_GLOBALS_AF_LIST {
		if len(v.ImportRTS) > 0 || len(v.ExportRTS) > 0 {
			for _, vv := range v.ImportRTS {
				urlsuffix := fmt.Sprintf("/restconf/data/sonic-bgp-global:sonic-bgp-global/BGP_GLOBALS_AF/BGP_GLOBALS_AF_LIST=%s,%s/import-rts=%s", v.VrfName, v.AFISAFI, vv)
				rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
				err := DeleteHandlerResolve(rsp)
				if err != nil {
					return err
				}

			}
			for _, vv := range v.ExportRTS {
				urlsuffix := fmt.Sprintf("/restconf/data/sonic-bgp-global:sonic-bgp-global/BGP_GLOBALS_AF/BGP_GLOBALS_AF_LIST=%s,%s/export-rts=%s", v.VrfName, v.AFISAFI, vv)
				rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
				err := DeleteHandlerResolve(rsp)
				if err != nil {
					return err
				}

			}
			return nil
		}
		//当子元素都为空时删除地址簇
		urlsuffix := fmt.Sprintf("/restconf/data/sonic-bgp-global:sonic-bgp-global/BGP_GLOBALS_AF/BGP_GLOBALS_AF_LIST=%s,%s", v.VrfName, v.AFISAFI)
		rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
		err := DeleteHandlerResolve(rsp)
		if err != nil {
			return err
		}
	}
	return nil
}

func ConfigSONICRouteCommon(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-route-common:sonic-route-common"
	routecommondata := t.SonicConfig[basic.SONICROUTECOMMONKEY].(sonicmodel.SonicRouteCommonroot)
	sonicroutecommon, err := json.Marshal(routecommondata)
	if err != nil {
		glog.Errorf("bgp root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicroutecommon)
	glog.Info("route common config data sending")
	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, b)
	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Info("route common config has completed")
	return nil
}

func RemoveSONICRouteCommon(t *tcontext.Tcontext) error {
	routecommondata := t.SonicConfig[basic.SONICROUTECOMMONKEY].(sonicmodel.SonicRouteCommonroot)
	for _, v := range routecommondata.SonicRouteCommon.RouteRedistributeList.RouteRedistributes {
		urlsuffix := fmt.Sprintf("/restconf/data/sonic-route-common:sonic-route-common/ROUTE_REDISTRIBUTE/ROUTE_REDISTRIBUTE_LIST=%s,%s,%s,%s",
			v.VrfName, v.SrcProtocol, v.DstProtocol, v.AddrFamily)
		glog.Infof("route common {%s} {%s} {%s} {%s} is deleting", v.VrfName, v.SrcProtocol, v.DstProtocol, v.AddrFamily)
		rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
		err := DeleteHandlerResolve(rsp)
		if err != nil {
			return err
		}
		glog.Infof("route common {%s} {%s} {%s} {%s} has deleted", v.VrfName, v.SrcProtocol, v.DstProtocol, v.AddrFamily)
	}
	return nil
}

func ConfigSONICRoutemapSet(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-routing-policy-sets:sonic-routing-policy-sets"
	routemapsetdata := t.SonicConfig[basic.SONICROUTEMAPSETKEY].(sonicmodel.SonicRoutingPolicySets)
	sonicroutemapset, err := json.Marshal(routemapsetdata)
	if err != nil {
		glog.Errorf("bgp root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicroutemapset)
	glog.Info("route policy set config data sending")
	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, b)
	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Info("route policy set config data completed")
	return nil
}

func RemoveSONICRoutemapSet(t *tcontext.Tcontext) error {
	//先查询Prefixlist
	//因为要获取删除所需要的字段
	var Prefixlist sonicmodel.SonicRoutingPolicyPrefixList
	rsp := httpclient.SONICCLENT.SendSonicRequest("get", "/restconf/data/sonic-routing-policy-sets:sonic-routing-policy-sets/PREFIX/PREFIX_LIST", nil)
	err := GetHandlerResolve(rsp)
	if err != nil {
		glog.Errorf("get prefix list to cachedata error:%s", err)
		return err
	}
	if rsp.Responese != nil {
		err := mapstructure.Decode(rsp.Responese, &Prefixlist)
		if err != nil {
			return err
		}
	} else {
		//no prefixlist return delete ok
		return nil
	}

	existmap1 := make(map[string]sonicmodel.PrefixList)
	existmap2 := make(map[string]int)
	for _, v := range Prefixlist.PrefixLists {
		indexkey := v.SetName + strconv.Itoa(v.SequenceNumber)
		existmap1[indexkey] = v
		existmap2[v.SetName] += 1
	}

	routemapsetdata := t.SonicConfig[basic.SONICROUTEMAPSETKEY].(sonicmodel.SonicRoutingPolicySets)
	if routemapsetdata.SonicRoutingPolicySetsWrapper.Prefix != nil {
		for _, v := range routemapsetdata.SonicRoutingPolicySetsWrapper.Prefix.PrefixList {
			mapkey := v.SetName + strconv.Itoa(v.SequenceNumber)
			if node, ok := existmap1[mapkey]; ok {
				//to delete
				prefixencode := url.QueryEscape(node.IPPrefix)
				urlsuffix := fmt.Sprintf("/restconf/data/sonic-routing-policy-sets:sonic-routing-policy-sets/PREFIX/PREFIX_LIST=%s,%d,%s,%s",
					node.SetName, node.SequenceNumber, prefixencode, node.MasklengthRange)
				glog.Infof("ip prefix list {%s} {%d} is deleting", node.SetName, node.SequenceNumber)
				rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
				err := DeleteHandlerResolve(rsp)
				if err != nil {
					return err
				}
				glog.Infof("ip prefix list {%s} {%d} has deleted", node.SetName, node.SequenceNumber)
				existmap2[v.SetName] -= 1
			} else {
				continue
			}
		}
	}

	//删除prefixset
	for k, v := range existmap2 {
		//prefixset下面list为空
		if v <= 0 {
			urlsuffix := fmt.Sprintf("/restconf/data/sonic-routing-policy-sets:sonic-routing-policy-sets/PREFIX_SET/PREFIX_SET_LIST=%s", k)
			glog.Infof("prefix set {%s} is deleting", k)
			rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
			err := DeleteHandlerResolve(rsp)
			if err != nil {
				return err
			}
			glog.Infof("prefix set {%s} ahs deleted", k)
		}
	}
	return nil
}

func ConfigSONICRoutemap(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-route-map:sonic-route-map"
	routemapdata := t.SonicConfig[basic.SONICROUTEMAPKEY].(sonicmodel.SonicRoutemaproot)
	sonicroutemap, err := json.Marshal(routemapdata)
	if err != nil {
		glog.Errorf("bgp root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicroutemap)
	glog.Infof("route map config data sending,data %s", b.String())
	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, b)
	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Info("route map config data completed")
	return nil
}

func RemoveSONICRoutemap(t *tcontext.Tcontext) error {
	var AllDelete bool
	existmap1 := make(map[string]bool)
	existmap2 := make(map[string]bool)

	routemapdata := t.SonicConfig[basic.SONICROUTEMAPKEY].(sonicmodel.SonicRoutemaproot)
	for _, v := range routemapdata.SonicRouteMap.RouteMap.RouteMapList {
		urlsuffix := fmt.Sprintf("/restconf/data/sonic-route-map:sonic-route-map/ROUTE_MAP/ROUTE_MAP_LIST=%s,%d", v.RouteMapName, v.StmtName)
		glog.Infof("route map {%s} {%d} is deleting", v.RouteMapName, v.StmtName)
		rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
		err := DeleteHandlerResolve(rsp)
		if err != nil {
			return err
		}
		existmap1[v.RouteMapName] = true
		glog.Infof("route map {%s} {%d} has deleted", v.RouteMapName, v.StmtName)
	}

	//如果routemapset下面没有子元素 则可以删除route-policy-Set
	var Routemaplist sonicmodel.SonicRoutingPolicySetList
	urlsuffix := "/restconf/data/sonic-route-map:sonic-route-map/ROUTE_MAP_LIST"
	rsp := httpclient.SONICCLENT.SendSonicRequest("get", urlsuffix, nil)
	err := GetHandlerResolve(rsp)
	if err != nil {
		glog.Errorf("get route map list to cachedata error:%s", err)
		return err
	}
	if rsp.Responese != nil {
		err := mapstructure.Decode(rsp.Responese, &Routemaplist)
		if err != nil {
			return err
		}
	} else {
		//all existmap1 need to  delete
		// return nil
		AllDelete = true
	}

	if AllDelete {
		for k, _ := range existmap1 {
			urlsuffix := fmt.Sprintf("/restconf/data/sonic-route-map:sonic-route-map/ROUTE_MAP_SET/ROUTE_MAP_SET_LIST=%s", k)
			glog.Infof("route policy set {%s} is deleting", k)
			rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
			err := DeleteHandlerResolve(rsp)
			if err != nil {
				return err
			}
			glog.Infof("route policy set {%s} ahs deleted", k)
		}
		return nil
	}

	for _, v := range Routemaplist.Get_RoutePolicyMaps {
		existmap2[v.RouteMapName] = true
	}

	for k, _ := range existmap1 {
		if _, ok := existmap2[k]; !ok {
			urlsuffix := fmt.Sprintf("/restconf/data/sonic-route-map:sonic-route-map/ROUTE_MAP_SET/ROUTE_MAP_SET_LIST=%s", k)
			glog.Infof("route policy set {%s} is deleting", k)
			rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
			err := DeleteHandlerResolve(rsp)
			if err != nil {
				return err
			}
			glog.Infof("route policy set {%s} has deleted", k)
		}
	}
	return nil
}

func ConfigSONICStaticRoute(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-static-route:sonic-static-route"
	staticroutedata := t.SonicConfig[basic.SONICSTATICROUTEKEY].(sonicmodel.SonicStaticRoute)
	sonicstaticroute, err := json.Marshal(staticroutedata)
	if err != nil {
		glog.Errorf("bgp root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicstaticroute)
	glog.Info("static route config data sending")
	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, b)
	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Info("static route config data has completed")
	return nil
}

func RemoveSONICStaticRoute(t *tcontext.Tcontext) error {
	staticroutedata := t.SonicConfig[basic.SONICSTATICROUTEKEY].(sonicmodel.SonicStaticRoute)
	for _, v := range staticroutedata.StaticRoute.StaticRouteListEntry.StaticRouteList {
		prefixencode := url.QueryEscape(v.Prefix)
		urlsuffix := fmt.Sprintf("/restconf/data/sonic-static-route:sonic-static-route/STATIC_ROUTE/STATIC_ROUTE_LIST=%s,%s", v.VrfName, prefixencode)
		glog.Infof("static route {%s} {%s} is deleting", v.VrfName, v.Prefix)
		rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
		err := DeleteHandlerResolve(rsp)
		if err != nil {
			return err
		}
		glog.Infof("static route {%s} {%s} has deleted", v.VrfName, v.Prefix)
	}
	return nil
}

func ConfigSONICOSPFv2(t *tcontext.Tcontext) error {
	urlsuffix := "/restconf/data/sonic-ospfv2:sonic-ospfv2"
	ospfv2data := t.SonicConfig[basic.SONICOSPFKEY].(sonicmodel.SonicOspfv2)
	sonicospfv2, err := json.Marshal(ospfv2data)
	if err != nil {
		glog.Errorf("bgp root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicospfv2)
	glog.Infof("ospfv2 config data sending,data %s", b.String())
	rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, b)
	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Info("ospfv2 config data has completed")
	return nil
}

func RemoveSONICOSPFv2(t *tcontext.Tcontext) error {
	ospfdata := t.SonicConfig[basic.SONICOSPFKEY].(sonicmodel.SonicOspfv2)
	for _, v := range ospfdata.SonicOspfv2tables.OSPFv2Router.OSPFv2RouterList {
		urlsuffix := fmt.Sprintf("/restconf/data/sonic-ospfv2:sonic-ospfv2/OSPFV2_ROUTER/OSPFV2_ROUTER_LIST=%s", v.VrfName)
		glog.Infof("ospf router {%s} is deleting", v.VrfName)
		rsp := httpclient.SONICCLENT.SendSonicRequest(t.Operation, urlsuffix, nil)
		err := DeleteHandlerResolve(rsp)
		if err != nil {
			return err
		}
		glog.Infof("ospf router {%s} has deleted", v.VrfName)
	}
	return nil
}

func CommonRemoveRouteRedistribute(vrfname string, src_protocol []string, dst_protocol string, addr_family []string) error {
	for _, family := range addr_family {
		for _, src := range src_protocol {
			urlsuffix := fmt.Sprintf("/restconf/data/sonic-route-common:sonic-route-common/ROUTE_REDISTRIBUTE/ROUTE_REDISTRIBUTE_LIST=%s,%s,%s,%s",
				vrfname, src, dst_protocol, family)
			rsp := httpclient.SONICCLENT.SendSonicRequest(basic.OPERREMOVE, urlsuffix, nil)
			err := DeleteHandlerResolve(rsp)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ConfigHandlerResolve(rsp *httpclient.SonicResp) error {
	code := rsp.Code
	if code > basic.DefaultSuccess {
		if len(rsp.ErrorMessage.SErrors.ErrorList) <= 0 {
			return errors.New("opreation failed,sonic resp body nil,unknown error")
		} else {
			apptag := rsp.ErrorMessage.SErrors.ErrorList[0].ErrorAppTag
			if apptag != "" && apptag == "vni-already-configured" {
				return nil
			}
			rsperr := rsp.ErrorMessage.SErrors.ErrorList[0].ErrorMessage
			if rsperr == "" {
				rsperr = rsp.ErrorMessage.SErrors.ErrorList[0].ErrorTag
			}
			errmsg := fmt.Sprintf("Opreation failed:%s", rsperr)
			glog.Error(errmsg)
			return errors.New(errmsg)
		}
	} else {
		return nil
	}
}

func DeleteHandlerResolve(rsp *httpclient.SonicResp) error {
	code := rsp.Code
	if code > basic.DefaultSuccess {
		if len(rsp.ErrorMessage.SErrors.ErrorList) <= 0 {
			return errors.New("opreation failed,sonic resp body nil,unknown error")
		} else {
			rsperr := rsp.ErrorMessage.SErrors.ErrorList[0].ErrorMessage
			if rsperr == "" {
				rsperr = rsp.ErrorMessage.SErrors.ErrorList[0].ErrorTag
			} else {
				//当做删除成功
				if rsperr == basic.RESOURCENOTFOUND {
					return nil
				}
			}
			errmsg := fmt.Sprintf("Opreation failed:%s", rsperr)
			glog.Error(errmsg)
			return errors.New(errmsg)
		}
	} else {
		return nil
	}
}

func SaveIndexToRedis(t *tcontext.Tcontext, key string) {
	for k, v := range t.SonicConfig[basic.SONICINDEX].(map[string]int) {
		if strings.HasPrefix(k, key) {
			redisclient.IndexSet(k, v)
		}
	}
}

//单独调用配置 目前在华三中没有这块的配置
//{配置loopback接口之前先声明loopback id实例}

func ConfigSONICLoopBack(loopbackname string) error {
	urlsuffix := "/restconf/data/sonic-loopback:sonic-loopback/LOOPBACK/LOOPBACK_LIST"
	loopbacknode := sonicmodel.Loopback{Name: loopbackname}
	var loopbackroot sonicmodel.SonicLoopback
	loopbackroot.LoopbackList = append(loopbackroot.LoopbackList, loopbacknode)
	sonicvlan, err := json.Marshal(loopbackroot)
	if err != nil {
		glog.Errorf("loopback root marshal error:%s", err)
		return err
	}
	b := bytes.NewBuffer(sonicvlan)
	glog.Info("loopback config is sending")
	rsp := httpclient.SONICCLENT.SendSonicRequest(basic.OPERMERGE, urlsuffix, b)
	err = ConfigHandlerResolve(rsp)
	if err != nil {
		return err
	}
	glog.Info("loopback config send completed")
	return nil
}
