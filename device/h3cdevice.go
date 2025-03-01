package device

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"regexp"
	"sonic-unis-framework/basic"
	"sonic-unis-framework/configuration"
	h3cmodel "sonic-unis-framework/model/h3c"
	sonicmodel "sonic-unis-framework/model/sonic"
	sonichandlers "sonic-unis-framework/sonichandlers"
	"sonic-unis-framework/tcontext"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
)

type H3cdevice struct {
}

const (
	IF_L2GE_TYPE = "19"
	IF_LAG_TYPE  = "56"
)

func CreateFeaturemap(configmap map[string]map[string]interface{}, str ...string) {
	for _, v := range str {
		if _, ok := configmap[v]; !ok {
			configmap[v] = make(map[string]interface{})
		}
	}
}

// For scenarios where indexes need to be queried, such as when there is a list type in the attribute, data needs to be added to the list incrementally
func IndexQueryContext(configmap map[string]map[string]interface{}, key string, childkey string) bool {
	node := configmap[key]
	if _, ok := node[childkey]; !ok {
		return false
	} else {
		return true
	}
}

func Parameters2Index(parameters ...string) string {
	var res string
	for _, param := range parameters {
		res += param + "@"
	}
	return strings.TrimSuffix(res, "@")
}

func (h3c H3cdevice) Role() string {
	configuration.ServiceConfiguration.Configmux.RLock()
	defer configuration.ServiceConfiguration.Configmux.RUnlock()
	return configuration.ServiceConfiguration.Role
}

// xml to go  struct
func (h3c H3cdevice) Decode(featuremap map[string]*xmlquery.Node, c *tcontext.Tcontext) error {
	for k, v := range featuremap {
		switch k {
		case "L3vpn":
			var l3vpn h3cmodel.L3vpn
			err := xml.Unmarshal([]byte(v.OutputXML(true)), &l3vpn)
			if err != nil {
				glog.Errorf("L3vpn xml illegal %v", err)
				return err
			}

			c.Features[k] = l3vpn
		case "L2VPN":
			var l2vpn h3cmodel.L2vpn
			err := xml.Unmarshal([]byte(v.OutputXML(true)), &l2vpn)
			if err != nil {
				glog.Errorf("L2VPN xml illegal %v", err)
				return err
			}
			c.Features[k] = l2vpn
		case "BGP":
			var bgp h3cmodel.BGP
			err := xml.Unmarshal([]byte(v.OutputXML(true)), &bgp)
			if err != nil {
				glog.Errorf("BGP xml illegal %v", err)
				return err
			}
			c.Features[k] = bgp
		case "Device":
			var device h3cmodel.Device
			err := xml.Unmarshal([]byte(v.OutputXML(true)), &device)
			if err != nil {
				glog.Errorf("Device xml illegal %v", err)
				return err
			}
			c.Features[k] = device
		case "LLDP":
			var lldp h3cmodel.LLDP
			err := xml.Unmarshal([]byte(v.OutputXML(true)), &lldp)
			if err != nil {
				glog.Errorf("LLDP xml illegal %v", err)
				return err
			}
			c.Features[k] = lldp
		case "Ifmgr":
			//分析控制器代码merge操作 border 只涉及mac下发
			var ifmgr h3cmodel.Ifmgr
			err := xml.Unmarshal([]byte(v.OutputXML(true)), &ifmgr)
			if err != nil {
				glog.Errorf("ifmgr xml illegal %v", err)
				return err
			}
			c.Features[k] = ifmgr
		case "LAGG":
			var lagg h3cmodel.LAGG
			err := xml.Unmarshal([]byte(v.OutputXML(true)), &lagg)
			if err != nil {
				glog.Errorf("LAGG xml illegal %v", err)
				return err
			}
			c.Features[k] = lagg
		case "StaticRoute":
			var staticroute h3cmodel.StaticRoute
			err := xml.Unmarshal([]byte(v.OutputXML(true)), &staticroute)
			if err != nil {
				glog.Errorf("StaticRoute xml illegal %v", err)
				return err
			}
			c.Features[k] = staticroute
		case "IPV4ADDRESS":
			var ipv4address h3cmodel.IPV4ADDRESS
			err := xml.Unmarshal([]byte(v.OutputXML(true)), &ipv4address)
			if err != nil {
				glog.Errorf("IPV4ADDRESS xml illegal %v", err)
				return err
			}
			c.Features[k] = ipv4address
		case "IPV6ADDRESS":
			var ipv6address h3cmodel.IPV6ADDRESS
			err := xml.Unmarshal([]byte(v.OutputXML(true)), &ipv6address)
			if err != nil {
				glog.Errorf("IPV6ADDRESS xml illegal %v", err)
				return err
			}
			c.Features[k] = ipv6address
		case "RoutePolicy":
			var routepolicy h3cmodel.RoutePolicy
			err := xml.Unmarshal([]byte(v.OutputXML(true)), &routepolicy)
			if err != nil {
				glog.Errorf("IPV6ADDRESS xml illegal %v", err)
				return err
			}
			c.Features[k] = routepolicy
		case "OSPF":
			var ospf h3cmodel.OSPF
			err := xml.Unmarshal([]byte(v.OutputXML(true)), &ospf)
			if err != nil {
				glog.Errorf("ospf xml illegal %v", err)
				return err
			}
			c.Features[k] = ospf
		}
	}
	// fmt.Println("_________%v", c.Features["L2VPN"].(h3cmodel.L2vpn))
	return nil
}

func (h3c H3cdevice) IntegrationReply(c *tcontext.Tcontext) (string, error) {
	if len(c.Features) == 0 {
		return "<data></data>", nil
	}
	var replyprefix string = "<data><top xmlns=\"http://www.h3c.com/netconf/data:1.0\">"
	var replysuffix string = "</top></data>"
	var dataxml string
	for k, v := range c.Features {
		switch k {
		case "Device":
			devicedata := v.(h3cmodel.Device)
			output, err := xml.MarshalIndent(devicedata, "", "  ")
			if err != nil {
				return "", err
			}
			dataxml += OutputLineBreak(output)
		case "LLDP":
			lldpdata := v.(h3cmodel.LLDP)
			if h3c.Role() == "Leaf" {
				configuration.ServiceConfiguration.Configmux.RLock()
				if len(configuration.ServiceConfiguration.Serverlldps) > 0 {
					lldpdata.LLDPNeighbors.LLDPNeighbor = append(lldpdata.LLDPNeighbors.LLDPNeighbor, configuration.ServiceConfiguration.Serverlldps...)
				}
				configuration.ServiceConfiguration.Configmux.RUnlock()
			}
			output, err := xml.MarshalIndent(lldpdata, "", "  ")
			if err != nil {
				return "", err
			}
			dataxml += OutputLineBreak(output)
			fmt.Println(dataxml)
		case "Ifmgr":
			ifmgrdata := v.(h3cmodel.Ifmgr)
			output, err := xml.MarshalIndent(ifmgrdata, "", "  ")
			if err != nil {
				return "", err
			}
			dataxml += OutputLineBreak(output)
		case "BGP":
			bgpasndata := v.(h3cmodel.BGP)
			output, err := xml.MarshalIndent(bgpasndata, "", "  ")
			if err != nil {
				return "", err
			}
			dataxml += OutputLineBreak(output)
		case "LAGG":
			laggdata := v.(h3cmodel.LAGG)
			output, err := xml.MarshalIndent(laggdata, "", "  ")
			if err != nil {
				return "", err
			}
			dataxml += OutputLineBreak(output)
		}
	}
	return replyprefix + dataxml + replysuffix, nil

}

func OutputLineBreak(output []byte) string {
	return string(output) + "\n"
}

// go struct to  sonic
func (h3c H3cdevice) EncodeMerge(c *tcontext.Tcontext) error {
	//元数据配置
	if _, ok := c.Features["L2VPN"]; ok {
		err := L2vpnEncodeMerge(c)
		if err != nil {
			return err
		}
	}

	if _, ok := c.Features["L3vpn"]; ok {
		err := L3vpnEncodeMerge(c)
		if err != nil {
			return err
		}
	}
	for k, _ := range c.Features {
		switch k {
		case "BGP":
			err := BGPEncodeMerge(c)
			if err != nil {
				return err
			}
		case "StaticRoute":
			err := StaticRouteEncodeMerge(c)
			if err != nil {
				return err
			}
		case "IPV4ADDRESS":
			err := IPV4ADDRESSEncodeMerge(c)
			if err != nil {
				return err
			}
		case "IPV6ADDRESS":
			err := IPV6ADDRESSEncodeMerge(c)
			if err != nil {
				return err
			}
		case "Ifmgr":
			err := IfmgrEncodeMerge(c)
			if err != nil {
				return err
			}
		case "RoutePolicy":
			err := RoutePolicyEncodeMerge(c)
			if err != nil {
				return err
			}
		case "OSPF":
			err := OSPFEncodeMerge(c)
			if err != nil {
				return err
			}
		}
	}
	c.DiscreteConfigurationIntegration()

	return nil
}

func (h3c H3cdevice) EncodeRemove(c *tcontext.Tcontext) error {

	for k, _ := range c.Features {
		switch k {
		case "L3vpn":
			err := L3vpnEncodeRemove(c)
			if err != nil {
				return err
			}
		case "L2vpn":
			err := L2vpnEncodeRemove(c)
			if err != nil {
				return err
			}

		case "BGP":
			err := BGPEncodeRemove(c)
			if err != nil {
				return err
			}
		case "StaticRoute":
			err := StaticRouteEncodeRemove(c)
			if err != nil {
				return err
			}
		case "RoutePolicy":
			err := RoutePolicyEncodeRemove(c)
			if err != nil {
				return err
			}
			// case "OSPF":
			// 	err := OSPFEncodeRemove(c)
			// 	if err != nil {
			// 		return err
			// 	}

		}
	}
	c.DiscreteConfigurationIntegration()
	return nil
}

