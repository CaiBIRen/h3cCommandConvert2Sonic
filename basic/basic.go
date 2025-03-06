package basic

import (
	"net"

	"github.com/coreos/pkg/capnslog"
)

var glog = capnslog.NewPackageLogger("sonic-unis-framework", "BASIC")
var DefaultUser string = "admin"
var DefaultPassword string = "YourPaSsWoRd"
var DefaultBGPLocalasn = 64515
var DefaultSuccess int = 204
var DefaultHttpErrorCode = 1000
var VLANBASE = 700
var VRFCAP = 299
var TUNNELNAME = "vtep1"
var OPERMERGE string = "merge"
var OPERREMOVE string = "remove"
var OPERGET string = "get"
var OPERACTION string = "action"
var RESOURCENOTFOUND = "Resource not found"

// 全局feature属性
var (
	SONICVLANKEY                    = "sonic_vlan"
	SONICVXLANKEY                   = "sonic_vxlan"
	SONICVRFKEY                     = "sonic_vrf"
	SONICVLANINTERFACEKEY           = "sonic_vlan_interface"
	SONICLOOPBACKKEY                = "sonic_loopback_interface"
	SONICVLANINTERFACEIPADDRKEY     = "sonic_vlan_interface_ip"
	SONICLOOPBACKINTERFACEIPADDRKEY = "sonic_loopback_interface_ip"
	SONICBGPKEY                     = "sonic_bgp"
	SONICSTATICROUTEKEY             = "sonic_staticroute"
	SONICROUTECOMMONKEY             = "sonic_routecommon"
	SONICINDEX                      = "sonic_index"
	SONICDEVICE                     = "sonic_device"
	SONICLLDP                       = "sonic_lldp"
	SONICPORT                       = "sonic_port"
	SONICSYSTEMID                   = "sonic_systemid"
	SONICPORTCHANNEL                = "sonic_portchannel"
	SONICPORTCHANNELMEMBERS         = "sonic_portchannel_member"
	SONICADDRESS                    = "sonic_address"
	SONICINTERFACEMAC               = "sonic_interface_mac"
	SONICROUTEMAPSETKEY             = "sonic_routemap_set"
	SONICROUTEMAPKEY                = "sonic_routemap"
	SONICOSPFKEY                    = "sonic_ospf"
)

// 子元素属性
var (
	SONICVLANELEMENT                    = "#VLAN"
	SONICVRFELEMENT                     = "#VRF"
	SONICBGPGLOBALELEMENT               = "#BGP_GLOBALS"
	SONICBGPGLOBALAFELEMENT             = "#BGP_GLOBALS_AF"
	SONICBGPGLOBALAFNETOWRKELEMENT      = "#BGP_GLOBALS_AF_NETWORK"
	SONICROUTECOMMONREDISTELEMENT       = "#ROUTECOMMONREDIST"
	SONICVLANINTERFACEELEMENT           = "#VLAN_INTERFACE"
	SONICLOOPBACKINTERFACEELEMENT       = "#LOOPBACK_INTERFACE"
	SONICVXLANTUNNELMAPELEMENT          = "#VXLAN_TUNNEL_MAP"
	SONICSTATICROUTEELEMENT             = "#STATIC_ROUTE"
	SONICVLANINTERFACEIPADDRELEMENT     = "#VLAN_INTERFACE_IPADDR_LIST"
	SONICLOOPBACKINTERFACEIPADDRELEMENT = "#LOOPBACK_INTERFACE_IPADDR_LIST"
	SONICINTERFACEMACELEMENT            = "#INTERFACE_MAC"
	SONICIPV4PREFIXSETELEMENT           = "#IPV4_PREFIX_SET"
	SONICIPV6PREFIXSETELEMENT           = "#IPV6_PREFIX_SET"
	SONICPREFIXNODEELEMENT              = "#PREFIX_NODE"
	SONICROUTEMAPSETELELMENT            = "#ROUTE_MAP_SET"
	SONICROUTEMAPELELMENT               = "#ROUTE_MAP"
	SONICOSPFINTERFACEELELMENT          = "#OSPF_INTERFACE"
	SONICOSPFINSTANCEELELMENT           = "#OSPF_Router"
	SONICOSPFAREAELELMENT               = "#OSPF_AREA"
	SONICOSPFREDISTELELMENT             = "#OSPF_ROUTER_DISTRIBUTE"
)

func FindAManagementIP() string {
	var addrs []net.Addr
	eth0, err := net.InterfaceByName("eth0")
	if err == nil {
		glog.Info("get eth0 info success")
		addrs, err = eth0.Addrs()
	}
	if err != nil {
		glog.Errorf("Could not read eth0 info; err=%v", err)
		return ""
	}

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err == nil && ip.To4() != nil {
			glog.Infof("get eth0 ip:%s", ip.String())
			return ip.String()
		}
	}

	glog.Warning("Could not find a management address!!")
	return ""
}
