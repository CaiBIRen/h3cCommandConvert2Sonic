package tcontext

import (
	"sonic-unis-framework/basic"
	"sonic-unis-framework/model"
	sonicmodel "sonic-unis-framework/model/sonic"
	"strings"

	"github.com/coreos/pkg/capnslog"
)

var glog = capnslog.NewPackageLogger("sonic-unis-framework", "TCONTEXT")

// 客户端的每个请求转换为一个context
type Tcontext struct {
	Messageid             string
	Operation             string
	Features              map[string]interface{}
	DiscreteConfiguration map[string]map[string]interface{} /* map [feature] map[config_index] config*/
	SonicConfig           map[string]interface{}
	Cachedata             map[string]string
	Err                   error
}

func NewTcontext() Tcontext {
	tcontext := Tcontext{
		Features:              make(map[string]interface{}),
		Cachedata:             make(map[string]string),
		DiscreteConfiguration: make(map[string]map[string]interface{}),
		SonicConfig:           make(map[string]interface{}),
	}
	// tcontext.SonicConfig[basic.SONICINDEX] = make(map[string]int)
	return tcontext
}

func (context *Tcontext) DiscreteConfigurationIntegration() error {
	for feature, configurations := range context.DiscreteConfiguration {
		if len(configurations) == 0 {
			continue
		}
		switch feature {
		case basic.SONICVLANKEY:
			glog.Info("Vlan Intergration")
			vlanroot := VlanIntergration(configurations)
			context.SonicConfig[basic.SONICVLANKEY] = vlanroot
		case basic.SONICVRFKEY:
			glog.Info("Vrf Intergration")
			vrfroot := VrfIntergration(configurations)
			context.SonicConfig[basic.SONICVRFKEY] = vrfroot
		case basic.SONICBGPKEY:
			glog.Info("BGP Intergration")
			bgproot := BGPIntergration(configurations)
			context.SonicConfig[basic.SONICBGPKEY] = bgproot
		case basic.SONICVLANINTERFACEKEY:
			glog.Info("VlanInterface Intergration")
			vlaninterfaceroot := VlanInterfaceIntergration(configurations)
			context.SonicConfig[basic.SONICVLANINTERFACEKEY] = vlaninterfaceroot
		case basic.SONICROUTECOMMONKEY:
			glog.Info("Route Common Intergration")
			routecommonroot := RouteCommonIntergration(configurations)
			context.SonicConfig[basic.SONICROUTECOMMONKEY] = routecommonroot
		case basic.SONICVXLANKEY:
			glog.Info("vxlan Intergration")
			vxlanroot := VxlanIntergration(configurations)
			context.SonicConfig[basic.SONICVXLANKEY] = vxlanroot
		case basic.SONICSTATICROUTEKEY:
			glog.Info("static route Intergration")
			staticrouteroot := StaticRouteIntergration(configurations)
			context.SonicConfig[basic.SONICSTATICROUTEKEY] = staticrouteroot
		case basic.SONICADDRESS:
			glog.Info("interface address Intergration")
			AddressIntergration(context, configurations)
		case basic.SONICINTERFACEMAC:
			glog.Info("interface mac Intergration")
			MACIntergration(context, configurations)
		case basic.SONICROUTEMAPSETKEY:
			routemapsetroot := RoutemapSetIntergration(configurations)
			context.SonicConfig[basic.SONICROUTEMAPSETKEY] = routemapsetroot
		case basic.SONICROUTEMAPKEY:
			routemaproot := RoutemapIntergration(configurations)
			context.SonicConfig[basic.SONICROUTEMAPKEY] = routemaproot
		case basic.SONICOSPFKEY:
			ospfroot := OSPFv2Intergration(configurations)
			context.SonicConfig[basic.SONICOSPFKEY] = ospfroot
		}
	}
	glog.Infof("all sonic unit config in context\n %+v", context.SonicConfig)
	glog.Infof("Intergration end")
	return nil
}

func OSPFv2Intergration(config map[string]interface{}) sonicmodel.SonicOspfv2 {
	var ospfroot sonicmodel.SonicOspfv2
	for key, node := range config {
		childkey := strings.Split(key, "#")[1]
		switch childkey {
		case "OSPF_INTERFACE":
			ospfinterface := node.(sonicmodel.OSPFv2Interface)
			ospfroot.SonicOspfv2.OSPFv2INTERFACES.OSPFV2_INTERFACE_LIST = append(ospfroot.SonicOspfv2.OSPFv2INTERFACES.OSPFV2_INTERFACE_LIST, ospfinterface)
		}
	}
	return ospfroot
}