func (h3c H3cdevice) EncodeGet(featuremap map[string]*xmlquery.Node, c *tcontext.Tcontext) error {
	for k, v := range featuremap {
		switch k {
		case "Device":
			err := DeviceEncodeGET(c)
			if err != nil {
				return err
			}
		case "Ifmgr":
			err := IfmgrEncodeGET(v, c)
			if err != nil {
				return err
			}
		case "LLDP":
			err := LLDPEncodeGET(c)
			if err != nil {
				return err
			}

		case "LAGG":
			err := LAGGEncodeGET(v, c)
			if err != nil {
				return err
			}
		//
		case "BGP":
			err := BGPEncodeGET(c)
			if err != nil {
				return err
			}
		}
	}
	glog.Info("EncondeGet context features", c.Features)
	return nil
}

func (h3c H3cdevice) EncodeAction(c *tcontext.Tcontext) error {
	for k, _ := range c.Features {
		switch k {
		case "Ifmgr":
			err := IfmgrEncodeAction(c)
			if err != nil {
				return err
			}

			// case "BGP":
			// 	err := BGPEncode_ACTION(c)
			// 	if err != nil {
			// 		return err
			// 	}
			// case "StaticRoute":
			// 	err := StaticRouteEncode_ACTION(c)
			// 	if err != nil {
			// 		return err
			// 	}
		}
	}
	return nil
}

func BGPEncodeGET(c *tcontext.Tcontext) error {
	err := sonichandlers.GetSONICBGPInstance(c)
	if err != nil {
		return err
	}
	if v, ok := c.SonicConfig[basic.SONICBGPKEY].(sonicmodel.BGPGlobalConfigASN); ok {
		var bgpnode h3cmodel.BGP
		var instancenode h3cmodel.Instance
		instancenode.ASNumber = strconv.Itoa(v.LocalASN)
		bgpnode.Instances.Items = append(bgpnode.Instances.Items, instancenode)
		c.Features["BGP"] = bgpnode
		return nil
	}
	glog.Error("BGP_EncodeGET assert failed")
	return errors.New("BGP_EncodeGET assert failed")
}

func LAGGEncodeGET(node *xmlquery.Node, c *tcontext.Tcontext) error {
	var featureFlag bool
	childnode := xmlquery.Find(node, "./*")
	var laggnode h3cmodel.LAGG
	for _, v := range childnode {
		switch v.Data {
		case "LAGGGroups":
			err := sonichandlers.GetSONICPortChannelList(c)
			if err != nil {
				return err
			}
			if v, ok := c.SonicConfig[basic.SONICPORTCHANNEL].(sonicmodel.PortChannelList); ok {
				var groups h3cmodel.LAGG_Groups

				for _, portchannel := range v.LAGTableList {
					var group h3cmodel.LAGG_Group
					re := regexp.MustCompile(`\d+`)
					group.GroupId = re.FindString(portchannel.Name)
					group.IfIndex = re.FindString(portchannel.Name)
					groups.LAGGGroup = append(groups.LAGGGroup, group)
				}
				if len(groups.LAGGGroup) > 0 {
					featureFlag = true
					laggnode.LAGGGroups = &groups
				}
			} else {
				glog.Error("LAGGGroups assert failed")
				return errors.New("LAGGGroups assert failed")
			}

		case "LAGGMembers":
			err := sonichandlers.GetSONICPortChannelMembers(c)
			if err != nil {
				return err
			}
			if v, ok := c.SonicConfig[basic.SONICPORTCHANNELMEMBERS].(sonicmodel.PortChannelMembers); ok {
				var members h3cmodel.LAGG_Members

				for _, portchannelmember := range v.PortChannelMemberList {
					var member h3cmodel.LAGG_Member
					re := regexp.MustCompile(`\d+`)
					member.GroupId = re.FindString(portchannelmember.Name)
					member.IfIndex = re.FindString(portchannelmember.Ifname)
					members.LAGGMember = append(members.LAGGMember, member)
				}
				if len(members.LAGGMember) > 0 {
					featureFlag = true
					laggnode.LAGGMembers = &members
				}
			} else {
				glog.Error("LAGGGroups assert failed")
				return errors.New("LAGGGroups assert failed")
			}
		case "Base":
			err := sonichandlers.GetSONICBridgeMAC(c)
			if err != nil {
				return err
			}
			if v, ok := c.SonicConfig[basic.SONICSYSTEMID].(sonicmodel.SonicDeviceMetadata); ok {
				featureFlag = true
				basenode := &h3cmodel.LAGG_Base{SystemID: strings.ReplaceAll(v.MAC, ":", "-")}
				laggnode.Base = basenode
			} else {
				glog.Error("LAGGGroups assert failed")
				return errors.New("LAGGGroups assert failed")
			}
		}

		if featureFlag {
			c.Features["LAGG"] = laggnode
		}
	}
	return nil
}

func DeviceEncodeGET(c *tcontext.Tcontext) error {
	err := sonichandlers.GetSONICDevice(c)
	if err != nil {
		return err
	}
	//Base + PhysicalEntities不需要做什么处理
	if v, ok := c.SonicConfig[basic.SONICDEVICE].(sonicmodel.Device); ok {
		var devicenode h3cmodel.Device
		devicenode.Base.HostName = v.Base.HostName
		devicenode.Base.HostDescription = v.Base.HostDescription
		devicenode.PhysicalEntities.Entity.SoftwareRev = v.PhysicalEntities.Entity.SoftwareRev
		c.Features["Device"] = devicenode
		return nil
	}
	glog.Error("DeviceEncodeGET assert failed")
	return errors.New("DeviceEncodeGET assert failed")
}

func LLDPEncodeGET(c *tcontext.Tcontext) error {
	err := sonichandlers.GetSONICLLDP(c)
	if err != nil {
		return err
	}
	if v, ok := c.SonicConfig[basic.SONICLLDP].(sonicmodel.OpenConfigLLDP); ok {
		var lldpnode h3cmodel.LLDP
		for _, iface := range v.Interface {
			for _, neighbor := range iface.Neighbors.Neighbor {
				idx, err := GetInterfaceString(neighbor.Id)
				if err != nil {
					return err
				}
				var h3cneighbor h3cmodel.LLDPNeighbor
				h3cneighbor.ChassisId = strings.ReplaceAll(neighbor.State.ChassisID, ":", "-")
				h3cneighbor.IfIndex = idx
				h3cneighbor.PortId = neighbor.State.PortID
				h3cneighbor.SystemName = neighbor.State.SystemName
				lldpnode.LLDPNeighbors.LLDPNeighbor = append(lldpnode.LLDPNeighbors.LLDPNeighbor, h3cneighbor)
			}
		}
		if len(lldpnode.LLDPNeighbors.LLDPNeighbor) > 0 {
			c.Features["LLDP"] = lldpnode
		}
		return nil
	}
	glog.Error("LLDPEncodeGET assert failed")
	return errors.New("LLDPEncodeGET assert failed")
}

