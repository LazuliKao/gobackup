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

// Run encryptor on archive path
func Run(archivePath string, model config.ModelConfig) (string, error) {
	logger := logger.Tag("Encryptor")

	// If no encryption configured, return path as-is
	if model.EncryptWith.Type == "" {
		return archivePath, nil
	}

	logger.Info("encrypt | " + model.EncryptWith.Type)

	base := newBase(archivePath, model)
	var enc Encryptor
	switch model.EncryptWith.Type {
	case "openssl":
		enc = NewOpenSSL(base)
	default:
		return archivePath, nil
	}

	encryptPath, err := enc.perform()
	if err != nil {
		return "", err
	}
	logger.Info("encrypted:", encryptPath)

	// save Extension
	model.Viper.Set("Ext", model.Viper.GetString("Ext")+".enc")

	return encryptPath, nil
}
