package encryptor

import (
	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/logger"
	"github.com/spf13/viper"
)

// Base encryptor
type Base struct {
	model       config.ModelConfig
	viper       *viper.Viper
	archivePath string
}

// Encryptor interface
type Encryptor interface {
	perform() (encryptPath string, err error)
}

func newBase(archivePath string, model config.ModelConfig) (base *Base) {
	base = &Base{
		archivePath: archivePath,
		model:       model,
		viper:       model.EncryptWith.Viper,
	}
	return
}

// Run encryptor on multiple archive paths
func Run(archivePaths []string, model config.ModelConfig) (encryptPaths []string, err error) {
	logger := logger.Tag("Encryptor")

	// If no encryption configured, return paths as-is
	if model.EncryptWith.Type == "" {
		return archivePaths, nil
	}

	logger.Info("encrypt | " + model.EncryptWith.Type)

	encryptPaths = make([]string, 0, len(archivePaths))
	for _, archivePath := range archivePaths {
		base := newBase(archivePath, model)
		var enc Encryptor
		switch model.EncryptWith.Type {
		case "openssl":
			enc = NewOpenSSL(base)
		default:
			encryptPaths = append(encryptPaths, archivePath)
			continue
		}

		encryptPath, err := enc.perform()
		if err != nil {
			return nil, err
		}
		logger.Info("encrypted:", encryptPath)
		encryptPaths = append(encryptPaths, encryptPath)
	}

	// save Extension
	model.Viper.Set("Ext", model.Viper.GetString("Ext")+".enc")

	return encryptPaths, nil
}
