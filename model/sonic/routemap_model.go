package sonicmodel

type SonicRoutemaproot struct {
	SonicRouteMap SonicRouteMap `json:"sonic-route-map:sonic-route-map"`
}

type SonicRouteMap struct {
	RoutemapSet Routemapset `json:"ROUTE_MAP_SET"`
	RouteMap    RouteMap    `json:"ROUTE_MAP"`
}

type Routemapset struct {
	RoutemapsetList []RoutemapsetEntry `json:"ROUTE_MAP_SET_LIST"`
}

type RoutemapsetEntry struct {
	Name string `json:"name"`
}

// RouteMap contains a list of route map entries.
type RouteMap struct {
	RouteMapList []RouteMapEntry `json:"ROUTE_MAP_LIST"`
}

// RouteMapEntry defines the structure of each entry in the route map.
type RouteMapEntry struct {
	RouteMapName                string `json:"route_map_name"`
	StmtName                    int    `json:"stmt_name"`
	RouteOperation              string `json:"route_operation"`
	MatchPrefixSet              string `json:"match_prefix_set,omitempty"`
	MatchIPv6PrefixSet          string `json:"match_ipv6_prefix_set,omitempty"`
	MatchTag                    []int  `json:"match_tag,omitempty"`
	MatchEvpnDefaultType5Route  bool   `json:"match_evpn_default_type5_route,omitempty"`
	MatchEvpnAdvertiseRouteType string `json:"match_evpn_advertise_route_type,omitempty"`
	MatchEvpnVniNumber          int    `json:"match_evpn_vni_number,omitempty"`
	SetLocalPref                int    `json:"set_local_pref,omitempty"`
	SetNextHop                  string `json:"set_next_hop,omitempty"`
	SetIPv6NextHopGlobal        string `json:"set_ipv6_next_hop_global,omitempty"`
	SetIPv6NextHopPreferGlobal  bool   `json:"set_ipv6_next_hop_prefer_global,omitempty"`
}

type SonicRoutingPolicySets struct {
	SonicRoutingPolicySetsWrapper SonicRoutingPolicySetsWrapper `json:"sonic-routing-policy-sets:sonic-routing-policy-sets,omitempty"`
}

// SonicRoutingPolicySetsWrapper wraps the actual content of the routing policy sets.
type SonicRoutingPolicySetsWrapper struct {
	PrefixSet *PrefixSet `json:"PREFIX_SET,omitempty"`
	Prefix    *Prefix    `json:"PREFIX,omitempty"`
	AsPathSet *AsPathSet `json:"AS_PATH_SET,omitempty"`
}

// PrefixSet contains a list of prefix set entries.
type PrefixSet struct {
	PrefixSetList []PrefixSetEntry `json:"PREFIX_SET_LIST,omitempty"`
}

// PrefixSetEntry defines the structure of each entry in the prefix set.
type PrefixSetEntry struct {
	Name string `json:"name,omitempty"`
	Mode string `json:"mode,omitempty"`
}

// Prefix contains a list of prefix entries.
type Prefix struct {
	PrefixList []PrefixEntry `json:"PREFIX_LIST,omitempty"`
}

// PrefixEntry defines the structure of each entry in the prefix list.
type PrefixEntry struct {
	SetName         string `json:"set_name,omitempty"`
	SequenceNumber  int    `json:"sequence_number,omitempty"`
	IPPrefix        string `json:"ip_prefix,omitempty"`
	MasklengthRange string `json:"masklength_range,omitempty"`
	Action          string `json:"action,omitempty"`
}

// AsPathSet contains a list of AS path set entries.
type AsPathSet struct {
	AsPathSetList []AsPathSetEntry `json:"AS_PATH_SET_LIST,omitempty"`
}

// AsPathSetEntry defines the structure of each entry in the AS path set.
type AsPathSetEntry struct {
	Name            string   `json:"name,omitempty"`
	Action          string   `json:"action,omitempty"`
	AsPathSetMember []string `json:"as_path_set_member,omitempty"`
}
