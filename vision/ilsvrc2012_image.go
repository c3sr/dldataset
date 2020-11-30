package vision

import (
	"github.com/c3sr/dlframework"
	"github.com/c3sr/dlframework/framework/feature"
	"github.com/c3sr/image/types"
)

// ILSVRC2012ValidationLabeledImage ...
type ILSVRC2012ValidationLabeledImage struct {
	label string
	data  *types.RGBImage
}

// Label ...
func (l ILSVRC2012ValidationLabeledImage) Label() string {
	return l.label
}

// Data ...
func (l ILSVRC2012ValidationLabeledImage) Data() (interface{}, error) {
	return l.data, nil
}

// Feature ...
func (d *ILSVRC2012ValidationLabeledImage) Feature() *dlframework.Feature {
	return feature.New(
		feature.ClassificationLabel(d.Label()),
	)
}

// Features ...
func (l ILSVRC2012ValidationLabeledImage) Features() dlframework.Features {
	return dlframework.Features([]*dlframework.Feature{l.Feature()})
}
