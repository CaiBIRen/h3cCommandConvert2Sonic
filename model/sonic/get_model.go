package sonicmodel

// Device 代表整个设备的信息
type Device struct {
	PhysicalEntities PhysicalEntities
	Base             Base
}

// PhysicalEntities 代表设备中的物理实体集合
type PhysicalEntities struct {
	Entity Entity
}

// Base 代表设备的基本信息
type Base struct {
	HostName        string
	HostDescription string
}

// Entity 代表设备中的单个物理实体
type Entity struct {
	SoftwareRev string
	Model       string
}
type OpenConfigLLDP struct {
	Interface []Interface `json:"openconfig-lldp:interface,omitempty" mapstructure:"openconfig-lldp:interface"`
}

// Interface 代表一个接口及其邻居信息
type Interface struct {
	Name      string    `json:"name,omitempty" mapstructure:"name"`
	Neighbors Neighbors `json:"neighbors,omitempty" mapstructure:"neighbors"`
}

// Neighbors 代表邻居信息集合
type Neighbors struct {
	Neighbor []Neighbor `json:"neighbor,omitempty" mapstructure:"neighbor"`
}

// Neighbor 代表单个邻居的信息
type Neighbor struct {
	Id    string `json:"id,omitempty" mapstructure:"id"`
	State State  `json:"state,omitempty" mapstructure:"state"`
}

// State 代表邻居的状态信息
type State struct {
	ChassisID         string `json:"chassis-id,omitempty" mapstructure:"chassis-id"`
	ChassisIDType     string `json:"chassis-id-type,omitempty" mapstructure:"chassis-id-type"`
	Id                string `json:"id,omitempty" mapstructure:"id"`
	ManagementAddress string `json:"management-address,omitempty" mapstructure:"management-address"`
	PortDescription   string `json:"port-description,omitempty" mapstructure:"port-description"`
	PortID            string `json:"port-id,omitempty" mapstructure:"port-id"`
	PortIDType        string `json:"port-id-type,omitempty" mapstructure:"port-id-type"`
	SystemDescription string `json:"system-description,omitempty" mapstructure:"system-description"`
	SystemName        string `json:"system-name,omitempty" mapstructure:"system-name"`
	TTL               int    `json:"ttl,omitempty" mapstructure:"ttl"`
}

type InterfaceState struct {
	AdminStatus  string `json:"admin-status,omitempty" mapstructure:"admin-status"`
	Description  string `json:"description,omitempty" mapstructure:"description"`
	Enabled      bool   `json:"enabled,omitempty" mapstructure:"enabled"`
	Mtu          int    `json:"mtu,omitempty" mapstructure:"mtu"`
	Name         string `json:"name,omitempty" mapstructure:"name"`
	RateInterval int    `json:"openconfig-interfaces-ext:rate-interval,omitempty" mapstructure:"openconfig-interfaces-ext:rate-interval"`
	MAC          string
	OperStatus   string `json:"oper-status,omitempty" mapstructure:"oper-status"`
}

// Interface represents an interface.
type Interface1 struct {
	Name     string         `json:"name,omitempty" mapstructure:"name"`
	IntState InterfaceState `json:"state,omitempty" mapstructure:"state"`
}

// Interfaces represents a collection of interfaces.
type OpenconfigInterface struct {
	Interface []Interface1 `json:"interface,omitempty" mapstructure:"interface"`
}

type OpenconfigInterfaces struct {
	OpenconfigInterface OpenconfigInterface `mapstructure:"openconfig-interfaces:interfaces"`
}

type SonicDeviceMetadata struct {
	MAC string `json:"sonic-device-metadata:mac" mapstructure:"sonic-device-metadata:mac"`
}

type BGPGlobalConfigASN struct {
	LocalASN int `json:"sonic-bgp-global:local_asn" mapstructure:"sonic-bgp-global:local_asn"`
}

type PortChannelMemberList struct {
	Ifname string `json:"ifname" mapstructure:"ifname"` // 端口号
	Name   string `json:"name" mapstructure:"name"`     // 端口通道名称
}

// PortChannelMembers 代表包含端口通道成员列表的结构体
type PortChannelMembers struct {
	PortChannelMemberList []PortChannelMemberList `json:"sonic-portchannel:PORTCHANNEL_MEMBER_LIST" mapstructure:"sonic-portchannel:PORTCHANNEL_MEMBER_LIST"`
}

type LAGTableItem struct {
	Name       string `json:"lagname" mapstructure:"lagname"` // 端口通道名称
	OperStatus string `json:"oper_status" mapstructure:"oper_status"`
	MAC        string
}

// PortChannelList 代表包含端口通道条目的列表
type PortChannelList struct {
	LAGTableList []LAGTableItem `json:"sonic-portchannel:LAG_TABLE_LIST" mapstructure:"sonic-portchannel:LAG_TABLE_LIST"`
}

