package vision

import (
	"image"
	"path"
	"strconv"
	"strings"

	context "context"

	"github.com/pkg/errors"
	"github.com/c3sr/config"
	"github.com/c3sr/dldataset"
	"github.com/c3sr/dlframework"
	"github.com/c3sr/dlframework/framework/feature"
	"github.com/c3sr/image/types"
	mnistLoader "github.com/unixpickle/mnist"
)

// MNIST ...
type MNIST struct {
	base
	trainingData mnistLoader.DataSet
	testData     mnistLoader.DataSet
}

var mnist *MNIST

// MNISTLabeledImage ...
type MNISTLabeledImage struct {
	label string
	data  *types.RGBImage
}

// Label ...
func (l MNISTLabeledImage) Label() string {
	return l.label
}

// Feature ...
func (l MNISTLabeledImage) Feature() *dlframework.Feature {
	return feature.New(
		feature.ClassificationLabel(l.Label()),
	)
}

// Features ...
func (l MNISTLabeledImage) Features() dlframework.Features {
	return dlframework.Features([]*dlframework.Feature{l.Feature()})
}

// Data ...
func (l MNISTLabeledImage) Data() (interface{}, error) {
	return l.data, nil
}

// Name ...
func (*MNIST) Name() string {
	return "MNIST"
}

// CanonicalName ...
func (d *MNIST) CanonicalName() string {
	category := strings.ToLower(d.Category())
	name := strings.ToLower(d.Name())
	key := path.Join(category, name)
	return key
}

// New ...
func (d *MNIST) New(ctx context.Context) (dldataset.Dataset, error) {
	return mnist, nil
}

func (d *MNIST) Load(ctx context.Context) error {
	return nil
}

// Download ...
func (d *MNIST) Download(ctx context.Context) error {
	return nil
}

// List ...
func (d *MNIST) List(ctx context.Context) ([]string, error) {
	lst := []string{}
	for ii := range d.trainingData.Samples {
		lst = append(lst, "train/"+strconv.Itoa(ii))
	}
	for ii := range d.testData.Samples {
		lst = append(lst, "test/"+strconv.Itoa(ii))
	}
	return lst, nil
}

// Get ...
func (d *MNIST) Get(ctx context.Context, name string) (dldataset.LabeledData, error) {
	var dataset mnistLoader.DataSet
	if strings.HasPrefix(name, "train/") {
		name = strings.TrimPrefix(name, "train/")
		dataset = d.trainingData
	} else if strings.HasPrefix(name, "test/") {
		name = strings.TrimPrefix(name, "test/")
		dataset = d.trainingData
	} else {
		return nil, errors.Errorf("cannot find %s in the mnist dataset", name)
	}
	idx, err := strconv.Atoi(name)
	if err != nil {
		return nil, errors.Errorf("expecting an integer, but got %s", name)
	}
	if idx >= len(dataset.Samples) {
		return nil, errors.Errorf("the index %d is out of range %d", idx, len(dataset.Samples))
	}

	elem := dataset.Samples[idx]

	img := types.NewRGBImage(image.Rect(0, 0, dataset.Width, dataset.Height))
	data := img.Pix

	for ii, intensity := range elem.Intensities {
		if intensity == 1 {
			data[3*ii+0] = byte(1)
			data[3*ii+1] = byte(1)
			data[3*ii+2] = byte(1)
		} else {
			data[3*ii+0] = byte(0)
			data[3*ii+1] = byte(0)
			data[3*ii+2] = byte(0)
		}
	}

	return &MNISTLabeledImage{
		data:  img,
		label: strconv.Itoa(elem.Label),
	}, nil
}

func (d *MNIST) Next(ctx context.Context) (dldataset.LabeledData, error) {
	return nil, errors.New("next iterator is not implemented for " + d.CanonicalName())
}

// Close ...
func (d *MNIST) Close() error {
	return nil
}

func init() {
	config.AfterInit(func() {
		mnist = &MNIST{
			base: base{
				ctx: context.Background(),
			},
			trainingData: mnistLoader.LoadTestingDataSet(),
			testData:     mnistLoader.LoadTestingDataSet(),
		}
		dldataset.Register(mnist)
	})
}