func RoutemapIntergration(config map[string]interface{}) sonicmodel.SonicRouteMap {
	var routemaproot sonicmodel.SonicRouteMap
	for key, node := range config {
		childkey := strings.Split(key, "#")[1]
		switch childkey {
		case "ROUTE_MAP":
			routemap := node.(sonicmodel.RouteMapEntry)
			routemaproot.RouteMap.RouteMapList = append(routemaproot.RouteMap.RouteMapList, routemap)
		}
	}
	return routemaproot
}

func RoutemapSetIntergration(config map[string]interface{}) sonicmodel.SonicRoutingPolicySets {
	var routemapsetroot sonicmodel.SonicRoutingPolicySets

	var prefixset sonicmodel.PrefixSet
	var prefix sonicmodel.Prefix

	for key, node := range config {
		childkey := strings.Split(key, "#")[1]
		switch childkey {
		case "PREFIX_NODE":
			prefixnode := node.(sonicmodel.PrefixEntry)
			prefix.PrefixList = append(prefix.PrefixList, prefixnode)
		case "IPV4_PREFIX_SET":
			ipv4prefixset := node.(sonicmodel.PrefixSetEntry)
			prefixset.PrefixSetList = append(prefixset.PrefixSetList, ipv4prefixset)
		case "IPV6_PREFIX_SET":
			ipv6prefixset := node.(sonicmodel.PrefixSetEntry)
			prefixset.PrefixSetList = append(prefixset.PrefixSetList, ipv6prefixset)
		}
	}

	if len(prefix.PrefixList) > 0 {
		routemapsetroot.SonicRoutingPolicySetsWrapper.Prefix = &prefix
	}

	if len(prefixset.PrefixSetList) > 0 {
		routemapsetroot.SonicRoutingPolicySetsWrapper.PrefixSet = &prefixset
	}
	return routemapsetroot
}

func VlanIntergration(config map[string]interface{}) sonicmodel.Vlanroot {
	var vlanroot sonicmodel.Vlanroot
	for key, node := range config {
		childkey := strings.Split(key, "#")[1]
		switch childkey {
		case "VLAN":
			vlan := node.(sonicmodel.VLANNode)
			vlanroot.SonicVLAN.VLAN.VLANList = append(vlanroot.SonicVLAN.VLAN.VLANList, vlan)
		}
	}
	return vlanroot
}

func VxlanIntergration(config map[string]interface{}) sonicmodel.Vxlanroot {
	var vxlanroot sonicmodel.Vxlanroot
	for key, node := range config {
		childkey := strings.Split(key, "#")[1]
		switch childkey {
		case "VXLAN_TUNNEL_MAP":
			vxlan := node.(sonicmodel.VxlanTunnelMap)
			vxlanroot.SonicVxlan.VXLAN_TUNNEL_MAP_LIST = append(vxlanroot.SonicVxlan.VXLAN_TUNNEL_MAP_LIST, vxlan)
		}
	}
	return vxlanroot
}

func VlanInterfaceIntergration(config map[string]interface{}) sonicmodel.VlanInterfaceroot {
	var vlaninterfaceroot sonicmodel.VlanInterfaceroot
	for key, node := range config {
		childkey := strings.Split(key, "#")[1]
		switch childkey {
		case "VLAN_INTERFACE":
			vlaninterface := node.(sonicmodel.VlanInterface)
			vlaninterfaceroot.SonicVLANInterface.VLAN_INTERFACE.VLAN_INTERFACE_LIST = append(vlaninterfaceroot.SonicVLANInterface.VLAN_INTERFACE.VLAN_INTERFACE_LIST, vlaninterface)
		}
	}
	return vlaninterfaceroot
}

func VrfIntergration(config map[string]interface{}) sonicmodel.Vrfroot {
	var vrfroot sonicmodel.Vrfroot
	for key, node := range config {
		childkey := strings.Split(key, "#")[1]
		switch childkey {
		case "VRF":
			vrf := node.(sonicmodel.Vrf)
			vrfroot.SonicVrf.VRF.VRF_LIST = append(vrfroot.SonicVrf.VRF.VRF_LIST, vrf)
		}
	}
	return vrfroot
}

