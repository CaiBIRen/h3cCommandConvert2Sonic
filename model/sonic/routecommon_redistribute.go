package sonicmodel

type SonicRouteCommonroot struct {
	SonicRouteCommon RouteRedistribute `json:"sonic-route-common:sonic-route-common"`
}

// RouteRedistribute 代表路由重分发配置
type RouteRedistribute struct {
	RouteRedistributeList RouteRedistributeList `json:"ROUTE_REDISTRIBUTE"`
}

type RouteRedistributeList struct {
	RouteRedistributes []RouteRedistributenode `json:"ROUTE_REDISTRIBUTE_LIST"`
}

// RouteRedistributeList 代表路由重分发列表项
type RouteRedistributenode struct {
	VrfName     string   `json:"vrf_name"`
	SrcProtocol string   `json:"src_protocol"`
	DstProtocol string   `json:"dst_protocol"`
	AddrFamily  string   `json:"addr_family"`
	RouteMap    []string `json:"route_map"`
	Metric      int      `json:"metric"`
}
