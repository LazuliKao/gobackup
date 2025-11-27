package splitter

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
	"github.com/spf13/viper"
)

// Run splitter on archive path
func Run(archivePath string, model config.ModelConfig) (string, error) {
	logger := logger.Tag("Splitter")

	splitter := model.Splitter
	if splitter == nil {
		// No splitting configured, return path as-is
		return archivePath, nil
	}

	logger.Info("Split to chunks")

	splitter.SetDefault("suffix_length", 3)
	splitter.SetDefault("numeric_suffixes", true)
	if len(splitter.GetString("chunk_size")) == 0 {
		return "", fmt.Errorf("chunk_size option is required")
	}

	ext := model.Viper.GetString("Ext")
	// /tmp/gobackup3755903383/1670167448676759530/2022.12.04.07.24.08
	archiveDirPath := strings.TrimSuffix(archivePath, ext)
	if err := helper.MkdirP(archiveDirPath); err != nil {
		return "", err
	}
	// /tmp/gobackup3755903383/1670167448676759530/2022.12.04.07.24.08/2022.12.04.07.24.08.tar.xz-
	splitSuffix := filepath.Join(archiveDirPath, filepath.Base(archivePath)+"-")

	opts := options(splitter)
	opts = append(opts, archivePath, splitSuffix)
	_, err := helper.Exec("split", opts...)
	if err != nil {
		return "", err
	}

	logger.Info("Split done")

	err = os.Remove(archivePath)
	if err != nil {
		return "", err
	}

	// Return the directory path containing split files
	// The storage layer will handle reading files from this directory
	return archiveDirPath, nil
}

func options(splitter *viper.Viper) (opts []string) {
	bytes := splitter.GetString("chunk_size")
	opts = append(opts, "-b", bytes)
	suffixLength := splitter.GetInt("suffix_length")
	opts = append(opts, "-a", strconv.Itoa(suffixLength))
	if splitter.GetBool("numeric_suffixes") {
		opts = append(opts, "--numeric-suffixes")
	}

	return
}
