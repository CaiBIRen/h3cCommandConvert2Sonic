package h3cmodel

import "encoding/xml"

// BGP 结构体
type BGP struct {
	XMLName       xml.Name       `xml:"BGP"`
	VRFs          *VRFs          `xml:"VRFs,omitempty"`
	Familys       *Familys       `xml:"Familys"`
	Networks      *Networks      `xml:"Networks"`
	Redistributes *Redistributes `xml:"Redistributes"`
	//get
	Instances *Instances `xml:"Instances,omitempty"`
}

// VRFs 结构体
type VRFs struct {
	BGPVRF []BGPVRF `xml:"VRF,omitempty"`
}

// VRF 结构体
type BGPVRF struct {
	Name string `xml:"Name,omitempty"`
	VRF  string `xml:"VRF,omitempty"`
}

type Familys struct {
	Family []Family `xml:"Family"`
}

type Family struct {
	Name                    string              `xml:"Name,omitempty"`
	VRF                     string              `xml:"VRF,omitempty"`
	Type                    int                 `xml:"Type,omitempty"`
	Preference              Preference          `xml:"Preference"`
	BalanceAsPathNeglect    int                 `xml:"BalanceAsPathNeglect,omitempty"`
	BalanceAsPathRelax      int                 `xml:"BalanceAsPathRelax,omitempty"`
	Balance                 Balance             `xml:"Balance"`
	DefaultRtImport         int                 `xml:"DefaultRtImport,omitempty"`
	AdvertiseEvpnRoute      AdvertiseEvpnRoute  `xml:"AdvertiseEvpnRoute"`
	AdvertiseL2vpnEvpnRoute int                 `xml:"AdvertiseL2vpnEvpnRoute,omitempty"`
	AdvertiseL3vpnRoute     AdvertiseL3vpnRoute `xml:"AdvertiseL3vpnRoute"`
	ImportEvpnMacIp         int                 `xml:"ImportEvpnMacIp,omitempty"`
}

// Preference 结构体
type Preference struct {
	Ebgp  int `xml:"Ebgp,omitempty"`
	Ibgp  int `xml:"Ibgp,omitempty"`
	Local int `xml:"Local,omitempty"`
}

// Balance 结构体
type Balance struct {
	MaxBalance          int  `xml:"MaxBalance,omitempty"`
	BalanceNexthop      bool `xml:"BalanceNexthop,omitempty"`
	MaxEBGPBalance      int  `xml:"MaxEBGPBalance,omitempty"`
	EBGPBalanceNexthop  bool `xml:"EBGPBalanceNexthop,omitempty"`
	MaxIBGPBalance      int  `xml:"MaxIBGPBalance,omitempty"`
	IBGPBalanceNexthop  bool `xml:"IBGPBalanceNexthop,omitempty"`
	MaxEIBGPBalance     int  `xml:"MaxEIBGPBalance,omitempty"`
	EIBGPBalanceNexthop bool `xml:"EIBGPBalanceNexthop,omitempty"`
}

// AdvertiseEvpnRoute 结构体
type AdvertiseEvpnRoute struct {
	AdvertiseEvpnEnable    int    `xml:"AdvertiseEvpnEnable,omitempty"`
	AdvertiseEvpnPolicy    string `xml:"AdvertiseEvpnPolicy,omitempty"`
	AdvertiseEvpnReplaceRt int    `xml:"AdvertiseEvpnReplaceRt,omitempty"`
}

// AdvertiseL3vpnRoute 结构体
type AdvertiseL3vpnRoute struct {
	AdvertiseL3vpnEnable    int    `xml:"AdvertiseL3vpnEnable,omitempty"`
	AdvertiseL3vpnPolicy    string `xml:"AdvertiseL3vpnPolicy,omitempty"`
	AdvertiseL3vpnReplaceRt int    `xml:"AdvertiseL3vpnReplaceRt,omitempty"`
}

type Networks struct {
	Network []Network `xml:"Network"`
}

// Network 结构体
type Network struct {
	Name        string `xml:"Name,omitempty"`
	VRF         string `xml:"VRF,omitempty"`
	Family      int    `xml:"Family,omitempty"`
	IpAddress   string `xml:"IpAddress,omitempty"`
	Mask        int    `xml:"Mask,omitempty"`
	RoutePolicy string `xml:"RoutePolicy,omitempty"`
}

type Redistributes struct {
	Redist []Redist `xml:"Redist"`
}

type Redist struct {
	Name        string `xml:"Name,omitempty"`
	VRF         string `xml:"VRF,omitempty"`
	Family      int    `xml:"Family,omitempty"`
	Protocol    int    `xml:"Protocol,omitempty"`
	RedistName  string `xml:"RedistName,omitempty"`
	AllowDirect int    `xml:"AllowDirect,omitempty"`
	Med         uint32 `xml:"Med,omitempty"`
	RoutePolicy string `xml:"RoutePolicy,omitempty"`
}

// Instances 定义了BGP实例列表
type Instances struct {
	XMLName xml.Name   `xml:"Instances"`
	Items   []Instance `xml:"Instance,omitempty"`
}

// Instance 定义了单个BGP实例
type Instance struct {
	Name     string `xml:"Name,omitempty"`
	ASNumber string `xml:"ASNumber,omitempty"`
}
