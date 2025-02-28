package sonicmodel

type VlanInterface struct {
	VlanName             string `json:"vlanName,omitempty"`
	NatZone              int    `json:"nat_zone,omitempty"`
	VrfName              string `json:"vrf_name,omitempty"`
	Ipv6UseLinkLocalOnly string `json:"ipv6_use_link_local_only,omitempty"`
	Unnumbered           string `json:"unnumbered,omitempty"`
}


// 定义 VLAN 接口对象结构体
type VLAN_INTERFACE struct {
	VLAN_INTERFACE_LIST        []VlanInterface  `json:"VLAN_INTERFACE_LIST,omitempty"`
}

// 定义顶层结构体 SonicVLANInterface
type SonicVLANInterface struct {
	VLAN_INTERFACE VLAN_INTERFACE `json:"VLAN_INTERFACE,omitempty"`
}

// 定义 JSON 根结构体 VlanInterfaceroot
type VlanInterfaceroot struct {
	SonicVLANInterface SonicVLANInterface `json:"sonic-vlan-interface:sonic-vlan-interface,omitempty"`
}

type VLANInterfaceIPAddrList struct {
	VLANINTERFACEIPADDRLIST []VLANInterfaceIPAddr `json:"sonic-vlan-interface:VLAN_INTERFACE_IPADDR_LIST"`
}

// VLANInterfaceIPAddr represents each element in the VLAN_INTERFACE_IPADDR_LIST slice.
type VLANInterfaceIPAddr struct {
	VlanName  string `json:"vlanName"`
	IpPrefix  string `json:"ip_prefix"`
	Secondary bool   `json:"secondary"`
}
