package compressor

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestSevenZip_options(t *testing.T) {
	v := viper.New()
	base := newBase(config.ModelConfig{
		CompressWith: config.SubConfig{
			Type:  "7z",
			Name:  "7z",
			Viper: v,
		},
	})

	sz := &SevenZip{base}
	opts := sz.options()
	assert.Equal(t, opts[0], "a")
	assert.Equal(t, len(opts), 1)
}

func TestSevenZip_optionsWithPassword(t *testing.T) {
	v := viper.New()
	v.Set("password", "test123")
	base := newBase(config.ModelConfig{
		CompressWith: config.SubConfig{
			Type:  "7z",
			Name:  "7z",
			Viper: v,
		},
	})

	sz := &SevenZip{base}
	opts := sz.options()
	assert.Equal(t, opts[0], "a")
	assert.Equal(t, opts[1], "-ptest123")
	assert.Equal(t, opts[2], "-mhe=on")
}

func TestSevenZip_optionsWithCompressionLevel(t *testing.T) {
	v := viper.New()
	v.Set("compression_level", 9)
	base := newBase(config.ModelConfig{
		CompressWith: config.SubConfig{
			Type:  "7z",
			Name:  "7z",
			Viper: v,
		},
	})

	sz := &SevenZip{base}
	opts := sz.options()
	assert.Equal(t, opts[0], "a")
	assert.Equal(t, opts[1], "-mx=9")
}

func TestSevenZip_optionsWithArgs(t *testing.T) {
	v := viper.New()
	v.Set("args", "-mmt=4")
	base := newBase(config.ModelConfig{
		CompressWith: config.SubConfig{
			Type:  "7z",
			Name:  "7z",
			Viper: v,
		},
	})

	sz := &SevenZip{base}
	opts := sz.options()
	assert.Equal(t, opts[0], "a")
	assert.Equal(t, opts[1], "-mmt=4")
}

func TestSevenZip_optionsWithAllOptions(t *testing.T) {
	v := viper.New()
	v.Set("password", "secret")
	v.Set("compression_level", 5)
	v.Set("args", "-mmt=4")
	base := newBase(config.ModelConfig{
		CompressWith: config.SubConfig{
			Type:  "7z",
			Name:  "7z",
			Viper: v,
		},
	})

	sz := &SevenZip{base}
	opts := sz.options()
	assert.Equal(t, opts[0], "a")
	assert.Equal(t, opts[1], "-psecret")
	assert.Equal(t, opts[2], "-mhe=on")
	assert.Equal(t, opts[3], "-mx=5")
	assert.Equal(t, opts[4], "-mmt=4")
}
