package sonicmodel

// 定义LOOPBACK_INTERFACE_LIST元素的结构体
type LoopbackInterface struct {
	LoIfName             string `json:"loIfName"`
	NatZone              int    `json:"nat_zone"`
	VrfName              string `json:"vrf_name"`
	Ipv6UseLinkLocalOnly string `json:"ipv6_use_link_local_only"`
}

// 定义包含LOOPBACK_INTERFACE_LIST的结构体
type LoopbackInterfacesroot struct {
	LoopbackInterfaceList []LoopbackInterface `json:"sonic-loopback-interface:LOOPBACK_INTERFACE_LIST"`
}

type LoopbackInterfaceIPAddr struct {
	LoIfName  string `json:"loIfName"`
	IpPrefix  string `json:"ip_prefix"`
	Secondary bool   `json:"secondary"`
}

// 定义包含LOOPBACK_INTERFACE_IPADDR_LIST的结构体
type LoopbackInterfacesIPAddrList struct {
	LoopbackInterfaceIPAddrList []LoopbackInterfaceIPAddr `json:"sonic-loopback-interface:LOOPBACK_INTERFACE_IPADDR_LIST"`
}
