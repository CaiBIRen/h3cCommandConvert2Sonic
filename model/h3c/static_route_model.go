package h3cmodel

import "encoding/xml"

// StaticRoute represents the top-level structure for static route configuration.
type StaticRoute struct {
	XMLName                       xml.Name                      `xml:"StaticRoute"`
	Ipv4StaticRouteConfigurations Ipv4StaticRouteConfigurations `xml:"Ipv4StaticRouteConfigurations"`
	Ipv6StaticRouteConfigurations Ipv6StaticRouteConfigurations `xml:"Ipv6StaticRouteConfigurations"`
}

// Ipv4StaticRouteConfigurations contains a list of IPv4 static route entries.
type Ipv4StaticRouteConfigurations struct {
	RouteEntries []RouteEntry `xml:"RouteEntry"`
}

type Ipv6StaticRouteConfigurations struct {
	IPv6RouteEntries []IPv6RouteEntry `xml:"RouteEntry"`
}

// RouteEntry defines the structure of each IPv4 static route entry.
type RouteEntry struct {
	DestVrfIndex       string `xml:"DestVrfIndex"`
	DestTopologyIndex  string `xml:"DestTopologyIndex"`
	Ipv4Address        string `xml:"Ipv4Address"`
	Ipv4PrefixLength   string `xml:"Ipv4PrefixLength"`
	NexthopVrfIndex    string `xml:"NexthopVrfIndex"`
	NexthopIpv4Address string `xml:"NexthopIpv4Address"`
	IfIndex            string `xml:"IfIndex"`
	Tag                string `xml:"Tag"`
	Preference         string `xml:"Preference"`
	Description        string `xml:"Description"`
}

type IPv6RouteEntry struct {
	DestVrfIndex       string `xml:"DestVrfIndex"`
	DestTopologyIndex  string `xml:"DestTopologyIndex"`
	Ipv6Address        string `xml:"Ipv6Address"`
	Ipv6PrefixLength   string `xml:"Ipv6PrefixLength"`
	NexthopVrfIndex    string `xml:"NexthopVrfIndex"`
	NexthopIpv6Address string `xml:"NexthopIpv6Address"`
	IfIndex            string `xml:"IfIndex"`
	Tag                string `xml:"Tag"`
	Preference         string `xml:"Preference"`
	Description        string `xml:"Description"`
}
