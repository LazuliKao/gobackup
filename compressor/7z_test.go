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

func TestSevenZip_optionsWithCompressionLevelZero(t *testing.T) {
	v := viper.New()
	v.Set("compression_level", 0)
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
	assert.Equal(t, opts[1], "-mx=0")
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

func TestSevenZip_optionsWithMethod(t *testing.T) {
	v := viper.New()
	v.Set("method", "LZMA2")
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
	assert.Equal(t, opts[1], "-m0=LZMA2")
}

func TestSevenZip_optionsWithVolumeSize(t *testing.T) {
	v := viper.New()
	v.Set("volume_size", "100m")
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
	assert.Equal(t, opts[1], "-v100m")
}

func TestSevenZip_optionsWithMethodAndVolumeSize(t *testing.T) {
	v := viper.New()
	v.Set("method", "PPMd")
	v.Set("volume_size", "1g")
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
	assert.Equal(t, opts[1], "-m0=PPMd")
	assert.Equal(t, opts[2], "-v1g")
}

func TestSevenZip_HasVolumeSize(t *testing.T) {
	v := viper.New()
	base := newBase(config.ModelConfig{
		CompressWith: config.SubConfig{
			Type:  "7z",
			Name:  "7z",
			Viper: v,
		},
	})

	sz := &SevenZip{base}
	assert.Equal(t, sz.HasVolumeSize(), false)

	v.Set("volume_size", "100m")
	assert.Equal(t, sz.HasVolumeSize(), true)
}

func TestSevenZip_optionsWithAllNewOptions(t *testing.T) {
	v := viper.New()
	v.Set("password", "secret")
	v.Set("method", "LZMA2")
	v.Set("compression_level", 9)
	v.Set("volume_size", "500m")
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
	assert.Equal(t, opts[3], "-m0=LZMA2")
	assert.Equal(t, opts[4], "-mx=9")
	assert.Equal(t, opts[5], "-v500m")
	assert.Equal(t, opts[6], "-mmt=4")
}
