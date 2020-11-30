package vision

import (
	"testing"

	context "context"

	"github.com/c3sr/dldataset"
	"github.com/stretchr/testify/assert"
)

// TestDownloadCIFAR10 ...
func TestDownloadCIFAR10(t *testing.T) {
	ctx := context.Background()

	cifar10, err := dldataset.Get("vision", "cifar10")
	assert.NoError(t, err)
	assert.NotEmpty(t, cifar10)

	defer cifar10.Close()

	err = cifar10.Download(ctx)
	assert.NoError(t, err)

	fileList, err := cifar10.List(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, fileList)

	lbl, err := cifar10.Get(ctx, "train/1")
	assert.NoError(t, err)
	assert.NotEmpty(t, lbl)

	lbl, err = cifar10.Get(ctx, "test/1")
	assert.NoError(t, err)
	assert.NotEmpty(t, lbl)

	// pp.Println(lbl)

}
