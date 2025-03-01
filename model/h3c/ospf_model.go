package h3cmodel

import "encoding/xml"

// 定义Interface元素的结构体
type OSPF_Interface struct {
	IfIndex     string   `xml:"IfIndex,omitempty"`
	IfEnable    IfEnable `xml:"IfEnable,omitempty"`
	NetworkType int      `xml:"NetworkType,omitempty"`
}

// 定义IfEnable元素的结构体
type IfEnable struct {
	Name          string `xml:"Name,omitempty"`
	AreaId        string `xml:"AreaId,omitempty"`
	ExcludedSubIp bool   `xml:"ExcludedSubIp,omitempty"`
}

type OSPF_Instance struct {
	Name                  string `xml:"Name,omitempty"`
	VRF                   string `xml:"VRF,omitempty"`
	RouterId              string `xml:"RouterId,omitempty"`
	VpnInstanceCapability int    `xml:"VpnInstanceCapability,omitempty"`
}

type OSPF_Area struct {
	Name   string `xml:"Name,omitempty"`
	AreaId string `xml:"AreaId,omitempty"`
}

type OSPF_Redist struct {
	Name     string `xml:"Name,omitempty"`
	TopoId   int    `xml:"TopoId,omitempty"`
	Protocol int    `xml:"Protocol,omitempty"`
}

// 定义Interfaces元素的结构体
type OSPF_Interfaces struct {
	XMLName   xml.Name         `xml:"Interfaces"`
	Interface []OSPF_Interface `xml:"Interface,omitempty"`
}

type OSPF_Instances struct {
	XMLName  xml.Name        `xml:"Instances"`
	Instance []OSPF_Instance `xml:"Instance,omitempty"`
}

type OSPF_Areas struct {
	XMLName xml.Name    `xml:"Areas"`
	Area    []OSPF_Area `xml:"Area,omitempty"`
}

type OSPF_Redistributes struct {
	XMLName xml.Name      `xml:"Redistributes"`
	Redist  []OSPF_Redist `xml:"Redist,omitempty"`
}

// 定义OSPF根元素的结构体
type OSPF struct {
	XMLName       xml.Name           `xml:"OSPF"`
	Interfaces    OSPF_Interfaces    `xml:"Interfaces,omitempty"`
	Instances     OSPF_Instances     `xml:"Instances,omitempty"`
	Areas         OSPF_Areas         `xml:"Areas,omitempty"`
	Redistributes OSPF_Redistributes `xml:"Redistributes,omitempty"`
}
