package vision

import (
	"bufio"
	"encoding/binary"
	"image"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	context "context"

	"github.com/Unknwon/com"
	"github.com/pkg/errors"
	"github.com/c3sr/config"
	"github.com/c3sr/dldataset"
	"github.com/c3sr/dlframework"
	"github.com/c3sr/dlframework/framework/feature"
	"github.com/c3sr/downloadmanager"
	"github.com/c3sr/image/types"
	"github.com/c3sr/utils"
)

var cifar100 *CIFAR100

// CIFAR100 ...
type CIFAR100 struct {
	base
	url                  string
	fileName             string
	extractedFolderName  string
	md5sum               string
	trainFileNameList    map[string]string
	testFileNameList     map[string]string
	fineLabelsFileName   string
	coarseLabelsFileName string
	fineLabels           []string
	coarseLabels         []string
	fineLabelByteSize    int
	coarseLabelByteSize  int
	pixelByteSize        int
	imageDimensions      []int
	data                 map[string]CIFAR100LabeledImage
	isDownloaded         bool
}

// CIFAR100LabeledImage ...
type CIFAR100LabeledImage struct {
	coarseLabel string
	fineLabel   string
	data        *types.RGBImage
}

// CoarseLabel ...
func (l CIFAR100LabeledImage) CoarseLabel() string {
	return l.coarseLabel
}

// FineLabel ...
func (l CIFAR100LabeledImage) FineLabel() string {
	return l.fineLabel
}

// Label ...
func (l CIFAR100LabeledImage) Label() string {
	return l.FineLabel()
}

// Feature ...
func (l CIFAR100LabeledImage) Feature() *dlframework.Feature {
	return feature.New(
		feature.ClassificationLabel(l.fineLabel),
	)
}

// Features ...
func (l CIFAR100LabeledImage) Features() dlframework.Features {
	return dlframework.Features([]*dlframework.Feature{l.Feature()})
}

// Data ...
func (l CIFAR100LabeledImage) Data() (interface{}, error) {
	return l.data, nil
}

// Name ...
func (*CIFAR100) Name() string {
	return "CIFAR100"
}

// CanonicalName ...
func (d *CIFAR100) CanonicalName() string {
	category := strings.ToLower(d.Category())
	name := strings.ToLower(d.Name())
	key := path.Join(category, name)
	return key
}

// New ...
func (d *CIFAR100) New(ctx context.Context) (dldataset.Dataset, error) {
	return cifar100, nil
}

func (d *CIFAR100) Load(ctx context.Context) error {
	return nil
}

// Download ...
func (d *CIFAR100) Download(ctx context.Context) error {
	if d.isDownloaded {
		return nil
	}
	workingDir := d.workingDir()
	downloadedFileName := filepath.Join(workingDir, d.fileName)
	downloadedFileName, ifDownload, err := downloadmanager.DownloadFile(d.url, downloadedFileName, downloadmanager.Context(ctx), downloadmanager.MD5Sum(d.md5sum))
	if err != nil {
		return err
	}
	if ifDownload {
		if err := downloadmanager.Unarchive(workingDir, downloadedFileName); err != nil {
			return err
		}
	}
	if err := d.move(ctx); err != nil {
		return err
	}
	archiveOutputDir := filepath.Join(workingDir, d.extractedFolderName)
	defer os.RemoveAll(archiveOutputDir)

	return nil
}

