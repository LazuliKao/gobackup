package compressor

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
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

func (sz *SevenZip) perform() (archivePath string, err error) {
	filePath := sz.archiveFilePath(sz.ext)

	opts := sz.options()
	opts = append(opts, filePath)
	opts = append(opts, sz.name)

	_, err = helper.Exec("7z", opts...)
	if err != nil {
		return "", err
	}

	// When volume_size is set, 7z creates multiple volume files like: archive.7z.001, archive.7z.002, etc.
	// We need to collect all these volume files
	if sz.HasVolumeSize() {
		archivePath, err = sz.collectVolumeFiles(filePath)
		if err != nil {
			return "", err
		}
	} else {
		archivePath = filePath
	}

	return archivePath, nil
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
// If only one volume file exists (.001), it renames it to remove the .001 suffix.
// If multiple volume files exist, it moves them into a date-named directory (similar to splitter).
func (sz *SevenZip) collectVolumeFiles(basePath string) (string, error) {
	logger := logger.Tag("Compressor")

	// 7z volume files follow the pattern: basePath.001, basePath.002, etc.
	pattern := basePath + ".*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to find volume files: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no volume files found matching pattern: %s", pattern)
	}

	// Sort to ensure consistent ordering (001, 002, 003, ...)
	sort.Strings(matches)

	// If only one volume file (.001), rename it to remove the suffix
	if len(matches) == 1 {
		oldPath := matches[0]
		// Remove .001 suffix: archive.7z.001 -> archive.7z
		newPath := strings.TrimSuffix(oldPath, ".001")
		if err := os.Rename(oldPath, newPath); err != nil {
			return "", fmt.Errorf("failed to rename single volume file: %w", err)
		}
		logger.Infof("Renamed single volume %s -> %s", filepath.Base(oldPath), filepath.Base(newPath))
		return newPath, nil
	}

	// Multiple volume files: move them into a date-named directory (similar to splitter)
	// basePath: /tmp/gobackup.../2022.12.04.07.24.08.7z
	// archiveDirPath: /tmp/gobackup.../2022.12.04.07.24.08
	archiveDirPath := strings.TrimSuffix(basePath, sz.ext)
	if err := helper.MkdirP(archiveDirPath); err != nil {
		return "", fmt.Errorf("failed to create archive directory: %w", err)
	}

	// Move all volume files into the directory
	for _, oldPath := range matches {
		fileName := filepath.Base(oldPath)
		newPath := filepath.Join(archiveDirPath, fileName)
		if err := os.Rename(oldPath, newPath); err != nil {
			return "", fmt.Errorf("failed to move volume file %s: %w", fileName, err)
		}
	}

	logger.Infof("Moved %d volume files into %s", len(matches), filepath.Base(archiveDirPath))

	// Return the directory path containing volume files
	// The storage layer will handle reading files from this directory
	return archiveDirPath, nil
}
