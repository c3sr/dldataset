package reader

import "github.com/c3sr/image/types"

// ImageRecord ...
type ImageRecord struct {
	ID         uint64
	LabelIndex float32
	Image      *types.RGBImage
}

// ImageSegmentationRecord ...
type ImageSegmentationRecord struct {
	ID         uint64
	LabelIndex float32
	Image      *types.RGBImage
}