func (d *CIFAR100) move(ctx context.Context) error {
	workingDir := d.workingDir()
	archiveOutputDir := filepath.Join(workingDir, d.extractedFolderName)
	filesHashes := map[string]string{}
	for fileName, md5 := range d.trainFileNameList {
		filesHashes[fileName] = md5
	}
	for fileName, md5 := range d.testFileNameList {
		filesHashes[fileName] = md5
	}
	for fileName, md5 := range filesHashes {
		filePath := filepath.Join(archiveOutputDir, fileName)
		if !com.IsFile(filePath) {
			return errors.Errorf("the file %s for %s was not found in the extracted directory", fileName, d.CanonicalName())
		}
		newPath := filepath.Join(workingDir, fileName)
		if err := os.Rename(filePath, newPath); err != nil {
			return errors.Wrapf(err, "cannot move the file %s to %s", filePath, newPath)
		}
		ok, err := utils.MD5Sum.CheckFile(newPath, md5)
		if err != nil {
			return err
		}
		if !ok {
			return errors.Wrapf(err, "the md5 sum for %s did not match expected %s", newPath, md5)
		}
	}
	fineLabelFilePath := filepath.Join(archiveOutputDir, d.fineLabelsFileName)
	if !com.IsFile(fineLabelFilePath) {
		return errors.Errorf("the file %s for %s was not found in the extracted directory", fineLabelFilePath, d.CanonicalName())
	}
	newFineLabelFilePath := filepath.Join(workingDir, d.fineLabelsFileName)
	if err := os.Rename(fineLabelFilePath, newFineLabelFilePath); err != nil {
		return errors.Wrapf(err, "cannot move the file %s to %s", fineLabelFilePath, newFineLabelFilePath)
	}

	coarseLabelFilePath := filepath.Join(archiveOutputDir, d.coarseLabelsFileName)
	if !com.IsFile(coarseLabelFilePath) {
		return errors.Errorf("the file %s for %s was not found in the extracted directory", coarseLabelFilePath, d.CanonicalName())
	}
	newCoarseLabelFilePath := filepath.Join(workingDir, d.coarseLabelsFileName)
	if err := os.Rename(coarseLabelFilePath, newCoarseLabelFilePath); err != nil {
		return errors.Wrapf(err, "cannot move the file %s to %s", coarseLabelFilePath, newCoarseLabelFilePath)
	}

	return nil
}

// List ...
func (d *CIFAR100) List(ctx context.Context) ([]string, error) {
	if err := d.read(ctx); err != nil {
		return nil, err
	}
	keys := []string{}
	for key := range d.data {
		keys = append(keys, key)
	}
	return keys, nil
}

// Get ...
func (d *CIFAR100) Get(ctx context.Context, name string) (dldataset.LabeledData, error) {
	if err := d.read(ctx); err != nil {
		return nil, err
	}
	data, ok := d.data[name]
	if !ok {
		return nil, errors.Errorf("unable to find %s in the %s dataset", name, d.CanonicalName())
	}
	return data, nil
}

func (d *CIFAR100) Next(ctx context.Context) (dldataset.LabeledData, error) {
	return nil, errors.New("next iterator is not implemented for " + d.CanonicalName())
}

func (d *CIFAR100) read(ctx context.Context) error {
	if err := d.readLabels(ctx); err != nil {
		return err
	}
	if err := d.readData(ctx); err != nil {
		return err
	}
	return nil
}

func (d *CIFAR100) readData(ctx context.Context) error {
	if len(d.data) != 0 {
		return nil
	}

	workingDir := d.workingDir()
	data := map[string]CIFAR100LabeledImage{}

	read := func(offset int, class, fileName string) (int, error) {
		idx := offset
		filePath := filepath.Join(workingDir, fileName)
		f, err := os.Open(filePath)
		if err != nil {
			return idx, errors.Wrapf(err, "failed to open %s while performing md5 checksum", filePath)
		}
		defer f.Close()

		for {
			entry, err := d.readEntry(ctx, f)
			if err == io.EOF {
				return 0, nil
			}
			if err != nil {
				return idx, errors.Wrapf(err, "failed reading entry for %s", filePath)
			}
			data[class+"/"+strconv.Itoa(idx)] = *entry
			idx++
		}
		return idx, nil
	}
	idx := 0
	for fileName := range d.trainFileNameList {
		newIdx, err := read(idx, "train", fileName)
		if err != nil {
			return err
		}
		idx = newIdx
	}
	idx = 0
	for fileName := range d.trainFileNameList {
		newIdx, err := read(idx, "test", fileName)
		if err != nil {
			return err
		}
		idx = newIdx
	}

	d.data = data

	return nil
}

