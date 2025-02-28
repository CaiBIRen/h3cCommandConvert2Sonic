package h3cmodel

import (
	"encoding/xml"
)

// Device 代表整个设备的信息
type Device struct {
	XMLName          xml.Name         `xml:"Device"`
	PhysicalEntities PhysicalEntities `xml:"PhysicalEntities,omitempty"`
	Base             Base             `xml:"Base,omitempty"`
}

// PhysicalEntities 代表设备中的物理实体集合
type PhysicalEntities struct {
	XMLName xml.Name `xml:"PhysicalEntities"`
	Entity  Entity   `xml:"Entity,omitempty"`
}

// Base 代表设备的基本信息
type Base struct {
	XMLName         xml.Name `xml:"Base"`
	HostName        string   `xml:"HostName,omitempty"`
	HostDescription string   `xml:"HostDescription,omitempty"`
}

// Entity 代表设备中的单个物理实体
type Entity struct {
	XMLName     xml.Name `xml:"Entity"`
	SoftwareRev string   `xml:"SoftwareRev,omitempty"`
	Model       string   `xml:"Model,omitempty"`
}

type LLDP struct {
	XMLName       xml.Name      `xml:"LLDP"`
	LLDPNeighbors LLDPNeighbors `xml:"LLDPNeighbors,omitempty"`
}

type LLDPNeighbors struct {
	XMLName      xml.Name       `xml:"LLDPNeighbors"`
	LLDPNeighbor []LLDPNeighbor `xml:"LLDPNeighbor,omitempty"`
}

type LLDPNeighbor struct {
	XMLName    xml.Name `xml:"LLDPNeighbor"`
	IfIndex    string   `xml:"IfIndex,omitempty" json:"LocalInterface"`
	SystemName string   `xml:"SystemName,omitempty" json:"SystemName"`
	ChassisId  string   `xml:"ChassisId,omitempty" json:"ChasssID"`
	PortId     string   `xml:"PortId,omitempty" json:"PortID"`
}

type LAGG struct {
	XMLName     xml.Name      `xml:"LAGG"`
	Base        *LAGG_Base    `xml:"Base,omitempty"`
	LAGGGroups  *LAGG_Groups  `xml:"LAGGGroups,omitempty"`
	LAGGMembers *LAGG_Members `xml:"LAGGMembers,omitempty"`
}

type LAGG_Base struct {
	SystemID string `xml:"SystemID,omitempty"`
}

type LAGG_Groups struct {
	LAGGGroup []LAGG_Group `xml:"LAGGGroup"`
}

type LAGG_Group struct {
	GroupId string `xml:"GroupId,omitempty"`
	IfIndex string `xml:"IfIndex,omitempty"`
}

type LAGG_Members struct {
	LAGGMember []LAGG_Member `xml:"LAGGMember,omitempty"`
}

type LAGG_Member struct {
	GroupId string `xml:"GroupId,omitempty"`
	IfIndex string `xml:"IfIndex,omitempty"`
}

