package compressor

import (
	"fmt"

	"github.com/gobackup/gobackup/helper"
)

// SevenZip compressor for 7z archives with optional password encryption
//
// type: 7z
// password: (optional) password for encryption
// compression_level: (optional) 0-9, default 5
// args: (optional) additional 7z arguments
type SevenZip struct {
	Base
}

func (sz *SevenZip) perform() (archivePath string, err error) {
	filePath := sz.archiveFilePath(sz.ext)

	opts := sz.options()
	opts = append(opts, filePath)
	opts = append(opts, sz.name)
	archivePath = filePath

	_, err = helper.Exec("7z", opts...)

	return
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

	// Get compression level (0-9)
	compressionLevel := sz.viper.GetInt("compression_level")
	if compressionLevel > 0 && compressionLevel <= 9 {
		opts = append(opts, fmt.Sprintf("-mx=%d", compressionLevel))
	}

	// Get additional args
	args := sz.viper.GetString("args")
	if len(args) > 0 {
		opts = append(opts, args)
	}

	return
}
