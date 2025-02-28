package sonicmodel

type SonicStaticRoute struct {
	StaticRoute StaticRoute `json:"sonic-static-route:sonic-static-route"`
}

type StaticRoute struct {
	StaticRouteListEntry StaticRouteListEntry `json:"STATIC_ROUTE"`
}

// StaticRoute contains a list of static route entries.
type StaticRouteListEntry struct {
	StaticRouteList []StaticRouteEntry `json:"STATIC_ROUTE_LIST"`
}

// StaticRouteEntry defines the structure of each static route entry.
type StaticRouteEntry struct {
	VrfName    string `json:"vrf-name"`
	Prefix     string `json:"prefix"`
	Nexthop    string `json:"nexthop"`
	Ifname     *string `json:"ifname"`
	Distance   *string `json:"distance"`
	NexthopVrf string `json:"nexthop-vrf"`
	Blackhole  *string `json:"blackhole"`
	Tag        *string `json:"tag"`
}