func IfmgrEncodeGET(node *xmlquery.Node, c *tcontext.Tcontext) error {
	ifname := xmlquery.FindOne(node, "//Name")
	if ifname != nil && ifname.InnerText() != "" {
		Intname := ifname.InnerText()
		if strings.HasPrefix(Intname, "vlan") || strings.HasPrefix(Intname, "Vlan") {
			var ifmgrnode h3cmodel.Ifmgr
			var ifmgrinterfaces h3cmodel.Interfaces
			vlanid, _ := GetInterfaceString(Intname)
			err := sonichandlers.GetSONICVlanInterfaceByName("Vlan" + vlanid)
			//返回err代表查找不到
			if err != nil {
				return nil
			}
			ifinterface := h3cmodel.Interface{IfIndex: Intname, Name: Intname}
			ifmgrinterfaces.Interface = append(ifmgrinterfaces.Interface, ifinterface)
			ifmgrnode.Interfaces = &ifmgrinterfaces
			c.Features["Ifmgr"] = ifmgrnode
			return nil
		}
	}
	iftypenode := xmlquery.FindOne(node, "//ifTypeExt")
	innertext := iftypenode.InnerText()
	switch innertext {
	case IF_L2GE_TYPE:
		err := sonichandlers.GetSONICPort(c)
		if err != nil {
			return err
		}
		err = Ifmgr_IF_L2GE_TYPE(c)
		if err != nil {
			return err
		}
	case IF_LAG_TYPE:
		err := sonichandlers.GetSONICPortChannelList(c)
		if err != nil {
			return err
		}
		err = Ifmgr_IF_L3GE_TYPE(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func Ifmgr_IF_L3GE_TYPE(c *tcontext.Tcontext) error {
	innertext := IF_LAG_TYPE
	if v, ok := c.SonicConfig[basic.SONICPORTCHANNEL].(sonicmodel.PortChannelList); ok {
		var ifmgrnode h3cmodel.Ifmgr
		var ifmgrinterfaces h3cmodel.Interfaces
		for _, iface := range v.LAGTableList {
			var interfacenode h3cmodel.Interface
			var OperStatus string
			interfacenode.Name = iface.Name
			if iface.OperStatus == "up" {
				OperStatus = "1"
			} else {
				OperStatus = "2"
			}
			if iface.MAC == "" {
				iface.MAC = "00-00-00-00-00-00"
			} else {
				iface.MAC = strings.ReplaceAll(iface.MAC, ":", "-")
			}
			idx, err := GetInterfaceString(iface.Name)
			if err != nil {
				return err
			}
			interfacenode.IfIndex = idx
			interfacenode.OperStatus = OperStatus
			interfacenode.MAC = iface.MAC
			interfacenode.IfTypeExt = innertext
			ifmgrinterfaces.Interface = append(ifmgrinterfaces.Interface, interfacenode)
		}

		if len(ifmgrinterfaces.Interface) > 0 {
			ifmgrnode.Interfaces = &ifmgrinterfaces
			c.Features["Ifmgr"] = ifmgrnode
		}
		return nil
	}

	glog.Error("Ifmgr_IF_L3GE_TYPE assert failed")
	return errors.New("Ifmgr_IF_L3GE_TYPE assert failed")
}

func Ifmgr_IF_L2GE_TYPE(c *tcontext.Tcontext) error {
	innertext := IF_L2GE_TYPE
	if v, ok := c.SonicConfig[basic.SONICPORT].(sonicmodel.PortTable); ok {
		var ifmgrnode h3cmodel.Ifmgr
		var ifmgrinterfaces h3cmodel.Interfaces
		for _, iface := range v.PortTableList {
			if iface.Ifname == "PortConfigDone" || iface.Ifname == "PortInitDone" {
				continue
			}
			var interfacenode h3cmodel.Interface
			var OperStatus string
			if iface.OperStatus == "up" {
				OperStatus = "1"
			} else {
				OperStatus = "2"
			}
			if iface.MAC == "" {
				iface.MAC = "00-00-00-00-00-00"
			} else {
				iface.MAC = strings.ReplaceAll(iface.MAC, ":", "-")
			}
			idx, err := GetInterfaceString(iface.Ifname)
			if err != nil {
				return err
			}
			interfacenode.IfIndex = idx
			interfacenode.Name = iface.Ifname
			interfacenode.Description = iface.Description
			interfacenode.OperStatus = OperStatus
			interfacenode.MAC = iface.MAC
			interfacenode.IfTypeExt = innertext
			ifmgrinterfaces.Interface = append(ifmgrinterfaces.Interface, interfacenode)
		}

		if len(ifmgrinterfaces.Interface) > 0 {
			ifmgrnode.Interfaces = &ifmgrinterfaces
			c.Features["Ifmgr"] = ifmgrnode
		}
		return nil
	}

	glog.Error("Ifmgr_IF_L2GE_TYPE assert failed")
	return errors.New("get interface assert failed")
}

// action一般都是单独请求
// 41 -- vlan  16 -- loopback
func IfmgrEncodeAction(c *tcontext.Tcontext) error {
	glog.Info("enter ifmgr action encoding")
	Ifmgr := c.Features["Ifmgr"]
	data, ok := Ifmgr.(h3cmodel.Ifmgr)
	if !ok {
		glog.Error("ifmgr action data assert failed")
		return errors.New("ifmgr action data assert failed")
	}
	//主要涉及逻辑接口的创建
	if data.LogicInterfaces != nil {
		v := data.LogicInterfaces.Interface
		err := IfmgrLogicalCheck(&v)
		if err != nil {
			return err
		}
		if v.IfTypeExt == "41" {
			var vlanroot sonicmodel.Vlanroot
			var vlaninterfaceroot sonicmodel.VlanInterfaceroot
			vlanid, _ := strconv.Atoi(v.Number)
			//vlan
			vlannode := VlanListOrganize(vlanid, 1500)

			//vlan-interface
			vlaninterfacenode := sonicmodel.VlanInterface{
				VlanName: vlannode.Name,
			}
			vlanroot.SonicVLAN.VLAN.VLANList = append(vlanroot.SonicVLAN.VLAN.VLANList, vlannode)
			vlaninterfaceroot.SonicVLANInterface.VLAN_INTERFACE.VLAN_INTERFACE_LIST = append(
				vlaninterfaceroot.SonicVLANInterface.VLAN_INTERFACE.VLAN_INTERFACE_LIST, vlaninterfacenode)
			c.SonicConfig[basic.SONICVLANKEY] = vlanroot
			c.SonicConfig[basic.SONICVLANINTERFACEKEY] = vlaninterfaceroot
			// fmt.Println(c.SonicConfig)
			if v.Remove != nil {
				c.Operation = basic.OPERREMOVE
				if err := sonichandlers.RemoveSONICVlanInterface(c); err != nil {
					return err
				}
			} else {
				if err = sonichandlers.ConfigSONICVlan(c); err != nil {
					return err
				}
				if err = sonichandlers.ConfigSONICVlanInterface(c); err != nil {
					return err
				}
			}
		}
		if v.IfTypeExt == "16" {
			var loopbackroot sonicmodel.LoopbackInterfacesroot
			//loopback-interface
			loopbackinterfacenode := sonicmodel.LoopbackInterface{
				LoIfName: "Loopback" + v.Number,
			}
			loopbackroot.LoopbackInterfaceList = append(loopbackroot.LoopbackInterfaceList, loopbackinterfacenode)
			c.SonicConfig[basic.SONICLOOPBACKKEY] = loopbackroot
			// fmt.Println(c.SonicConfig)
			if v.Remove != nil {
				c.Operation = basic.OPERREMOVE
				if err := sonichandlers.RemoveSONICLoopbackInterface(c); err != nil {
					return err
				}
			} else {
				if err = sonichandlers.ConfigSONICLoopbackInterface(c); err != nil {
					return err
				}
			}
		}

	}
	return nil
}

func OSPFNetworkTypeTrans(networktype int) (string, error) {
	switch networktype {
	case 1:
		return "BROADCAST_NETWORK", nil
	case 3:
		return "POINT_TO_POINT_NETWORK", nil
	}
	return "", errors.New("Unrecognized ospf  networktype")
}

func OSPFRedistProtocolTrans(protocol int) (string, error) {
	switch protocol {
	case 1:
		return "DIRECTLY_CONNECTED", nil
	case 2:
		return "STATIC", nil
	case 6:
		return "BGP", nil
	}
	return "", errors.New("Unrecognized protocol type")
}

func GetOSPFVrf(c *tcontext.Tcontext, Name string) (string, error) {
	routerindex := Name + basic.SONICOSPFINSTANCEELELMENT
	if ospfroutermap, ok := c.DiscreteConfiguration[basic.SONICOSPFKEY]; ok {
		if ospfrouternode, ok := ospfroutermap[routerindex]; ok {
			node := ospfrouternode.(sonicmodel.OSPFv2Router)
			return node.VrfName, nil
		}
	}
	return "", errors.New("The request does not contain the ospf instance configuration before configuring other tables")
}

func OSPFEncodeMerge(c *tcontext.Tcontext) error {
	glog.Info("enter ospfv2 merge encoding")
	ospf := c.Features["OSPF"]
	data, ok := ospf.(h3cmodel.OSPF)
	if !ok {
		glog.Error("OSPF data assert failed")
		return errors.New("OSPF data assert failed")
	}
	//vlan + vlaninterface + vxlanmapping
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICOSPFKEY)
	if len(data.Instances.Instance) > 0 {
		for _, v := range data.Instances.Instance {
			v.VRF = VrfNameFormat(v.VRF)
			var ospfinstance sonicmodel.OSPFv2Router
			ospfinstance.RouterID = v.RouterId
			ospfinstance.VrfName = v.VRF
			ospfinstance.Enable = true
			ospfinstance.Description = "OSPF_Name" + v.Name
			//TODO:VpnInstanceCapability 无对应配置
			ospfinstanceindex := Parameters2Index(v.Name) + basic.SONICOSPFINSTANCEELELMENT
			c.DiscreteConfiguration[basic.SONICOSPFKEY][ospfinstanceindex] = ospfinstance
		}
	}
	if len(data.Areas.Area) > 0 {
		for _, v := range data.Areas.Area {
			var ospfarea sonicmodel.OSPFv2RouterArea
			vrfname, err := GetOSPFVrf(c, v.Name)
			if err != nil {
				return err
			}
			ospfarea.VrfName = vrfname
			ospfarea.AreaID = v.AreaId
			ospfarea.Description = "OSPF_Name" + v.Name
			ospfarea.Enable = true
			ospfareaindex := Parameters2Index(v.Name, v.AreaId) + basic.SONICOSPFAREAELELMENT
			c.DiscreteConfiguration[basic.SONICOSPFKEY][ospfareaindex] = ospfarea
		}
	}
	if len(data.Interfaces.Interface) > 0 {
		for _, v := range data.Interfaces.Interface {
			if strings.Contains(v.IfIndex, "Vlan") || strings.Contains(v.IfIndex, "vlan") {
				if v.NetworkType <= 0 {
					return errors.New("unkown ospf network-type")
				}
				networktype, err := OSPFNetworkTypeTrans(v.NetworkType)
				if err != nil {
					return errors.New("unkown ospf network-type")
				}
				var ospfinterfacenode sonicmodel.OSPFv2Interface
				vlanid, _ := GetInterfaceString(v.IfIndex)
				ospfinterfacenode.Name = "Vlan" + vlanid
				ospfinterfacenode.AreaID = v.IfEnable.AreaId
				ospfinterfacenode.Enable = true
				ospfinterfacenode.NetworkType = networktype
				ipprefix, err := sonichandlers.GetSONICVlanInterfaceIPByName("Vlan900")
				if err != nil {
					glog.Errorf("vlan %s interface ip not config,can not config ospf", vlanid)
					return errors.New("the interface has not been configured with an IP")
				}
				ospfinterfacenode.Address = strings.Split(ipprefix, "/")[0] //获取到的IP都是cide形式不用考虑异常
				ospfinterfaceindex := Parameters2Index(vlanid, ospfinterfacenode.Address) + basic.SONICOSPFINTERFACEELELMENT
				c.DiscreteConfiguration[basic.SONICOSPFKEY][ospfinterfaceindex] = ospfinterfacenode
			}
		}
	}
	if len(data.Redistributes.Redist) > 0 {
		for _, v := range data.Redistributes.Redist {
			var ospfredistributenode sonicmodel.OSPFv2RouterDistributeRoute
			vrfname, err := GetOSPFVrf(c, v.Name)
			if err != nil {
				return err
			}
			protocol, err := OSPFRedistProtocolTrans(v.Protocol)
			if err != nil {
				return err
			}
			ospfredistributenode.VrfName = vrfname
			ospfredistributenode.TableID = v.TopoId
			ospfredistributenode.Direction = "IMPORT"
			ospfredistributenode.Protocol = protocol
			ospfredistributeindex := Parameters2Index(v.Name, protocol) + basic.SONICOSPFREDISTELELMENT
			c.DiscreteConfiguration[basic.SONICOSPFKEY][ospfredistributeindex] = ospfredistributenode
		}
	}
	return nil
}

func L2vpnEncodeMerge(c *tcontext.Tcontext) error {
	glog.Info("enter l2vpn merge encoding")
	L2vpn := c.Features["L2VPN"]
	data, ok := L2vpn.(h3cmodel.L2vpn)
	if !ok {
		glog.Error("L2vpn data assert failed")
		return errors.New("L2vpn data assert failed")
	}
	//vlan + vlaninterface + vxlanmapping
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICVLANKEY, basic.SONICVXLANKEY, basic.SONICVLANINTERFACEKEY)
	if len(data.VSIInterfaces.L2vpnVSIInterfaces) > 0 {
		for _, v := range data.VSIInterfaces.L2vpnVSIInterfaces {
			err := VSIInterfaceCheck(&v)
			if err != nil {
				return err
			}
			vlanid := L3vni2Vlan(v.L3VNI)
			//vlan
			vlannode := VlanListOrganize(vlanid, 2000)
			vlannode.Description = "L3VNI_" + strconv.Itoa(v.L3VNI) + "_MAPPING"
			vlanindex := Parameters2Index(vlannode.Name) + basic.SONICVLANELEMENT
			c.DiscreteConfiguration[basic.SONICVLANKEY][vlanindex] = vlannode
			//vxlan
			vxlannode := VxlanTunnelMapOrganize(vlanid, v.L3VNI)
			vxlanindex := Parameters2Index(vxlannode.Name, vxlannode.Mapname) + basic.SONICVXLANTUNNELMAPELEMENT
			c.DiscreteConfiguration[basic.SONICVXLANKEY][vxlanindex] = vxlannode
			//vlan-interface
			vlaninterfacenode := sonicmodel.VlanInterface{
				VlanName: vlannode.Name,
			}
			vlaninterfaceindex := Parameters2Index(vlaninterfacenode.VlanName) + basic.SONICVLANINTERFACEELEMENT
			c.DiscreteConfiguration[basic.SONICVLANINTERFACEKEY][vlaninterfaceindex] = vlaninterfacenode
		}
	}
	return nil
}

func L2vpnEncodeRemove(c *tcontext.Tcontext) error {
	glog.Info("enter l2vpn remove encoding")
	L2vpn := c.Features["L2VPN"]
	data, ok := L2vpn.(h3cmodel.L2vpn)
	if !ok {
		glog.Error("L2vpn data assert failed")
		return errors.New("L2vpn data assert failed")
	}
	//vlan + vlaninterface + vxlanmapping
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICVLANKEY, basic.SONICVXLANKEY, basic.SONICVLANINTERFACEKEY)
	if len(data.VSIInterfaces.L2vpnVSIInterfaces) > 0 {
		for _, v := range data.VSIInterfaces.L2vpnVSIInterfaces {
			err := VSIInterfaceCheck(&v)
			if err != nil {
				return err
			}
			vlanid := L3vni2Vlan(v.L3VNI)
			//vlan
			vlanName := "Vlan" + strconv.Itoa(vlanid)
			vlanindex := Parameters2Index(vlanName) + basic.SONICVLANELEMENT
			c.DiscreteConfiguration[basic.SONICVLANKEY][vlanindex] = sonicmodel.VLANNode{Name: vlanName}
			//vxlan
			vxlanName := "map_" + strconv.Itoa(v.L3VNI) + "_Vlan"
			vxlanindex := Parameters2Index(basic.TUNNELNAME, vxlanName) + basic.SONICVXLANTUNNELMAPELEMENT
			c.DiscreteConfiguration[basic.SONICVXLANKEY][vxlanindex] = sonicmodel.VxlanTunnelMap{Name: basic.TUNNELNAME, Mapname: vxlanName}
			//vlan-interface
			vlaninterfacenode := sonicmodel.VlanInterface{
				VlanName: vlanName,
			}
			vlaninterfaceindex := Parameters2Index(vlaninterfacenode.VlanName) + basic.SONICVLANINTERFACEELEMENT
			c.DiscreteConfiguration[basic.SONICVLANINTERFACEKEY][vlaninterfaceindex] = vlaninterfacenode
		}
	}
	return nil
}

