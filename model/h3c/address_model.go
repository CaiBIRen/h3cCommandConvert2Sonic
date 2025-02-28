package h3cmodel

type IPV4ADDRESS struct {
	Ipv4Addresses Ipv4Addresses `xml:"Ipv4Addresses"`
}

type Ipv4Addresses struct {
	Ipv4Address []Ipv4Address `xml:"Ipv4Address"`
}

type Ipv4Address struct {
	IfIndex       string `xml:"IfIndex"`
	Ipv4Address   string `xml:"Ipv4Address"`
	Ipv4Mask      string `xml:"Ipv4Mask"`
	AddressOrigin int    `xml:"AddressOrigin"`
}

type IPV6ADDRESS struct {
	Ipv6AddressesConfig Ipv6AddressesConfig `xml:"Ipv6AddressesConfig"`
}

type Ipv6AddressesConfig struct {
	AddressEntry []AddressEntry `xml:"AddressEntry"`
}

type AddressEntry struct {
	IfIndex          string `xml:"IfIndex"`
	Ipv6Address      string `xml:"Ipv6Address"`
	AddressOrigin    int    `xml:"AddressOrigin"`
	Ipv6PrefixLength string `xml:"Ipv6PrefixLength"`
	AnycastFlag      bool   `xml:"AnycastFlag"`
}
