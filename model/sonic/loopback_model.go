package sonicmodel

// 定义LOOPBACK_INTERFACE_LIST元素的结构体
type LoopbackInterface struct {
	LoIfName string  `json:"loIfName"`
	VrfName  *string `json:"vrf_name"`
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

// 定义LOOPBACK_LIST元素的结构体
type Loopback struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// 定义包含LOOPBACK_LIST的结构体
type SonicLoopback struct {
	LoopbackList []Loopback `json:"sonic-loopback:LOOPBACK_LIST"`
}