func L3vpnEncodeMerge(c *tcontext.Tcontext) error {
	L3vpn := c.Features["L3vpn"]
	data, ok := L3vpn.(h3cmodel.L3vpn)
	if !ok {
		return errors.New("L3vpn data assert failed")
	}
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICVRFKEY, basic.SONICVLANINTERFACEKEY, basic.SONICBGPKEY)
	if len(data.L3vpnVRF.VRFs) > 0 {
		for _, v := range data.L3vpnVRF.VRFs {
			v.VRF = VrfNameFormat(v.VRF)
			err := VRFCheck(&v)
			if err != nil {
				return err
			}
			vrfnode := VrfOrganize(v.VRF, 0)
			vrfindex := Parameters2Index(vrfnode.VrfName) + basic.SONICVRFELEMENT
			c.DiscreteConfiguration[basic.SONICVRFKEY][vrfindex] = vrfnode
			if v.RD != "" {
				var bgpafnode sonicmodel.BgpGlobalsAFList
				var bgpglobalnode sonicmodel.BgpGlobalsList

				bgpglobalindex := Parameters2Index(v.VRF) + basic.SONICBGPGLOBALELEMENT
				bgpglobalnode.VrfName = v.VRF
				c.DiscreteConfiguration[basic.SONICBGPKEY][bgpglobalindex] = bgpglobalnode

				bgpafindex := Parameters2Index(v.VRF, "l2vpn_evpn") + basic.SONICBGPGLOBALAFELEMENT
				if IndexQueryContext(c.DiscreteConfiguration, basic.SONICBGPKEY, bgpafindex) {
					bgpafnode = c.DiscreteConfiguration[basic.SONICBGPKEY][bgpafindex].(sonicmodel.BgpGlobalsAFList)
				} else {
					bgpafnode = BgpGlobalsAfOrganize(v.VRF, 4, "L3vpn")
				}
				bgpafnode.RouteDistinguisher = v.RD
				c.DiscreteConfiguration[basic.SONICBGPKEY][bgpafindex] = bgpafnode
			}
		}
	}

	if len(data.L3vpnIf.Binds) > 0 {
		//分两种情况1、l3vsi接口 找vlaninterface bind 2、实际interface bind
		for _, v := range data.L3vpnIf.Binds {
			v.VRF = VrfNameFormat(v.VRF)
			err := BindCheck(&v)
			if err != nil {
				return err
			}

			if strings.HasPrefix(v.IfIndex, "Vsi") || strings.HasPrefix(v.IfIndex, "vsi") {
				ifidx, err := GetInterfaceString(v.IfIndex)
				if err != nil {
					return err
				}
				l3vni, err := strconv.Atoi(ifidx)
				if err != nil {
					return err
				}
				var vrfnode sonicmodel.Vrf
				vrfindex := Parameters2Index(v.VRF) + basic.SONICVRFELEMENT
				vrfnode.VrfName = v.VRF
				vrfnode.Vni = l3vni
				c.DiscreteConfiguration[basic.SONICVRFKEY][vrfindex] = vrfnode

				vlan := L3vni2Vlan(l3vni)
				vlan_interface_node := sonicmodel.VlanInterface{VlanName: "Vlan" + strconv.Itoa(vlan), VrfName: v.VRF}
				vlaninterfaceindex := Parameters2Index(vlan_interface_node.VlanName) + basic.SONICVLANINTERFACEELEMENT
				c.DiscreteConfiguration[basic.SONICVLANINTERFACEKEY][vlaninterfaceindex] = vlan_interface_node
			} else if strings.Contains(v.IfIndex, "Vlan") || strings.Contains(v.IfIndex, "vlan") {
				ifidx, err := GetInterfaceString(v.IfIndex)
				if err != nil {
					return err
				}
				vlan_interface_node := sonicmodel.VlanInterface{VlanName: "Vlan" + ifidx, VrfName: v.VRF}
				vlaninterfaceindex := Parameters2Index(vlan_interface_node.VlanName) + basic.SONICVLANINTERFACEELEMENT
				c.DiscreteConfiguration[basic.SONICVLANINTERFACEKEY][vlaninterfaceindex] = vlan_interface_node
			} else if strings.Contains(v.IfIndex, "Loopback") || strings.Contains(v.IfIndex, "loopback") {
				ifidx, err := GetInterfaceString(v.IfIndex)
				if err != nil {
					return err
				}
				loopback_interface_node := sonicmodel.LoopbackInterface{LoIfName: "Loopback" + ifidx, VrfName: v.VRF}
				loopbackinterfaceindex := Parameters2Index(loopback_interface_node.LoIfName) + basic.SONICLOOPBACKINTERFACEELEMENT
				c.DiscreteConfiguration[basic.SONICLOOPBACKKEY][loopbackinterfaceindex] = loopback_interface_node
			}
		}
	}

	if len(data.L3vpnRT.RTs) > 0 {
		for _, v := range data.L3vpnRT.RTs {
			v.VRF = VrfNameFormat(v.VRF)
			err := L3vpnRTCheck(&v)
			if err != nil {
				return err
			}
			var bgpafnode sonicmodel.BgpGlobalsAFList
			bgpafnode.VrfName = v.VRF
			addressfamily := Familytrans(v.AddressFamily, "L3vpn")
			if addressfamily == "" {
				return errors.New("L3vpnRT addressfamily value err")
			}
			// //RT not config in this afs,all in evpn
			// if addressfamily == "ipv4_unicast" || addressfamily == "ipv6_unicast" {
			// 	bgpafindex := Parameters2Index(v.VRF, addressfamily) + basic.SONICBGPGLOBALAFELEMENT
			// 	bgpafnode.AFISAFI = addressfamily
			// 	c.DiscreteConfiguration[basic.SONICBGPKEY][bgpafindex] = bgpafnode
			// 	continue
			// }

			//evpn config rt
			bgpafindex := Parameters2Index(v.VRF, addressfamily) + basic.SONICBGPGLOBALAFELEMENT
			if !IndexQueryContext(c.DiscreteConfiguration, basic.SONICBGPKEY, bgpafindex) {
				bgpafnode = BgpGlobalsAfOrganize(v.VRF, v.AddressFamily, "L3vpn")
			} else {
				bgpafnode = c.DiscreteConfiguration[basic.SONICBGPKEY][bgpafindex].(sonicmodel.BgpGlobalsAFList)
			}
			//华三evpn默认导入IPV4/IPV6
			if addressfamily == "l2vpn_evpn" {
				bgpafnode.AdvertiseIPv4Unicast = true
				bgpafnode.AdvertiseIPv6Unicast = true
			}
			switch v.RTType {
			case 1:
				bgpafnode.ImportRTS = append(bgpafnode.ImportRTS, v.RTEntry)
			case 2:
				bgpafnode.ExportRTS = append(bgpafnode.ExportRTS, v.RTEntry)
			}
			c.DiscreteConfiguration[basic.SONICBGPKEY][bgpafindex] = bgpafnode
		}
	}
	return nil
}

