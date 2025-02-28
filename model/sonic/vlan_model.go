package sonicmodel

type Vlanroot struct {
	SonicVLAN VLANData `json:"sonic-vlan:sonic-vlan"`
}

type VLANData struct {
	VLAN       VLAN       `json:"VLAN,omitempty"`
	VLANMember VLANMember `json:"VLAN_MEMBER,omitempty"`
}

type VLAN struct {
	VLANList []VLANNode `json:"VLAN_LIST"`
}

type VLANNode struct {
	Name                   string   `json:"name"`
	VLANID                 int      `json:"vlanid"`
	MTU                    int      `json:"mtu,omitempty"`
	AdminStatus            string   `json:"admin_status,omitempty"`
	Description            string   `json:"description,omitempty"`
}

type VLANMember struct {
	VLANMemberList []VLANMemberList `json:"VLAN_MEMBER_LIST"`
}

type VLANMemberList struct {
	Name        string `json:"name,omitempty"`
	IfName      string `json:"ifname,omitempty"`
	TaggingMode string `json:"tagging_mode,omitempty"`
}
