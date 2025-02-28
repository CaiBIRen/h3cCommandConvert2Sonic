package h3cmodel

import "encoding/xml"

type L2vpn struct {
	XMLName       xml.Name           `xml:"L2VPN"`
	VSIs          L2vpnVSIS          `xml:"VSIs,omitempty"`
	VSIInterfaces L2vpnVSIInterfaces `xml:"VSIInterfaces,omitempty"`
}

type L2vpnVSIS struct {
	L2vpnVSIS []VSI `xml:"VSI,omitempty"`
}

type L2vpnVSIInterfaces struct {
	L2vpnVSIInterfaces []VSIInterface `xml:"Interface,omitempty"`
}

type VSI struct {
	VsiName        string `xml:"VsiName,omitempty"`
	ArpSuppression bool   `xml:"ArpSuppression,omitempty"`
	NdSuppression  bool   `xml:"NdSuppression,omitempty"`
	MacLearning    bool   `xml:"MacLearning,omitempty"`
	Flooding       bool   `xml:"Flooding,omitempty"`
	FloodType      int    `xml:"FloodType,omitempty"`
	VsiInterfaceID int    `xml:"VsiInterfaceID,omitempty"`
	Statistics     bool   `xml:"Statistics,omitempty"`
	Description    string `xml:"Description,omitempty"`
	AdminStatus    string `xml:"AdminStatus,omitempty"`
}
type VSIInterface struct {
	ID          int  `xml:"ID,omitempty"`
	LocalEnable bool `xml:"LocalEnable,omitempty"`
	L3VNI       int  `xml:"L3VNI,omitempty"`
}
