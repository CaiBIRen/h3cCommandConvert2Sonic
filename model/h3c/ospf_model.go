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

// 定义Interfaces元素的结构体
type OSPF_Interfaces struct {
	XMLName   xml.Name    `xml:"Interfaces"`
	Interface []OSPF_Interface `xml:"Interface,omitempty"`
}

// 定义OSPF根元素的结构体
type OSPF struct {
	XMLName    xml.Name   `xml:"OSPF"`
	Interfaces OSPF_Interfaces `xml:"Interfaces,omitempty"`
}
