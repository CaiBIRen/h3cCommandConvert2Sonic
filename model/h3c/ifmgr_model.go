package h3cmodel

import "encoding/xml"

type Ifmgr struct {
	XMLName         xml.Name         `xml:"Ifmgr"`
	Interfaces      *Interfaces      `xml:"Interfaces"`
	LogicInterfaces *LogicInterfaces `xml:"LogicInterfaces"`
	SubInterfaces   *SubInterfaces   `xml:"SubInterfaces"`
}

// Interfaces contains a list of Interface elements.
type Interfaces struct {
	Interface []Interface `xml:"Interface"`
}

// Interface represents an individual network interface.
type Interface struct {
	IfIndex             string `xml:"IfIndex"`
	Name                string `xml:"Name,omitempty"`
	IfTypeExt           string `xml:"ifTypeExt,omitempty"`
	Description         string `xml:"Description"`
	OperStatus          string `xml:"OperStatus,omitempty"`
	AdminStatus         string `xml:"AdminStatus"`
	ActualSpeed         string `xml:"ActualSpeed,omitempty"`
	LinkType            string `xml:"LinkType"`
	PVID                string `xml:"PVID"`
	MAC                 string `xml:"MAC"`
	InetAddressIPV4     string `xml:"InetAddressIPV4,omitempty"`
	InetAddressIPV4Mask string `xml:"InetAddressIPV4Mask,omitempty"`
	ConfigMTU           string `xml:"ConfigMTU"`
}

type LogicInterfaces struct {
	Interface Interface_logical `xml:"Interface"`
}

type SubInterfaces struct {
	Interface Interface_sub `xml:"Interface"`
}
type Interface_sub struct {
	IfIndex string  `xml:"IfIndex"`
	SubNum  string  `xml:"SubNum"`
	Remove  *string `xml:"Remove"`
}

type Interface_logical struct {
	IfTypeExt string  `xml:"IfTypeExt"`
	Number    string  `xml:"Number"`
	Remove    *string `xml:"Remove"`
}
