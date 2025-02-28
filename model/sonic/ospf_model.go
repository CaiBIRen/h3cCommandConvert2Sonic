package sonicmodel

// 定义OSPFV2_INTERFACE_LIST元素的结构体
type OSPFv2Interface struct {
	Name        string `json:"name"`
	AreaID      string `json:"area-id"`
	Address     string `json:"address"`
	Enable      bool   `json:"enable"`
	NetworkType string `json:"network-type"`
}

// 定义包含OSPFV2_INTERFACE_LIST的结构体
type OSPFv2Interfaces struct {
	OSPFV2_INTERFACE_LIST []OSPFv2Interface `json:"OSPFV2_INTERFACE_LIST"`
}
type OSPFv2INTERFACE struct {
	OSPFv2INTERFACES OSPFv2Interfaces `json:"OSPFV2_INTERFACE"`
}

// 定义顶层结构体
type SonicOspfv2 struct {
	SonicOspfv2 OSPFv2INTERFACE `json:"sonic-ospfv2:sonic-ospfv2"`
}
