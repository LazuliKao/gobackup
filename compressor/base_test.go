package compressor

import (
	"path"
	"strings"
	"testing"
	"time"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

type Monkey struct {
	Base
}

func (c Monkey) perform() (archivePath string, err error) {
	result := "aaa"
	return result, nil
}

func TestBase_archiveFilePath(t *testing.T) {
    viper := viper.New()
	viper.SetDefault("type", "tar")
	viper.SetDefault("filename_format", "backup-2006.01.02.15.04.05")
    model := config.ModelConfig{}
	model.CompressWith = config.SubConfig{
		Type:  viper.GetString("type"),
		Viper: viper,
	}
	base := newBase(model)
	prefixPath := path.Join(base.model.TempPath, time.Now().Format("backup-2006.01.02.15.04"))
	archivePath := base.archiveFilePath(".tar")
	assert.True(t, strings.HasPrefix(archivePath, prefixPath))
	assert.True(t, strings.HasSuffix(archivePath, ".tar"))
}

func TestBaseInterface(t *testing.T) {
	model := config.ModelConfig{
		Name: "TestMoneky",
	}
	base := newBase(model)
	assert.Equal(t, base.name, model.Name)
	assert.Equal(t, base.model, model)

	c := Monkey{Base: base}
	result, err := c.perform()
	assert.Equal(t, result, "aaa")
	assert.Nil(t, err)
}

func TestRun_SevenZipVolumeSizeWithSplitter_ReturnsError(t *testing.T) {
	compressViper := viper.New()
	compressViper.Set("type", "7z")
	compressViper.Set("volume_size", "100m")
	compressViper.SetDefault("filename_format", "2006.01.02.15.04.05")

	splitterViper := viper.New()
	splitterViper.Set("chunk_size", "50m")

	modelViper := viper.New()

	model := config.ModelConfig{
		Name:     "test",
		TempPath: "/tmp",
		DumpPath: "/tmp/test",
		CompressWith: config.SubConfig{
			Type:  "7z",
			Name:  "7z",
			Viper: compressViper,
		},
		Splitter: splitterViper,
		Viper:    modelViper,
	}

	_, err := Run(model)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "cannot use both 7z native volume splitting"))
}