func L3vpnEncodeRemove(c *tcontext.Tcontext) error {
	L3vpn := c.Features["L3vpn"]
	data, ok := L3vpn.(h3cmodel.L3vpn)
	if !ok {
		return errors.New("L3vpn data assert failed")
	}
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICVRFKEY, basic.SONICVLANINTERFACEKEY, basic.SONICBGPKEY)
	if len(data.L3vpnVRF.VRFs) > 0 {
		for _, v := range data.L3vpnVRF.VRFs {
			v.VRF = VrfNameFormat(v.VRF)
			err := VRFCheck(&v)
			if err != nil {
				return err
			}
			//sonic删除vrf前需要将其他引用配置清理干净{主要是接口下的引用}
			vrfnode := sonicmodel.Vrf{VrfName: v.VRF}
			vrfindex := Parameters2Index(vrfnode.VrfName) + basic.SONICVRFELEMENT
			c.DiscreteConfiguration[basic.SONICVRFKEY][vrfindex] = vrfnode
		}
	}
	//not consider interface unbind vrf
	// if len(data.L3vpnIf.Binds) > 0 {

	// }

	if len(data.L3vpnRT.RTs) > 0 {
		for _, v := range data.L3vpnRT.RTs {
			v.VRF = VrfNameFormat(v.VRF)
			err := L3vpnRTCheck(&v)
			if err != nil {
				return err
			}
			var bgpafnode sonicmodel.BgpGlobalsAFList
			addressfamily := Familytrans(v.AddressFamily, "L3vpn")
			if addressfamily == "l2vpn_evpn" {
				bgpafindex := Parameters2Index(v.VRF, addressfamily) + basic.SONICBGPGLOBALAFELEMENT
				if !IndexQueryContext(c.DiscreteConfiguration, basic.SONICBGPKEY, bgpafindex) {
					bgpafnode = BgpGlobalsAfOrganize(v.VRF, v.AddressFamily, "L3vpn")
				} else {
					bgpafnode = c.DiscreteConfiguration[basic.SONICBGPKEY][bgpafindex].(sonicmodel.BgpGlobalsAFList)
				}
				switch v.RTType {
				case 1:
					bgpafnode.ImportRTS = append(bgpafnode.ImportRTS, v.RTEntry)
				case 2:
					bgpafnode.ExportRTS = append(bgpafnode.ExportRTS, v.RTEntry)
				}
				//list rts to be deleted
				c.DiscreteConfiguration[basic.SONICBGPKEY][bgpafindex] = bgpafnode
			}
		}
	}
	return nil
}

func BGPEncodeMerge(c *tcontext.Tcontext) error {
	BGP := c.Features["BGP"]
	data, ok := BGP.(h3cmodel.BGP)
	if !ok {
		return errors.New("BGP data translate failed")
	}
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICBGPKEY, basic.SONICROUTECOMMONKEY)
	if len(data.VRFs.BGPVRF) > 0 {
		for _, v := range data.VRFs.BGPVRF {
			v.VRF = VrfNameFormat(v.VRF)
			if v.VRF == "" {
				return errors.New("bgp vrf index missing")
			}
			bgpglobalindex := Parameters2Index(v.VRF) + basic.SONICBGPGLOBALELEMENT
			bgpglobalnode := sonicmodel.BgpGlobalsList{VrfName: v.VRF}
			c.DiscreteConfiguration[basic.SONICBGPKEY][bgpglobalindex] = bgpglobalnode
		}
	}

	if len(data.Familys.Family) > 0 {
		for _, v := range data.Familys.Family {
			v.VRF = VrfNameFormat(v.VRF)
			if v.Balance.MaxBalance == 0 {
				continue
			}
			var bgpafnode sonicmodel.BgpGlobalsAFList
			addressfamily := Familytrans(v.Type, "BGP")
			bgpafindex := Parameters2Index(v.VRF, addressfamily) + basic.SONICBGPGLOBALAFELEMENT
			if !IndexQueryContext(c.DiscreteConfiguration, basic.SONICBGPKEY, bgpafindex) {
				bgpafnode = BgpGlobalsAfOrganize(v.VRF, v.Type, "BGP")
			} else {
				bgpafnode = c.DiscreteConfiguration[basic.SONICBGPKEY][bgpafindex].(sonicmodel.BgpGlobalsAFList)
			}
			bgpafnode.MaxEBGPPaths = v.Balance.MaxBalance
			bgpafnode.MaxIBGPPaths = v.Balance.MaxBalance
			c.DiscreteConfiguration[basic.SONICBGPKEY][bgpafindex] = bgpafnode
		}
	}

	if len(data.Redistributes.Redist) > 0 {
		for _, v := range data.Redistributes.Redist {
			v.VRF = VrfNameFormat(v.VRF)
			err := BGPRedistCheck(&v)
			if err != nil {
				return err
			}
			protocol := BGPProtocoltrans(v.Protocol)
			bgpfamliy := Familytrans(v.Family, "REDISTRIBUTE")
			redistributeindex := Parameters2Index(v.VRF, bgpfamliy, protocol) + basic.SONICROUTECOMMONREDISTELEMENT
			if IndexQueryContext(c.DiscreteConfiguration, basic.SONICROUTECOMMONKEY, redistributeindex) {
				continue
			}
			redistributenode := sonicmodel.RouteRedistributenode{VrfName: v.VRF, AddrFamily: bgpfamliy, SrcProtocol: protocol, DstProtocol: "bgp"}
			c.DiscreteConfiguration[basic.SONICROUTECOMMONKEY][redistributeindex] = redistributenode
		}
	}

	return nil
}

func BGPEncodeRemove(c *tcontext.Tcontext) error {
	BGP := c.Features["BGP"]
	data, ok := BGP.(h3cmodel.BGP)
	if !ok {
		return errors.New("BGP data translate failed")
	}
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICBGPKEY, basic.SONICROUTECOMMONKEY)
	if len(data.VRFs.BGPVRF) > 0 {
		for _, v := range data.VRFs.BGPVRF {
			v.VRF = VrfNameFormat(v.VRF)
			if v.VRF == "" {
				return errors.New("bgp vrf index missing")
			}
			bgpglobalindex := Parameters2Index(v.VRF) + basic.SONICBGPGLOBALELEMENT
			bgpglobalnode := sonicmodel.BgpGlobalsList{VrfName: v.VRF}
			c.DiscreteConfiguration[basic.SONICBGPKEY][bgpglobalindex] = bgpglobalnode
			redistributeindex1 := Parameters2Index(v.VRF, "ipv4", "static") + basic.SONICROUTECOMMONREDISTELEMENT
			redistributeindex2 := Parameters2Index(v.VRF, "ipv4", "connected") + basic.SONICROUTECOMMONREDISTELEMENT
			redistributeindex3 := Parameters2Index(v.VRF, "ipv6", "static") + basic.SONICROUTECOMMONREDISTELEMENT
			redistributeindex4 := Parameters2Index(v.VRF, "ipv6", "connected") + basic.SONICROUTECOMMONREDISTELEMENT
			redistributenode1 := sonicmodel.RouteRedistributenode{VrfName: v.VRF, AddrFamily: "ipv4", SrcProtocol: "static", DstProtocol: "bgp"}
			redistributenode2 := sonicmodel.RouteRedistributenode{VrfName: v.VRF, AddrFamily: "ipv4", SrcProtocol: "connected", DstProtocol: "bgp"}
			redistributenode3 := sonicmodel.RouteRedistributenode{VrfName: v.VRF, AddrFamily: "ipv6", SrcProtocol: "static", DstProtocol: "bgp"}
			redistributenode4 := sonicmodel.RouteRedistributenode{VrfName: v.VRF, AddrFamily: "ipv6", SrcProtocol: "connected", DstProtocol: "bgp"}
			c.DiscreteConfiguration[basic.SONICROUTECOMMONKEY][redistributeindex1] = redistributenode1
			c.DiscreteConfiguration[basic.SONICROUTECOMMONKEY][redistributeindex2] = redistributenode2
			c.DiscreteConfiguration[basic.SONICROUTECOMMONKEY][redistributeindex3] = redistributenode3
			c.DiscreteConfiguration[basic.SONICROUTECOMMONKEY][redistributeindex4] = redistributenode4
		}
	}

	if len(data.Familys.Family) > 0 {
		for _, v := range data.Familys.Family {
			v.VRF = VrfNameFormat(v.VRF)
			var bgpafnode sonicmodel.BgpGlobalsAFList
			addressfamily := Familytrans(v.Type, "BGP")
			bgpafindex := Parameters2Index(v.VRF, addressfamily) + basic.SONICBGPGLOBALAFELEMENT
			bgpafnode = BgpGlobalsAfOrganize(v.VRF, v.Type, "BGP")
			c.DiscreteConfiguration[basic.SONICBGPKEY][bgpafindex] = bgpafnode
			//redistribute
			famliy := Familytrans(v.Type, "REDISTRIBUTE")
			redistributeindex1 := Parameters2Index(v.VRF, famliy, "static") + basic.SONICROUTECOMMONREDISTELEMENT
			redistributeindex2 := Parameters2Index(v.VRF, famliy, "connected") + basic.SONICROUTECOMMONREDISTELEMENT
			redistributenode1 := sonicmodel.RouteRedistributenode{VrfName: v.VRF, AddrFamily: famliy, SrcProtocol: "static", DstProtocol: "bgp"}
			redistributenode2 := sonicmodel.RouteRedistributenode{VrfName: v.VRF, AddrFamily: famliy, SrcProtocol: "connected", DstProtocol: "bgp"}
			c.DiscreteConfiguration[basic.SONICROUTECOMMONKEY][redistributeindex1] = redistributenode1
			c.DiscreteConfiguration[basic.SONICROUTECOMMONKEY][redistributeindex2] = redistributenode2
		}
	}

	//重发布是对SONIC来说是单独别的模块配置
	if len(data.Redistributes.Redist) > 0 {
		for _, v := range data.Redistributes.Redist {
			v.VRF = VrfNameFormat(v.VRF)
			err := BGPRedistCheck(&v)
			if err != nil {
				return err
			}
			protocol := BGPProtocoltrans(v.Protocol)
			bgpfamliy := Familytrans(v.Family, "REDISTRIBUTE")
			redistributeindex := Parameters2Index(v.VRF, bgpfamliy, protocol) + basic.SONICROUTECOMMONREDISTELEMENT
			if IndexQueryContext(c.DiscreteConfiguration, basic.SONICROUTECOMMONKEY, redistributeindex) {
				continue
			}
			redistributenode := sonicmodel.RouteRedistributenode{VrfName: v.VRF, AddrFamily: bgpfamliy, SrcProtocol: protocol, DstProtocol: "bgp"}
			c.DiscreteConfiguration[basic.SONICROUTECOMMONKEY][redistributeindex] = redistributenode
		}
	}
	return nil
}

