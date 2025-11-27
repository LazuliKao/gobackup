package compressor

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/gobackup/gobackup/helper"
)

// SevenZip compressor for 7z archives with optional password encryption
//
// type: 7z
// password: (optional) password for encryption
// compression_level: (optional) 0-9, default 5
// method: (optional) compression method, e.g., LZMA, LZMA2, PPMd, BZip2, Deflate, Copy
// volume_size: (optional) split archive into volumes of specified size (e.g., "100m", "1g")
// args: (optional) additional 7z arguments
type SevenZip struct {
	Base
}

// HasVolumeSize returns true if native 7z volume splitting is enabled
func (sz *SevenZip) HasVolumeSize() bool {
	return len(sz.viper.GetString("volume_size")) > 0
}

func (sz *SevenZip) perform() (archivePaths []string, err error) {
	filePath := sz.archiveFilePath(sz.ext)

	opts := sz.options()
	opts = append(opts, filePath)
	opts = append(opts, sz.name)

	_, err = helper.Exec("7z", opts...)
	if err != nil {
		return nil, err
	}

	// When volume_size is set, 7z creates multiple volume files like: archive.7z.001, archive.7z.002, etc.
	// We need to collect all these volume files
	if sz.HasVolumeSize() {
		archivePaths, err = sz.collectVolumeFiles(filePath)
		if err != nil {
			return nil, err
		}
	} else {
		archivePaths = []string{filePath}
	}

	return archivePaths, nil
}

func (sz *SevenZip) options() (opts []string) {
	opts = append(opts, "a")

	// Get password from config
	password := sz.viper.GetString("password")
	if len(password) > 0 {
		opts = append(opts, "-p"+password)
		// Enable header encryption when password is set
		opts = append(opts, "-mhe=on")
	}

	// Get compression method (e.g., LZMA, LZMA2, PPMd, BZip2, Deflate, Copy)
	method := sz.viper.GetString("method")
	if len(method) > 0 {
		opts = append(opts, fmt.Sprintf("-m0=%s", method))
	}

	// Get compression level (0-9)
	compressionLevel := sz.viper.GetInt("compression_level")
	if sz.viper.IsSet("compression_level") && compressionLevel >= 0 && compressionLevel <= 9 {
		opts = append(opts, fmt.Sprintf("-mx=%d", compressionLevel))
	}

	// Get volume size for split archives (e.g., "100m", "1g")
	volumeSize := sz.viper.GetString("volume_size")
	if len(volumeSize) > 0 {
		opts = append(opts, fmt.Sprintf("-v%s", volumeSize))
	}

	// Get additional args
	args := sz.viper.GetString("args")
	if len(args) > 0 {
		opts = append(opts, args)
	}

	return
}

// collectVolumeFiles finds all volume files created by 7z when using volume splitting.
// 7z creates files like: archive.7z.001, archive.7z.002, etc.
func (sz *SevenZip) collectVolumeFiles(basePath string) ([]string, error) {
	// 7z volume files follow the pattern: basePath.001, basePath.002, etc.
	pattern := basePath + ".*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find volume files: %w", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no volume files found matching pattern: %s", pattern)
	}

	// Sort to ensure consistent ordering (001, 002, 003, ...)
	sort.Strings(matches)

	return matches, nil
}