func BGPIntergration(config map[string]interface{}) sonicmodel.BGProot {
	var bgproot sonicmodel.BGProot
	for key, node := range config {
		childkey := strings.Split(key, "#")[1]
		switch childkey {
		case "BGP_GLOBALS":
			bgpglobal := node.(sonicmodel.BgpGlobalsList)
			bgproot.Sonicbgpglobal.BGP_GLOBALS.BGP_GLOBALS_LIST = append(bgproot.Sonicbgpglobal.BGP_GLOBALS.BGP_GLOBALS_LIST, bgpglobal)
		case "BGP_GLOBALS_AF":
			bgpglobalaf := node.(sonicmodel.BgpGlobalsAFList)
			bgproot.Sonicbgpglobal.BGP_GLOBALS_AF.BGP_GLOBALS_AF_LIST = append(bgproot.Sonicbgpglobal.BGP_GLOBALS_AF.BGP_GLOBALS_AF_LIST, bgpglobalaf)
		case "BGP_GLOBALS_AF_NETWORK":
			bgpglobalafnetwork := node.(sonicmodel.BgpGlobalsAFNetworkList)
			bgproot.Sonicbgpglobal.BGP_GLOBALS_AF_NETWORK.BGP_GLOBALS_AF_NETWORK_LIST = append(bgproot.Sonicbgpglobal.BGP_GLOBALS_AF_NETWORK.BGP_GLOBALS_AF_NETWORK_LIST, bgpglobalafnetwork)
		}
	}
	return bgproot
}

func RouteCommonIntergration(config map[string]interface{}) sonicmodel.SonicRouteCommonroot {
	var routecommonroot sonicmodel.SonicRouteCommonroot
	for key, node := range config {
		childkey := strings.Split(key, "#")[1]
		switch childkey {
		case "ROUTECOMMONREDIST":
			routecommonredist := node.(sonicmodel.RouteRedistributenode)
			routecommonroot.SonicRouteCommon.RouteRedistributeList.RouteRedistributes = append(routecommonroot.SonicRouteCommon.RouteRedistributeList.RouteRedistributes, routecommonredist)
		}
	}
	return routecommonroot
}

func StaticRouteIntergration(config map[string]interface{}) sonicmodel.SonicStaticRoute {
	var staticrouteroot sonicmodel.SonicStaticRoute
	for key, node := range config {
		childkey := strings.Split(key, "#")[1]
		switch childkey {
		case "STATIC_ROUTE":
			staticroute := node.(sonicmodel.StaticRouteEntry)
			staticrouteroot.StaticRoute.StaticRouteListEntry.StaticRouteList = append(staticrouteroot.StaticRoute.StaticRouteListEntry.StaticRouteList, staticroute)
		}
	}
	return staticrouteroot
}

// distinguish interface type
func AddressIntergration(context *Tcontext, config map[string]interface{}) {
	var vlaniproot sonicmodel.VLANInterfaceIPAddrList
	for key, config := range config {
		childkey := strings.Split(key, "#")[1]
		switch childkey {
		case "VLAN_INTERFACE_IPADDR_LIST":
			vlanipnode := config.(sonicmodel.VLANInterfaceIPAddr)
			vlaniproot.VLANINTERFACEIPADDRLIST = append(vlaniproot.VLANINTERFACEIPADDRLIST, vlanipnode)
		}
	}
	if len(vlaniproot.VLANINTERFACEIPADDRLIST) > 0 {
		context.SonicConfig[basic.SONICVLANINTERFACEIPADDRKEY] = vlaniproot
	}
}

// netlink
func MACIntergration(context *Tcontext, config map[string]interface{}) {
	var maclist model.Mac_interface_list
	for k, _ := range config {
		elements := strings.Split(k, "@")
		var macnode model.Mac_interface
		macnode.Ifname = elements[0]
		macnode.Mac = elements[1]
		maclist.Mac_interfaces = append(maclist.Mac_interfaces, macnode)
	}
	if len(maclist.Mac_interfaces) > 0 {
		context.SonicConfig[basic.SONICINTERFACEMAC] = maclist
	}
}
