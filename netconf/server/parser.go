//
// Software Name: sonic-netconf-server
// SPDX-FileCopyrightText: Copyright (c) Orange SA
// SPDX-License-Identifier: Apache 2.0
//
// This software is distributed under the Apache 2.0 licence,
// the text of which is available at https://opensource.org/license/apache-2-0/
// or see the "LICENSE" file for more details.
//
// Authors: hossam4.hassan@orange.com, abdelmuhaimen.seaudi@orange.com
// Software description: RFC compliant NETCONF server implementation for SONiC
//

package server

import (
	"errors"
	"sonic-unis-framework/basic"
	"sonic-unis-framework/device"
	"sonic-unis-framework/tcontext"

	"github.com/antchfx/xmlquery"
)

func ParseHelloRequest(node *xmlquery.Node) error {
	helloNode := xmlquery.FindOne(node, "//hello")
	if helloNode == nil {
		return errors.New("[Missing data] Need hello element")
	}

	cnode := xmlquery.FindOne(node, "//capability")
	if cnode == nil {
		return errors.New("[Missing data] Need capability element")
	}

	return nil

}

// 对xml报文进行整理,
func FormatingRPCXML(topnode []*xmlquery.Node) map[string]*xmlquery.Node {
	featuremap := make(map[string]*xmlquery.Node)
	for _, f := range topnode {
		if _, ok := featuremap[f.Data]; !ok {
			featuremap[f.Data] = f
		} else {
			f1 := xmlquery.Find(f, "./*")
			for _, ff := range f1 {
				xmlquery.AddChild(featuremap[f.Data], ff)
			}
		}
	}
	return featuremap
}

func Namespace() {}
func ParseGetRequest(node *xmlquery.Node) (map[string]*xmlquery.Node, error) {
	topNode := xmlquery.FindOne(node, "//top")
	if topNode == nil {
		return nil, errors.New("[Missing data] Need top element. Complete configuration retrival currently not supported")
	}

	containers := xmlquery.Find(topNode, "./*")

	featuremap := FormatingRPCXML(containers)
	// fmt.Println("____________", featuremap["LLDP"].OutputXML(true))
	return featuremap, nil
}

func ParseConfigRequest(node *xmlquery.Node) (*tcontext.Tcontext, *tcontext.Tcontext, error) {
	addc, delc := tcontext.NewTcontext(), tcontext.NewTcontext()
	messageid := node.SelectAttr("message-id")
	topNode := xmlquery.FindOne(node, "//top")
	if topNode == nil {
		return nil, nil, errors.New("[Missing data] Need top element. Complete configuration retrival currently not supported")
	}
	if topNode.SelectAttr("nc:operation") == basic.OPERMERGE {
		addc.Messageid = messageid
		addc.Operation = basic.OPERMERGE
		containers := xmlquery.Find(topNode, "./*")
		featuremap := FormatingRPCXML(containers)
		device.Devicehdl.Decode(featuremap, &addc)
		delc.Err = errors.New("no remove feature")
		return &addc, &delc, nil
	} else if topNode.SelectAttr("nc:operation") == basic.OPERREMOVE {
		delc.Messageid = messageid
		delc.Operation = basic.OPERREMOVE
		containers := xmlquery.Find(topNode, "./*")
		featuremap := FormatingRPCXML(containers)
		device.Devicehdl.Decode(featuremap, &delc)
		addc.Err = errors.New("no merge feature")
		return &addc, &delc, nil
	}
	containers := xmlquery.Find(topNode, "./*")
	var addlist, dellist []*xmlquery.Node
	for _, feature := range containers {
		operation := feature.SelectAttr("nc:operation")
		if operation == "" || operation == basic.OPERMERGE {
			addlist = append(addlist, feature)
		} else if operation == basic.OPERREMOVE {
			dellist = append(dellist, feature)
		} else {
			return nil, nil, errors.New("unsupported operation now")
		}
	}
	if len(addlist) > 0 {
		addc.Messageid = messageid
		addc.Operation = basic.OPERMERGE
		addfeaturemap := FormatingRPCXML(addlist)
		device.Devicehdl.Decode(addfeaturemap, &addc)
	} else {
		addc.Err = errors.New("no merge feature")
	}
	if len(dellist) > 0 {
		delc.Messageid = messageid
		delc.Operation = basic.OPERREMOVE
		delfeaturemap := FormatingRPCXML(addlist)
		device.Devicehdl.Decode(delfeaturemap, &delc)
	} else {
		delc.Err = errors.New("no remove feature")
	}
	return &addc, &delc, nil
}

func ParseActionRequest(node *xmlquery.Node) (*tcontext.Tcontext, error) {
	actionc := tcontext.NewTcontext()
	messageid := node.SelectAttr("message-id")
	actionc.Messageid = messageid
	actionc.Operation = "action"
	topNode := xmlquery.FindOne(node, "//top")
	if topNode == nil {
		return nil, errors.New("[Missing data] Need top element. Complete configuration retrival currently not supported")
	}

	containers := xmlquery.Find(topNode, "./*")

	featuremap := FormatingRPCXML(containers)
	err := device.Devicehdl.Decode(featuremap, &actionc)
	if err != nil {
		return nil, err
	}
	return &actionc, nil
}