func (d *CIFAR100) readEntry(ctx context.Context, reader io.Reader) (*CIFAR100LabeledImage, error) {
	var coarseLabelIdx int8
	coarseLabelByteSize := int64(d.coarseLabelByteSize)
	coarseLabelBytesReader := io.LimitReader(reader, coarseLabelByteSize)
	err := binary.Read(coarseLabelBytesReader, binary.LittleEndian, &coarseLabelIdx)
	if err == io.EOF {
		return nil, err
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to read fine label")
	}
	if int(coarseLabelIdx) >= len(d.coarseLabels) {
		return nil, errors.Errorf("the coarse label %v is out of range of %v", coarseLabelIdx, len(d.coarseLabels))
	}

	var fineLabelIdx int8
	fineLabelByteSize := int64(d.fineLabelByteSize)
	fineLabelBytesReader := io.LimitReader(reader, fineLabelByteSize)
	err = binary.Read(fineLabelBytesReader, binary.LittleEndian, &fineLabelIdx)
	if err == io.EOF {
		return nil, err
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to read fine label")
	}
	if int(fineLabelIdx) >= len(d.fineLabels) {
		return nil, errors.Errorf("the fine label %v is out of range of %v", fineLabelIdx, len(d.fineLabels))
	}

	pixelByteSize := int64(d.pixelByteSize)
	pixelBytesReader := io.LimitReader(reader, pixelByteSize)

	img := types.NewRGBImage(image.Rect(0, 0, d.imageDimensions[0], d.imageDimensions[1]))

	err = binary.Read(pixelBytesReader, binary.LittleEndian, img.Pix)
	if err == io.EOF {
		return nil, err
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to read label")
	}

	return &CIFAR100LabeledImage{
		coarseLabel: d.coarseLabels[coarseLabelIdx],
		fineLabel:   d.fineLabels[fineLabelIdx],
		data:        img,
	}, nil
}

func (d *CIFAR100) readLabels(ctx context.Context) error {
	if len(d.fineLabels) != 0 {
		return nil
	}

	readLabelsFor := func(fileName string) ([]string, error) {
		workingDir := d.workingDir()
		labelFilePath := filepath.Join(workingDir, fileName)
		if !com.IsFile(labelFilePath) {
			return nil, errors.Errorf("the label file %s was not found", labelFilePath)
		}

		var labels []string
		f, err := os.Open(labelFilePath)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read %s", labelFilePath)
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			labels = append(labels, line)
		}
		return labels, nil
	}

	fineLabels, err := readLabelsFor(d.fineLabelsFileName)
	if err != nil {
		return errors.Wrap(err, "unable to read fine labels")
	}

	coarseLabels, err := readLabelsFor(d.coarseLabelsFileName)
	if err != nil {
		return errors.Wrap(err, "unable to read coarse labels")
	}

	d.fineLabels = fineLabels
	d.coarseLabels = coarseLabels

	return nil
}

// Close ...
func (d *CIFAR100) Close() error {
	return nil
}

func (d *CIFAR100) workingDir() string {
	category := strings.ToLower(d.Category())
	name := strings.ToLower(d.Name())
	return filepath.Join(d.baseWorkingDir, category, name)
}

func init() {
	config.AfterInit(func() {
		cifar100 = &CIFAR100{
			base: base{
				ctx:            context.Background(),
				baseWorkingDir: filepath.Join(dldataset.Config.WorkingDirectory, "dldataset"),
			},
			url:                 "https://www.cs.toronto.edu/~kriz/cifar-100-binary.tar.gz",
			fileName:            "cifar-100-binary.tar.gz",
			extractedFolderName: "cifar-100-binary",
			md5sum:              "03b5dce01913d631647c71ecec9e9cb8",
			trainFileNameList: map[string]string{
				"train.bin": "6172c7755cfe09b2fe270c85cebc1b15",
			},
			testFileNameList: map[string]string{
				"test.bin": "4499cfba6c016c1be1438163640a0898",
			},
			fineLabelsFileName:   "fine_label_names.txt",
			coarseLabelsFileName: "coarse_label_names.txt",
			imageDimensions:      []int{32, 32, 3},
			fineLabelByteSize:    1,
			coarseLabelByteSize:  1,
			pixelByteSize:        3072,
			isDownloaded:         false,
		}
		dldataset.Register(cifar100)
	})
}