func RoutePolicyEncodeMerge(c *tcontext.Tcontext) error {
	BGP := c.Features["RoutePolicy"]
	data, ok := BGP.(h3cmodel.RoutePolicy)
	if !ok {
		return errors.New("Routepolicy data translate failed")
	}
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICROUTEMAPKEY, basic.SONICROUTEMAPSETKEY)

	if len(data.IPv4PrefixList.PrefixList) > 0 {
		for _, v := range data.IPv4PrefixList.PrefixList {
			err := PrefixListCheck(v)
			if err != nil {
				return err
			}
			ip_prefix := v.Ipv4Address + "/" + v.Ipv4PrefixLength
			nodeindex := strconv.Itoa(v.Index)
			prefixsetindex := Parameters2Index(v.PrefixListName) + basic.SONICIPV4PREFIXSETELEMENT
			prefixnodeindex := Parameters2Index(v.PrefixListName, nodeindex) + basic.SONICPREFIXNODEELEMENT
			var action string = "permit"
			if v.Mode == "1" {
				action = "deny"
			}
			prefixnode := sonicmodel.PrefixEntry{Action: action, IPPrefix: ip_prefix, SetName: v.PrefixListName, SequenceNumber: v.Index}
			var lengthrange string = ".."
			if v.MinPrefixLength != "" || v.MaxPrefixLength != "" {
				lengthrange = v.MinPrefixLength + lengthrange + v.MaxPrefixLength
			} else {
				lengthrange = "exact"
			}
			prefixnode.MasklengthRange = lengthrange
			prefixsetnode := sonicmodel.PrefixSetEntry{Name: v.PrefixListName, Mode: "IPv4"}
			c.DiscreteConfiguration[basic.SONICROUTEMAPSETKEY][prefixsetindex] = prefixsetnode
			c.DiscreteConfiguration[basic.SONICROUTEMAPSETKEY][prefixnodeindex] = prefixnode
		}
	}

	if len(data.IPv6PrefixList.PrefixList) > 0 {
		for _, v := range data.IPv6PrefixList.PrefixList {
			err := PrefixListCheck(v)
			if err != nil {
				return err
			}
			ip_prefix := v.Ipv6Address + "/" + v.Ipv6PrefixLength
			nodeindex := strconv.Itoa(v.Index)
			prefixsetindex := Parameters2Index(v.PrefixListName) + basic.SONICIPV6PREFIXSETELEMENT
			prefixnodeindex := Parameters2Index(v.PrefixListName, nodeindex) + basic.SONICPREFIXNODEELEMENT
			var action string = "permit"
			if v.Mode == "1" {
				action = "deny"
			}
			prefixnode := sonicmodel.PrefixEntry{Action: action, IPPrefix: ip_prefix, SetName: v.PrefixListName, SequenceNumber: v.Index}
			var lengthrange string = ".."
			if v.MinPrefixLength != "" || v.MaxPrefixLength != "" {
				lengthrange = v.MinPrefixLength + lengthrange + v.MaxPrefixLength
			} else {
				lengthrange = "exact"
			}
			prefixnode.MasklengthRange = lengthrange
			prefixsetnode := sonicmodel.PrefixSetEntry{Name: v.PrefixListName, Mode: "IPv6"}
			c.DiscreteConfiguration[basic.SONICROUTEMAPSETKEY][prefixsetindex] = prefixsetnode
			c.DiscreteConfiguration[basic.SONICROUTEMAPSETKEY][prefixnodeindex] = prefixnode
		}
	}

	if len(data.Policy.Entry) > 0 {
		for _, v := range data.Policy.Entry {
			err := RoutepolicyCheck(v)
			if err != nil {
				return err
			}

			policyindex := Parameters2Index(v.PolicyName, strconv.Itoa(v.Index)) + basic.SONICROUTEMAPELELMENT
			var opreation string = "permit"
			if v.Mode == "1" {
				opreation = "deny"
			}
			routemapnode := sonicmodel.RouteMapEntry{RouteMapName: v.PolicyName, StmtName: v.Index,
				RouteOperation: opreation}

			if v.Match.IPv4AddressPrefixList != "" {
				routemapnode.MatchPrefixSet = v.Match.IPv4AddressPrefixList
			}
			if v.Match.IPv6AddressPrefixList != "" {
				routemapnode.MatchIPv6PrefixSet = v.Match.IPv6AddressPrefixList
			}
			if v.Match.Tag != 0 {
				routemapnode.MatchTag = v.Match.Tag
			}
			if v.Apply.LocalPreference != 0 {
				routemapnode.SetLocalPref = v.Apply.LocalPreference
			}
			if v.Apply.IPv6NextHop != "" {
				routemapnode.SetIPv6NextHopGlobal = v.Apply.IPv6NextHop
				routemapnode.SetIPv6NextHopPreferGlobal = true
			}

			if v.ApplyIpv4NextHop.NextHopAddr != "" {
				routemapnode.SetNextHop = v.ApplyIpv4NextHop.NextHopAddr
			}
			c.DiscreteConfiguration[basic.SONICROUTEMAPKEY][policyindex] = routemapnode
		}
	}
	return nil
}

func RoutePolicyEncodeRemove(c *tcontext.Tcontext) error {
	BGP := c.Features["RoutePolicy"]
	data, ok := BGP.(h3cmodel.RoutePolicy)
	if !ok {
		return errors.New("Routepolicy data translate failed")
	}
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICROUTEMAPKEY, basic.SONICROUTEMAPSETKEY)

	if len(data.IPv4PrefixList.PrefixList) > 0 {
		for _, v := range data.IPv4PrefixList.PrefixList {
			err := PrefixListCheck(v)
			if err != nil {
				return err
			}
			nodeindex := strconv.Itoa(v.Index)
			prefixnodeindex := Parameters2Index(v.PrefixListName, nodeindex) + basic.SONICPREFIXNODEELEMENT
			prefixnode := sonicmodel.PrefixEntry{SetName: v.PrefixListName, SequenceNumber: v.Index}
			c.DiscreteConfiguration[basic.SONICROUTEMAPSETKEY][prefixnodeindex] = prefixnode
		}
	}

	if len(data.IPv6PrefixList.PrefixList) > 0 {
		for _, v := range data.IPv6PrefixList.PrefixList {
			err := PrefixListCheck(v)
			if err != nil {
				return err
			}
			nodeindex := strconv.Itoa(v.Index)
			prefixnodeindex := Parameters2Index(v.PrefixListName, nodeindex) + basic.SONICPREFIXNODEELEMENT
			prefixnode := sonicmodel.PrefixEntry{SetName: v.PrefixListName, SequenceNumber: v.Index}
			c.DiscreteConfiguration[basic.SONICROUTEMAPSETKEY][prefixnodeindex] = prefixnode
		}
	}

	if len(data.Policy.Entry) > 0 {
		for _, v := range data.Policy.Entry {
			err := RoutepolicyCheck(v)
			if err != nil {
				return err
			}
			policyindex := Parameters2Index(v.PolicyName, strconv.Itoa(v.Index)) + basic.SONICROUTEMAPELELMENT
			var opreation string = "permit"
			if v.Mode == "1" {
				opreation = "deny"
			}
			routemapnode := sonicmodel.RouteMapEntry{RouteMapName: v.PolicyName, StmtName: v.Index,
				RouteOperation: opreation}

			if v.Match.IPv4AddressPrefixList != "" {
				routemapnode.MatchPrefixSet = v.Match.IPv4AddressPrefixList
			}
			if v.Match.IPv6AddressPrefixList != "" {
				routemapnode.MatchIPv6PrefixSet = v.Match.IPv6AddressPrefixList
			}
			if v.Match.Tag != 0 {
				routemapnode.MatchTag = v.Match.Tag
			}
			if v.Apply.LocalPreference != 0 {
				routemapnode.SetLocalPref = v.Apply.LocalPreference
			}
			if v.Apply.IPv6NextHop != "" {
				routemapnode.SetIPv6NextHopGlobal = v.Apply.IPv6NextHop
				routemapnode.SetIPv6NextHopPreferGlobal = true
			}

			if v.ApplyIpv4NextHop.NextHopAddr != "" {
				routemapnode.SetNextHop = v.ApplyIpv4NextHop.NextHopAddr
			}
			c.DiscreteConfiguration[basic.SONICROUTEMAPKEY][policyindex] = routemapnode
		}
	}
	return nil
}

