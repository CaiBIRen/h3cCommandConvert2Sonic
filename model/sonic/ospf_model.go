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

type OSPFv2Router struct {
	VrfName     string `json:"vrf_name"`
	Enable      bool   `json:"enable"`
	Description string `json:"description"`
	RouterID    string `json:"router-id"`
}

// 定义包含OSPFV2_ROUTER_LIST的结构体
type OSPFv2Routers struct {
	OSPFv2RouterList []OSPFv2Router `json:"OSPFV2_ROUTER_LIST"`
}

// 定义OSPFV2_ROUTER_AREA_LIST元素的结构体
type OSPFv2RouterArea struct {
	VrfName     string `json:"vrf_name"`
	AreaID      string `json:"area-id"`
	Description string `json:"description"`
	Enable      bool   `json:"enable"`
}

// 定义包含OSPFV2_ROUTER_AREA_LIST的结构体
type OSPFv2RouterAreas struct {
	OSPFv2RouterAreaList []OSPFv2RouterArea `json:"OSPFV2_ROUTER_AREA_LIST"`
}

// 定义OSPFV2_ROUTER_DISTRIBUTE_ROUTE_LIST元素的结构体
type OSPFv2RouterDistributeRoute struct {
	VrfName    string `json:"vrf_name"`
	Protocol   string `json:"protocol"`
	Direction  string `json:"direction"`
	TableID    int    `json:"table-id"`
}

// 定义包含OSPFV2_ROUTER_DISTRIBUTE_ROUTE_LIST的结构体
type OSPFv2RouterDistributeRoutes struct {
	OSPFv2RouterDistributeRouteList []OSPFv2RouterDistributeRoute `json:"OSPFV2_ROUTER_DISTRIBUTE_ROUTE_LIST"`
}

type Sonicospfv2Tables struct {
	OSPFv2Router                OSPFv2Routers                `json:"OSPFV2_ROUTER"`
	OSPFv2RouterArea            OSPFv2RouterAreas            `json:"OSPFV2_ROUTER_AREA"`
	OSPFv2Interface             OSPFv2Interfaces             `json:"OSPFV2_INTERFACE"`
	OSPFv2RouterDistributeRoute OSPFv2RouterDistributeRoutes `json:"OSPFV2_ROUTER_DISTRIBUTE_ROUTE"`
}

// 定义顶层结构体
type SonicOspfv2 struct {
	SonicOspfv2tables Sonicospfv2Tables `json:"sonic-ospfv2:sonic-ospfv2"`
}