type Port struct {
	AdminStatus      string `json:"admin_status" mapstructure:"admin_status"`             // 管理状态
	Alias            string `json:"alias" mapstructure:"alias"`                           // 别名
	Description      string `json:"description" mapstructure:"description"`               // 描述
	FEC              string `json:"fec" mapstructure:"fec"`                               // FEC
	Ifname           string `json:"ifname" mapstructure:"ifname"`                         // 接口名称
	Index            int    `json:"index" mapstructure:"index"`                           // 索引
	Lanes            string `json:"lanes" mapstructure:"lanes"`                           // 车道
	MTU              int    `json:"mtu" mapstructure:"mtu"`                               // 最大传输单元
	OperStatus       string `json:"oper_status" mapstructure:"oper_status"`               // 操作状态
	PortLoadInterval int    `json:"port_load_interval" mapstructure:"port_load_interval"` // 端口负载间隔
	Speed            string `json:"speed" mapstructure:"speed"`                           // 速度
	MAC              string
	IPV4addr         string
	IPV4mask         string
}

// PortTable 表示 PORT_TABLE 中的列表
type PortTable struct {
	PortTableList []Port `json:"sonic-port:PORT_TABLE_LIST" mapstructure:"sonic-port:PORT_TABLE_LIST"`
}

type PrefixList struct {
	Action          string `json:"action" mapstructure:"action"`
	IPPrefix        string `json:"ip_prefix" mapstructure:"ip_prefix"`
	MasklengthRange string `json:"masklength_range" mapstructure:"masklength_range"`
	SequenceNumber  int    `json:"sequence_number" mapstructure:"sequence_number"`
	SetName         string `json:"set_name" mapstructure:"set_name"`
}

// SonicRoutingPolicySets 包含了一个PrefixList类型的切片
type SonicRoutingPolicyPrefixList struct {
	PrefixLists []PrefixList `json:"sonic-routing-policy-sets:PREFIX_LIST" mapstructure:"sonic-routing-policy-sets:PREFIX_LIST"`
}

type SonicRoutingPolicySetList struct {
	Get_RoutePolicyMaps []Get_RouteMap `json:"sonic-route-map:ROUTE_MAP_LIST" mapstructure:"sonic-route-map:ROUTE_MAP_LIST"`
}

type Get_RouteMap struct {
	RouteMapName   string `json:"route_map_name" mapstructure:"route_map_name"`
	StmtName       int    `json:"stmt_name" mapstructure:"stmt_name"`
	RouteOperation string `json:"route_operation" mapstructure:"route_operation"`
}

// 定义VLAN_INTERFACE_LIST元素的结构体
type Get_VLANInterface struct {
	VlanName string `json:"vlanName" mapstructure:"vlanName"`
	VrfName  string `json:"vrf_name" mapstructure:"vrf_name"`
}

// 定义包含VLAN_INTERFACE_LIST的顶层结构体
type Get_VLANInterfaceList struct {
	VLAN_INTERFACE_LIST []Get_VLANInterface `json:"sonic-vlan-interface:VLAN_INTERFACE_LIST" mapstructure:"sonic-vlan-interface:VLAN_INTERFACE_LIST"`
}

type Get_VLANInterfaceIP struct {
	VlanName string `json:"vlanName" mapstructure:"vlanName"`
	IPPrefix string `json:"ip_prefix" mapstructure:"ip_prefix"`
}

// 定义包含VLAN_INTERFACE_LIST的顶层结构体
type Get_VLANInterfaceListIPs struct {
	VLAN_INTERFACE_LIST_IP []Get_VLANInterfaceIP `json:"sonic-vlan-interface:VLAN_INTERFACE_IPADDR_LIST" mapstructure:"sonic-vlan-interface:VLAN_INTERFACE_IPADDR_LIST"`
}

// 定义OSPFV2_ROUTER_LIST元素的结构体
type Get_OSPFv2Router struct {
	Description string `json:"description" mapstructure:"description"`
	Enable      bool   `json:"enable" mapstructure:"enable"`
	RouterID    string `json:"router-id" mapstructure:"router-id"`
	VrfName     string `json:"vrf_name" mapstructure:"vrf_name"`
}

// 定义包含OSPFV2_ROUTER_LIST的结构体
type Get_OSPFv2Routers struct {
	OSPFv2RouterList []Get_OSPFv2Router `json:"OSPFV2_ROUTER_LIST" mapstructure:"OSPFV2_ROUTER_LIST"`
}

// 顶层结构体
type Get_SonicOspfv2Router struct {
	OSPFv2Router Get_OSPFv2Routers `json:"sonic-ospfv2:OSPFV2_ROUTER" mapstructure:"sonic-ospfv2:OSPFV2_ROUTER"`
}

type Get_StaticRoute struct {
	Prefix  string `json:"prefix" mapstructure:"prefix"`
	VrfName string `json:"vrf-name" mapstructure:"vrf-name"`
}

// 定义包含STATIC_ROUTE_LIST的主结构体
type Get_StaticRoutes struct {
	StaticRouteList []Get_StaticRoute `json:"sonic-static-route:STATIC_ROUTE_LIST" mapstructure:"sonic-static-route:STATIC_ROUTE_LIST"`
}

type Get_VniConfig struct {
	L3vni int `json:"sonic-vrf:vni" mapstructure:"sonic-vrf:vni"`
}