func StaticRouteEncodeMerge(c *tcontext.Tcontext) error {
	StaticRoute := c.Features["StaticRoute"]
	data, ok := StaticRoute.(h3cmodel.StaticRoute)
	if !ok {
		return errors.New("StaticRoute data xml translate failed")
	}
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICSTATICROUTEKEY)

	if len(data.Ipv4StaticRouteConfigurations.RouteEntries) > 0 {
		for _, v := range data.Ipv4StaticRouteConfigurations.RouteEntries {
			v.DestVrfIndex = VrfNameFormat(v.DestVrfIndex)
			v.NexthopVrfIndex = VrfNameFormat(v.NexthopVrfIndex)
			routeindex := Parameters2Index(v.DestVrfIndex, v.Ipv4Address, v.NexthopIpv4Address) + basic.SONICSTATICROUTEELEMENT
			prefix := v.Ipv4Address + "/" + v.Ipv4PrefixLength
			staticroutenode := sonicmodel.StaticRouteEntry{
				VrfName:    v.DestVrfIndex,
				Prefix:     prefix,
				Nexthop:    v.NexthopIpv4Address,
				NexthopVrf: v.NexthopVrfIndex,
			}
			if v.Preference != "" {
				staticroutenode.Distance = &v.Preference
			}
			if v.Tag != "" {
				staticroutenode.Tag = &v.Tag
			}
			if v.IfIndex != "" {
				if v.IfIndex == "NULL0" {
					blackhole := "true"
					staticroutenode.Blackhole = &blackhole
				} else if strings.Contains(v.IfIndex, "Vlan") || strings.Contains(v.IfIndex, "vlan") {
					vlanid, _ := GetInterfaceString(v.IfIndex)
					ifname := "Vlan" + vlanid
					staticroutenode.Ifname = &ifname
				}
			}
			c.DiscreteConfiguration[basic.SONICSTATICROUTEKEY][routeindex] = staticroutenode
		}
	}
	if len(data.Ipv6StaticRouteConfigurations.IPv6RouteEntries) > 0 {
		for _, v := range data.Ipv6StaticRouteConfigurations.IPv6RouteEntries {
			v.DestVrfIndex = VrfNameFormat(v.DestVrfIndex)
			v.NexthopVrfIndex = VrfNameFormat(v.NexthopVrfIndex)
			routeindex := Parameters2Index(v.DestVrfIndex, v.Ipv6Address, v.NexthopIpv6Address) + basic.SONICSTATICROUTEELEMENT
			prefix := v.Ipv6Address + "/" + v.Ipv6PrefixLength
			staticroutenode := sonicmodel.StaticRouteEntry{
				VrfName:    v.DestVrfIndex,
				Prefix:     prefix,
				Nexthop:    v.NexthopIpv6Address,
				NexthopVrf: v.NexthopVrfIndex,
			}
			if v.Preference != "" {
				staticroutenode.Distance = &v.Preference
			}
			if v.Tag != "" {
				staticroutenode.Tag = &v.Tag
			}
			if v.IfIndex != "" {
				if v.IfIndex == "NULL0" {
					blackhole := "true"
					staticroutenode.Blackhole = &blackhole
				} else if strings.Contains(v.IfIndex, "Vlan") || strings.Contains(v.IfIndex, "vlan") {
					vlanid, _ := GetInterfaceString(v.IfIndex)
					ifname := "Vlan" + vlanid
					staticroutenode.Ifname = &ifname
				}
			}
			c.DiscreteConfiguration[basic.SONICSTATICROUTEKEY][routeindex] = staticroutenode
		}
	}
	return nil
}

func StaticRouteEncodeRemove(c *tcontext.Tcontext) error {
	StaticRoute := c.Features["StaticRoute"]
	data, ok := StaticRoute.(h3cmodel.StaticRoute)
	if !ok {
		return errors.New("StaticRoute data xml translate failed")
	}
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICSTATICROUTEKEY)

	if len(data.Ipv4StaticRouteConfigurations.RouteEntries) > 0 {
		for _, v := range data.Ipv4StaticRouteConfigurations.RouteEntries {
			v.DestVrfIndex = VrfNameFormat(v.DestVrfIndex)
			v.NexthopVrfIndex = VrfNameFormat(v.NexthopVrfIndex)
			routeindex := Parameters2Index(v.DestVrfIndex, v.Ipv4Address) + basic.SONICSTATICROUTEELEMENT
			prefix := v.Ipv4Address + "/" + v.Ipv4PrefixLength
			staticroutenode := sonicmodel.StaticRouteEntry{
				VrfName: v.DestVrfIndex,
				Prefix:  prefix,
			}
			c.DiscreteConfiguration[basic.SONICSTATICROUTEKEY][routeindex] = staticroutenode
		}
	}
	if len(data.Ipv6StaticRouteConfigurations.IPv6RouteEntries) > 0 {
		for _, v := range data.Ipv6StaticRouteConfigurations.IPv6RouteEntries {
			v.DestVrfIndex = VrfNameFormat(v.DestVrfIndex)
			v.NexthopVrfIndex = VrfNameFormat(v.NexthopVrfIndex)
			routeindex := Parameters2Index(v.DestVrfIndex, v.Ipv6Address) + basic.SONICSTATICROUTEELEMENT
			prefix := v.Ipv6Address + "/" + v.Ipv6PrefixLength
			staticroutenode := sonicmodel.StaticRouteEntry{
				VrfName: v.DestVrfIndex,
				Prefix:  prefix,
			}
			c.DiscreteConfiguration[basic.SONICSTATICROUTEKEY][routeindex] = staticroutenode
		}
	}
	return nil
}

func IPV4ADDRESSEncodeMerge(c *tcontext.Tcontext) error {
	ipv4address := c.Features["IPV4ADDRESS"]
	data, ok := ipv4address.(h3cmodel.IPV4ADDRESS)
	if !ok {
		return errors.New("IPV4ADDRESS data xml translate failed")
	}
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICADDRESS)

	if len(data.Ipv4Addresses.Ipv4Address) > 0 {
		for _, v := range data.Ipv4Addresses.Ipv4Address {
			if v.IfIndex == "" || v.Ipv4Address == "" || v.Ipv4Mask == "" {
				glog.Errorf("feature ipv4address index miss")
				return errors.New("feature ipv4address index element misssing")
			}
			//接口类型分类
			if strings.Contains(v.IfIndex, "Vlan") || strings.Contains(v.IfIndex, "vlan") {
				vlanid, _ := GetInterfaceString(v.IfIndex)
				prefix := MaskToPrefix(v.Ipv4Mask)
				cidr := v.Ipv4Address + "/" + prefix
				addressindex := Parameters2Index("Vlan"+vlanid, cidr) + basic.SONICVLANINTERFACEIPADDRELEMENT
				var second bool
				if v.AddressOrigin == 2 {
					second = true
				}
				ipv4addressnode := sonicmodel.VLANInterfaceIPAddr{
					VlanName:  "Vlan" + vlanid,
					IpPrefix:  cidr,
					Secondary: second,
				}
				c.DiscreteConfiguration[basic.SONICADDRESS][addressindex] = ipv4addressnode
			}
			if strings.Contains(v.IfIndex, "Loopback") || strings.Contains(v.IfIndex, "loopback") {
				loopbackid, _ := GetInterfaceString(v.IfIndex)
				prefix := MaskToPrefix(v.Ipv4Mask)
				cidr := v.Ipv4Address + "/" + prefix
				addressindex := Parameters2Index("Loopback"+loopbackid, cidr) + basic.SONICLOOPBACKINTERFACEIPADDRELEMENT
				var second bool
				if v.AddressOrigin == 2 {
					second = true
				}
				ipv4addressnode := sonicmodel.LoopbackInterfaceIPAddr{
					LoIfName:  "Loopback" + loopbackid,
					IpPrefix:  cidr,
					Secondary: second,
				}
				c.DiscreteConfiguration[basic.SONICADDRESS][addressindex] = ipv4addressnode
			}
		}
	}
	return nil
}

// 先不考虑接口删除IP场景
func IPV4ADDRESSEncodeRemove(c *tcontext.Tcontext) error {
	return nil
}

func IPV6ADDRESSEncodeMerge(c *tcontext.Tcontext) error {
	ipv6address := c.Features["IPV6ADDRESS"]
	data, ok := ipv6address.(h3cmodel.IPV6ADDRESS)
	if !ok {
		return errors.New("IPV4ADDRESS data xml translate failed")
	}
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICADDRESS)

	if len(data.Ipv6AddressesConfig.AddressEntry) > 0 {
		for _, v := range data.Ipv6AddressesConfig.AddressEntry {
			if v.IfIndex == "" || v.Ipv6Address == "" || v.Ipv6PrefixLength == "" {
				glog.Errorf("feature ipv6address index miss")
				return errors.New("feature ipv6address index element misssing")
			}
			//接口类型分类
			if strings.Contains(v.IfIndex, "Vlan") || strings.Contains(v.IfIndex, "vlan") {
				vlanid, _ := GetInterfaceString(v.IfIndex)
				addressindex := Parameters2Index("Vlan"+vlanid, v.Ipv6Address, v.Ipv6PrefixLength) + basic.SONICVLANINTERFACEIPADDRELEMENT
				prefix := v.Ipv6Address + "/" + v.Ipv6PrefixLength
				ipv6addressnode := sonicmodel.VLANInterfaceIPAddr{
					VlanName:  "Vlan" + vlanid,
					IpPrefix:  prefix,
					Secondary: false,
				}
				c.DiscreteConfiguration[basic.SONICADDRESS][addressindex] = ipv6addressnode
			}
			if strings.Contains(v.IfIndex, "Loopback") || strings.Contains(v.IfIndex, "loopback") {
				vlanid, _ := GetInterfaceString(v.IfIndex)
				addressindex := Parameters2Index("Loopback"+vlanid, v.Ipv6Address, v.Ipv6PrefixLength) + basic.SONICLOOPBACKINTERFACEIPADDRELEMENT
				prefix := v.Ipv6Address + "/" + v.Ipv6PrefixLength
				ipv6addressnode := sonicmodel.LoopbackInterfaceIPAddr{
					LoIfName:  "Loopback" + vlanid,
					IpPrefix:  prefix,
					Secondary: false,
				}
				c.DiscreteConfiguration[basic.SONICADDRESS][addressindex] = ipv6addressnode
			}
		}
	}
	return nil
}

