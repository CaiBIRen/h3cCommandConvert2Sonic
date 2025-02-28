package h3cmodel

import "encoding/xml"

type Context struct {
	XMLName             xml.Name           `xml:"Context"`
	ContextInformations ContextInformation `xml:"ContextInformations>ContextInformation"`
}

// ContextInformation 表示单个上下文信息
type ContextInformation struct {
	XMLName   xml.Name `xml:"ContextInformation"`
	ContextID string   `xml:"ContextID"`
	Name      string   `xml:"Name"`
}