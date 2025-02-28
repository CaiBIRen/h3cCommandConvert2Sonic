package device

import (
	"sonic-unis-framework/configuration"
	"sonic-unis-framework/tcontext"

	"github.com/antchfx/xmlquery"
	"github.com/coreos/pkg/capnslog"
)

var glog = capnslog.NewPackageLogger("sonic-unis-framework", "DEVICE")

type Devicehandler interface {
	Role() string
	Decode(featuremap map[string]*xmlquery.Node, c *tcontext.Tcontext) error
	EncodeMerge(c *tcontext.Tcontext) error
	EncodeRemove(c *tcontext.Tcontext) error
	EncodeGet(featuremap map[string]*xmlquery.Node, c *tcontext.Tcontext) error
	EncodeAction(c *tcontext.Tcontext) error
	IntegrationReply(c *tcontext.Tcontext) (string, error) //to get request response
}

var Devicehdl Devicehandler

func NewDevice() {
	if configuration.ServiceConfiguration.Company == "H3C" {
		Devicehdl = H3cdevice{}
	}
}
