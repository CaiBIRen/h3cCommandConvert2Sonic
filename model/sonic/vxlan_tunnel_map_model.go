package sonicmodel

// 定义 VXLAN 隧道映射列表中的元素结构体
type VxlanTunnelMap struct {
	Name    string `json:"name,omitempty"`
	Mapname string `json:"mapname,omitempty"`
	Vlan    string `json:"vlan,omitempty"`
	Vni     int    `json:"vni,omitempty"`
}

// 定义 VXLAN 隧道映射对象结构体
type VXLAN_TUNNEL_MAP struct {
	VXLAN_TUNNEL_MAP_LIST []VxlanTunnelMap `json:"VXLAN_TUNNEL_MAP_LIST,omitempty"`
}

// 定义 JSON 根结构体 Vxlanroot
type Vxlanroot struct {
	SonicVxlan VXLAN_TUNNEL_MAP `json:"sonic-vxlan:VXLAN_TUNNEL_MAP,omitempty"`
}

type VxlanVlan struct {
	Vlan string `json:"sonic-vxlan:vlan"`
}
