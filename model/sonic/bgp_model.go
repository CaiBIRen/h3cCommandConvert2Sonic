package sonicmodel

type BgpGlobalsList struct {
	VrfName                     string `json:"vrf_name,omitempty"`
	RouterID                    string `json:"router_id,omitempty"`
	LocalASN                    int    `json:"local_asn,omitempty"`
	AlwaysCompareMED            bool   `json:"always_compare_med,omitempty"`
	LoadBalanceMPRelax          bool   `json:"load_balance_mp_relax,omitempty"`
	GracefulRestartEnable       bool   `json:"graceful_restart_enable,omitempty"`
	GRPreserveFWState           bool   `json:"gr_preserve_fw_state,omitempty"`
	GRRestartTime               int    `json:"gr_restart_time,omitempty"`
	GRStaleRoutesTime           int    `json:"gr_stale_routes_time,omitempty"`
	ExternalCompareRouterID     bool   `json:"external_compare_router_id,omitempty"`
	IgnoreASPathLength          bool   `json:"ignore_as_path_length,omitempty"`
	LogNbrStateChanges          bool   `json:"log_nbr_state_changes,omitempty"`
	RRClusterID                 string `json:"rr_cluster_id,omitempty"`
	RRAllowOutPolicy            bool   `json:"rr_allow_out_policy,omitempty"`
	DisableEBGPConnectedRTCheck bool   `json:"disable_ebgp_connected_rt_check,omitempty"`
	FastExternalFailover        bool   `json:"fast_external_failover,omitempty"`
	NetworkImportCheck          bool   `json:"network_import_check,omitempty"`
	GracefulShutdown            bool   `json:"graceful_shutdown,omitempty"`
	RRClntToClntReflection      bool   `json:"rr_clnt_to_clnt_reflection,omitempty"`
	DeterministicMED            bool   `json:"deterministic_med,omitempty"`
	MEDConfed                   bool   `json:"med_confed,omitempty"`
	MEDMissingASWorst           bool   `json:"med_missing_as_worst,omitempty"`
	CompareConfedASPath         bool   `json:"compare_confed_as_path,omitempty"`
	ASPathMPASSet               bool   `json:"as_path_mp_as_set,omitempty"`
	DefaultIPv4Unicast          bool   `json:"default_ipv4_unicast,omitempty"`
	DefaultLocalPreference      int    `json:"default_local_preference,omitempty"`
	DefaultShowHostname         bool   `json:"default_show_hostname,omitempty"`
	DefaultShutdown             bool   `json:"default_shutdown,omitempty"`
	MaxMEDAdmin                 bool   `json:"max_med_admin,omitempty"`
	ConfedID                    int    `json:"confed_id,omitempty"`
	ConfedPeers                 []int  `json:"confed_peers,omitempty"`
	Keepalive                   int    `json:"keepalive,omitempty"`
	Holdtime                    int    `json:"holdtime,omitempty"`
}

// 定义 BGP_GLOBALS 对象结构体
type BgpGlobals struct {
	BGP_GLOBALS_LIST []BgpGlobalsList `json:"BGP_GLOBALS_LIST,omitempty"`
}

// 定义 BGP_GLOBALS_AF 列表中的元素结构体
type BgpGlobalsAFList struct {
	VrfName                          string   `json:"vrf_name,omitempty"`
	AFISAFI                          string   `json:"afi_safi,omitempty"`
	MaxEBGPPaths                     int      `json:"max_ebgp_paths,omitempty"`
	MaxIBGPPaths                     int      `json:"max_ibgp_paths,omitempty"`
	ImportVRF                        string   `json:"import_vrf,omitempty"`
	ImportVRFRouteMap                string   `json:"import_vrf_route_map,omitempty"`
	RouteDownloadFilter              string   `json:"route_download_filter,omitempty"`
	EBGPRouteDistance                int      `json:"ebgp_route_distance,omitempty"`
	IBGPRouteDistance                int      `json:"ibgp_route_distance,omitempty"`
	LocalRouteDistance               int      `json:"local_route_distance,omitempty"`
	IBGPEqualClusterLength           bool     `json:"ibgp_equal_cluster_length,omitempty"`
	RouteFlapDampen                  bool     `json:"route_flap_dampen,omitempty"`
	RouteFlapDampenHalfLife          int      `json:"route_flap_dampen_half_life,omitempty"`
	RouteFlapDampenReuseThreshold    int      `json:"route_flap_dampen_reuse_threshold,omitempty"`
	RouteFlapDampenSuppressThreshold int      `json:"route_flap_dampen_suppress_threshold,omitempty"`
	RouteFlapDampenMaxSuppress       int      `json:"route_flap_dampen_max_suppress,omitempty"`
	AdvertiseDefaultGW               bool     `json:"advertise-default-gw,omitempty"`
	AdvertiseSVIIP                   bool     `json:"advertise-svi-ip,omitempty"`
	RouteDistinguisher               string   `json:"route-distinguisher,omitempty"`
	ImportRTS                        []string `json:"import-rts,omitempty"`
	ExportRTS                        []string `json:"export-rts,omitempty"`
	AdvertiseAllVNI                  bool     `json:"advertise-all-vni,omitempty"`
	AdvertiseIPv4Unicast             bool     `json:"advertise-ipv4-unicast,omitempty"`
	AdvertiseIPv6Unicast             bool     `json:"advertise-ipv6-unicast,omitempty"`
	DefaultOriginateIPv4             bool     `json:"default-originate-ipv4,omitempty"`
	DefaultOriginateIPv6             bool     `json:"default-originate-ipv6,omitempty"`
	AutoRT                           string   `json:"autort,omitempty"`
	DADEnabled                       bool     `json:"dad-enabled,omitempty"`
	DADMaxMoves                      int      `json:"dad-max-moves,omitempty"`
	DADTime                          int      `json:"dad-time,omitempty"`
	DADFreeze                        string   `json:"dad-freeze,omitempty"`
}

// 定义 BGP_GLOBALS_AF 对象结构体
type BgpGlobalsAF struct {
	BGP_GLOBALS_AF_LIST []BgpGlobalsAFList `json:"BGP_GLOBALS_AF_LIST,omitempty"`
}

// 定义 BGP_GLOBALS_AF_NETWORK 列表中的元素结构体
type BgpGlobalsAFNetworkList struct {
	VrfName  string `json:"vrf_name,omitempty"`
	AFISAFI  string `json:"afi_safi,omitempty"`
	IPPrefix string `json:"ip_prefix,omitempty"`
	Policy   string `json:"policy,omitempty"`
	Backdoor bool   `json:"backdoor,omitempty"`
}

// 定义 BGP_GLOBALS_AF_NETWORK 对象结构体
type BgpGlobalsAFNetwork struct {
	BGP_GLOBALS_AF_NETWORK_LIST []BgpGlobalsAFNetworkList `json:"BGP_GLOBALS_AF_NETWORK_LIST,omitempty"`
}

// 定义 JSON 根结构体
type SonicBGPGlobal struct {
	BGP_GLOBALS            BgpGlobals          `json:"BGP_GLOBALS,omitempty"`
	BGP_GLOBALS_AF         BgpGlobalsAF        `json:"BGP_GLOBALS_AF,omitempty"`
	BGP_GLOBALS_AF_NETWORK BgpGlobalsAFNetwork `json:"BGP_GLOBALS_AF_NETWORK,omitempty"`
}

type BGProot struct {
	Sonicbgpglobal SonicBGPGlobal `json:"sonic-bgp-global:sonic-bgp-global,omitempty"`
}

