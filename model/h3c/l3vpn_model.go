package h3cmodel

import "encoding/xml"

type L3vpn struct {
	XMLName  xml.Name `xml:"L3vpn"`
	L3vpnVRF L3vpnVRF `xml:"L3vpnVRF,omitempty"`
	L3vpnIf  L3vpnIf  `xml:"L3vpnIf,omitempty"`
	L3vpnRT  L3vpnRT  `xml:"L3vpnRT,omitempty"`
}

type L3vpnVRF struct {
	VRFs []VRF `xml:"VRF,omitempty"`
}

type L3vpnIf struct {
	Binds []Bind `xml:"Bind,omitempty"`
}

type L3vpnRT struct {
	RTs []RT `xml:"RT,omitempty"`
}

type VRF struct {
	VRF                   string `xml:"VRF"`
	VrfIndex              int    `xml:"VrfIndex,omitempty"`
	Description           string `xml:"Description,omitempty"`
	RD                    string `xml:"RD,omitempty"`
	ExportRoutePolicy     string `xml:"ExportRoutePolicy,omitempty"`
	EVPNExportRoutePolicy string `xml:"EVPNExportRoutePolicy,omitempty"`
	Ipv4ExportRoutePolicy string `xml:"Ipv4ExportRoutePolicy,omitempty"`
	Ipv6ExportRoutePolicy string `xml:"Ipv6ExportRoutePolicy,omitempty"`
	ImportRoutePolicy     string `xml:"ImportRoutePolicy,omitempty"`
	EVPNImportRoutePolicy string `xml:"EVPNImportRoutePolicy,omitempty"`
	Ipv4ImportRoutePolicy string `xml:"Ipv4ImportRoutePolicy,omitempty"`
	Ipv6ImportRoutePolicy string `xml:"Ipv6ImportRoutePolicy,omitempty"`
}

type Bind struct {
	VRF     string `xml:"VRF"`
	IfIndex string `xml:"IfIndex,omitempty"`
}

type RT struct {
	VRF           string `xml:"VRF"`
	AddressFamily int    `xml:"AddressFamily,omitempty"`
	RTType        int    `xml:"RTType,omitempty"`
	RTEntry       string `xml:"RTEntry,omitempty"`
}
