package h3cmodel

import "encoding/xml"

type RoutePolicy struct {
	XMLName        xml.Name       `xml:"RoutePolicy"`
	Policy         Policy         `xml:"Policy"`
	IPv4PrefixList IPv4PrefixList `xml:"IPv4PrefixList"`
	IPv6PrefixList IPv6PrefixList `xml:"IPv6PrefixList"`
}

// Policy contains entries that define routing policies.
type Policy struct {
	Entry []Entry `xml:"Entry"`
}

// Entry defines a single entry in the policy.
type Entry struct {
	PolicyName       string           `xml:"PolicyName"`
	Index            int              `xml:"Index"`
	Mode             string           `xml:"Mode"`
	Match            Match            `xml:"Match"`
	Apply            Apply            `xml:"Apply"`
	ApplyIpv4NextHop ApplyIpv4NextHop `xml:"ApplyIpv4NextHop"`
	MatchASPathList  MatchASPathList  `xml:"MatchASPathList"`
}

// Match contains matching conditions for a policy entry.
type Match struct {
	IPv4AddressPrefixList string `xml:"IPv4AddressPrefixList"`
	IPv6AddressPrefixList string `xml:"IPv6AddressPrefixList"`
	Tag                   int    `xml:"Tag"`
	RouteType             string `xml:"RouteType"`
}

// Apply contains actions to apply when a policy entry matches.
type Apply struct {
	LocalPreference int    `xml:"LocalPreference"`
	IPv6NextHop     string `xml:"IPv6NextHop"`
}

// ApplyIpv4NextHop specifies the IPv4 next hop to apply.
type ApplyIpv4NextHop struct {
	NextHopAddr string `xml:"NextHopAddr"`
}

// MatchASPathList contains AS path list matching conditions.
type MatchASPathList struct {
	AspathList string `xml:"AspathList"`
}

// IPv4PrefixList contains a list of IPv4 prefix definitions.
type IPv4PrefixList struct {
	PrefixList []PrefixList `xml:"PrefixList"`
}

// IPv6PrefixList contains a list of IPv6 prefix definitions.
type IPv6PrefixList struct {
	PrefixList []PrefixList `xml:"PrefixList"`
}

// PrefixList defines an individual prefix list entry.
type PrefixList struct {
	PrefixListName   string `xml:"PrefixListName"`
	Index            int    `xml:"Index"`
	Mode             string `xml:"Mode"`
	Ipv4Address      string `xml:"Ipv4Address,omitempty"`
	Ipv6Address      string `xml:"Ipv6Address,omitempty"`
	Ipv4PrefixLength string `xml:"Ipv4PrefixLength,omitempty"`
	Ipv6PrefixLength string `xml:"Ipv6PrefixLength,omitempty"`
	MinPrefixLength  string `xml:"MinPrefixLength"`
	MaxPrefixLength  string `xml:"MaxPrefixLength"`
}
