package dldataset

import (
	"github.com/k0kubun/pp"
	"github.com/c3sr/config"
	"github.com/c3sr/vipertags"
)

type dldatasetConfig struct {
	WorkingDirectory string        `json:"working_directory" config:"dldataset.working_directory" default:""`
	done             chan struct{} `json:"-" config:"-"`
}

// Config ...
var (
	// Config holds the data read by c3sr/config
	Config = &dldatasetConfig{
		done: make(chan struct{}),
	}
)

// ConfigName ...
func (dldatasetConfig) ConfigName() string {
	return "DLDataset"
}

// SetDefaults ...
func (c *dldatasetConfig) SetDefaults() {
	vipertags.SetDefaults(c)
}

// Read ...
func (c *dldatasetConfig) Read() {
	defer close(c.done)
	config.App.Wait()
	vipertags.Fill(c)
	if c.WorkingDirectory == "" || c.WorkingDirectory == "default" {
		c.WorkingDirectory = config.App.TempDir
	}
}

// Wait ...
func (c dldatasetConfig) Wait() {
	<-c.done
}

// String ...
func (c dldatasetConfig) String() string {
	return pp.Sprintln(c)
}

// Debug ...
func (c dldatasetConfig) Debug() {
	log.Debug("DLDataset Config = ", c)
}

func init() {
	config.Register(Config)
}