func IPV6ADDRESSEncodeRemove(c *tcontext.Tcontext) error {
	return nil
}

func IfmgrEncodeMerge(c *tcontext.Tcontext) error {
	ifmgr := c.Features["Ifmgr"]
	data, ok := ifmgr.(h3cmodel.Ifmgr)
	if !ok {
		return errors.New("ifmgr data xml translate failed")
	}
	CreateFeaturemap(c.DiscreteConfiguration, basic.SONICINTERFACEMAC)

	if len(data.Interfaces.Interface) > 0 {
		for _, v := range data.Interfaces.Interface {
			if v.IfIndex == "" {
				glog.Errorf("feature ifmgr index miss")
				return errors.New("feature ifmgr index element misssing")
			}
			//暂时只考虑下发mac
			if v.MAC == "" {
				continue
			}
			//接口类型分类
			if strings.Contains(v.IfIndex, "Vlan") || strings.Contains(v.IfIndex, "vlan") {
				vlanid, _ := GetInterfaceString(v.IfIndex)
				interfaceindex := Parameters2Index("Vlan"+vlanid) + basic.SONICINTERFACEMACELEMENT
				c.DiscreteConfiguration[basic.SONICINTERFACEMAC][interfaceindex] = true
			}
		}
	}
	return nil
}

// 不考虑
func IfmgrEncodeRemove(c *tcontext.Tcontext) error {
	return nil
}

func PrefixListCheck(value h3cmodel.PrefixList) error {
	if value.PrefixListName == "" || value.Index == 0 {
		return errors.New("prefix_list index fields missing")
	}
	return nil
}

func RoutepolicyCheck(value h3cmodel.Entry) error {
	if value.PolicyName == "" || value.Index == 0 {
		return errors.New("routepolicy policy index fields missing")
	}
	return nil
}

func BGPProtocoltrans(protocol int) string {
	switch protocol {
	case 1:
		return "connected"
	case 2:
		return "static"
	}
	return ""
}

func Familytrans(familytype int, feature string) string {
	switch feature {
	case "L3vpn":
		switch familytype {
		case 1:
			return "ipv4_unicast"
		case 2:
			return "ipv6_unicast"
		case 4:
			return "l2vpn_evpn"
		}
	case "BGP":
		switch familytype {
		case 1:
			return "ipv4_unicast"
		case 5:
			return "ipv6_unicast"
		case 9:
			return "l2vpn_evpn"
		}
	case "REDISTRIBUTE":
		switch familytype {
		case 1:
			return "ipv4"
		case 5:
			return "ipv6"
		}
	}
	return ""
}

func BGPRedistCheck(bgpredist *h3cmodel.Redist) error {
	if bgpredist.Protocol <= 0 {
		return errors.New("protocol element misssing")
	}
	if bgpredist.VRF == "" {
		return errors.New("vrf index element misssing")
	}
	if bgpredist.Family <= 0 {
		return errors.New("family element misssing")
	}
	return nil
}

func IfmgrLogicalCheck(logicalint *h3cmodel.Interface_logical) error {
	if logicalint.IfTypeExt == "" {
		return errors.New("IfTypeExt index element misssing")
	}
	if logicalint.Number == "" {
		return errors.New("Number index element misssing")
	}
	return nil
}

func VSIInterfaceCheck(vsiintface *h3cmodel.VSIInterface) error {
	if vsiintface.ID == 0 {
		return errors.New("ID index element misssing")
	}
	return nil
}

func VRFCheck(vrf *h3cmodel.VRF) error {
	if vrf.VRF == "" {
		return errors.New("vrf index element misssing")
	}
	return nil
}

func BindCheck(bind *h3cmodel.Bind) error {
	if bind.VRF == "" {
		return errors.New("[Bind]vrf index element missing")
	}
	if bind.IfIndex == "" {
		return errors.New("bind's ifindex element missing")
	}
	return nil
}

func L3vpnRTCheck(rt *h3cmodel.RT) error {
	if rt.VRF == "" {
		return errors.New("[RT]vrf index element misssing")
	}
	if rt.AddressFamily <= 0 {
		return errors.New("[RT]address family element error")
	}
	if rt.RTType <= 0 {
		return errors.New("[RT]RTType element error")
	}
	if rt.RTEntry == "" {
		return errors.New("[RT]RTEntry element error")
	}
	return nil
}

func VlanListOrganize(id int, mtu int) sonicmodel.VLANNode {
	var vlan sonicmodel.VLANNode
	vlan.VLANID = id
	vlan.Name = "Vlan" + strconv.Itoa(id)
	vlan.MTU = mtu
	vlan.AdminStatus = "up"
	return vlan
}

func VrfOrganize(name string, vni int) sonicmodel.Vrf {
	var vrf sonicmodel.Vrf
	vrf.VrfName = name
	vrf.Vni = vni
	vrf.Fallback = false
	return vrf
}
func VxlanTunnelMapOrganize(vlan, vni int) sonicmodel.VxlanTunnelMap {
	var vxlantunnel sonicmodel.VxlanTunnelMap
	vxlantunnel.Name = basic.TUNNELNAME
	vlanstr := strconv.Itoa(vlan)
	vxlantunnel.Mapname = "map_" + strconv.Itoa(vni) + "_Vlan"
	vxlantunnel.Vlan = "Vlan" + vlanstr
	vxlantunnel.Vni = vni
	return vxlantunnel
}

func BgpGlobalsAfOrganize(vrf string, afi_sfi int, feature string) sonicmodel.BgpGlobalsAFList {
	var bgpglobalafnode sonicmodel.BgpGlobalsAFList
	bgpglobalafnode.VrfName = vrf
	switch feature {
	case "L3vpn":
		switch afi_sfi {
		case 1:
			bgpglobalafnode.AFISAFI = "ipv4_unicast"
		case 2:
			bgpglobalafnode.AFISAFI = "ipv6_unicast"
		case 4:
			bgpglobalafnode.AFISAFI = "l2vpn_evpn"
		}
	case "BGP":
		switch afi_sfi {
		case 1:
			bgpglobalafnode.AFISAFI = "ipv4_unicast"
		case 5:
			bgpglobalafnode.AFISAFI = "ipv6_unicast"
		case 9:
			bgpglobalafnode.AFISAFI = "l2vpn_evpn"
		}
	}
	return bgpglobalafnode
}

// { ↑所有设备数据在这解析好,直接把数据丢给sonic处理↑  }

// func GetResourceInfo(tcache map[string]int, key string) (idx int, err error) {
// 	idx, ok := tcache[key]
// 	if !ok {
// 		v, err := redisclient.IndexGet(key)
// 		if err != nil {
// 			return 0, err
// 		}
// 		idx, _ = strconv.Atoi(v)
// 	}
// 	return idx, nil
// }

// func SetResourceInfo(c *tcontext.Tcontext, key string, value int) {
// 	c.Cachedata[key] = value
// 	indexmap := c.SonicConfig[basic.SONICINDEX].(map[string]int)
// 	indexmap[key] = value
// }

// func AllocationResouceIndex(c *tcontext.Tcontext, rtype string, rname string) (Exist bool, index int, err error) {
// 	indexkey := rtype + rname + "_INDEX"
// 	if index, ok := c.Cachedata[indexkey]; ok {
// 		return false, index, nil
// 	}
// 	v, err := redisclient.IndexGet(indexkey)
// 	if err != nil {
// 		if err == redis.Nil {
// 			switch rtype {
// 			case "VRF":
// 				index = rand.Intn(1000)
// 			case "VLANMapping":
// 				index = basic.VLANBASE + rand.Intn(300)
// 			}
// 			c.Cachedata[indexkey] = index
// 			indexmap := c.SonicConfig[basic.SONICINDEX].(map[string]int)
// 			indexmap[indexkey] = index
// 			return false, index, nil
// 		} else {
// 			return false, 0, err
// 		}
// 	}
// 	index, _ = strconv.Atoi(v)
// 	return true, index, nil
// }

// 直接取余得出vlan
func L3vni2Vlan(l3vni int) int {
	return basic.VLANBASE + l3vni%basic.VRFCAP
}

func VrfNameFormat(name string) string {
	if len(name) < 12 {
		return "Vrf" + name
	} else {
		return "Vrf" + name[len(name)-12:]
	}
}

func GetInterfaceString(example string) (num string, err error) {

	//Vsi-interface60001
	intre := regexp.MustCompile(`\d+`)
	nums := intre.FindAllString(example, -1)
	if nums == nil || len(nums) < 1 {
		restr := fmt.Sprintf("example string inconrrect %s", example)
		glog.Errorf(restr)
		return "", errors.New(restr)
	} else {
		return nums[0], nil
	}
}

func MaskToPrefix(mask string) string {
	maskbyte := net.ParseIP(mask).To4()
	sz, _ := net.IPv4Mask(maskbyte[0], maskbyte[1], maskbyte[2], maskbyte[3]).Size()
	return strconv.Itoa(sz)
}
