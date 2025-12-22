package database

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// FoundationDB database
//
// ref:
// https://apple.github.io/foundationdb/backups.html
//
// # Keys
//
//   - type: foundationdb
//   - cluster_file: /etc/foundationdb/fdb.cluster
//   - tag: default
//   - continuous: false
//   - snapshot_interval: 864000
//   - partitioned_log: false
//   - key_ranges: []
//   - blob_credentials:
//   - args:
type FoundationDB struct {
	Base
	clusterFile       string
	tag               string
	continuous        bool
	snapshotInterval  int
	partitionedLog    bool
	keyRanges         []string
	blobCredentials   string
	args              string
	_backupURL        string
	_dumpFilePath     string
}

func (db *FoundationDB) init() (err error) {
	viper := db.viper
	viper.SetDefault("cluster_file", "/etc/foundationdb/fdb.cluster")
	viper.SetDefault("tag", "default")
	viper.SetDefault("continuous", false)
	viper.SetDefault("snapshot_interval", 864000)
	viper.SetDefault("partitioned_log", false)

	db.clusterFile = viper.GetString("cluster_file")
	db.tag = viper.GetString("tag")
	db.continuous = viper.GetBool("continuous")
	db.snapshotInterval = viper.GetInt("snapshot_interval")
	db.partitionedLog = viper.GetBool("partitioned_log")
	db.keyRanges = viper.GetStringSlice("key_ranges")
	db.blobCredentials = viper.GetString("blob_credentials")
	db.args = viper.GetString("args")

	// backup_url is optional
	// If not specified, use local file path for integration with GoBackup's archive/compressor
	backupURL := viper.GetString("backup_url")
	if len(backupURL) == 0 {
		// Use dumpPath directly - fdbbackup will create timestamped subdirectory
		db._backupURL = fmt.Sprintf("file://%s", db.dumpPath)
	} else {
		db._backupURL = backupURL
	}

	db._dumpFilePath = path.Join(db.dumpPath, "backup_info.txt")

	return nil
}

func (db *FoundationDB) build() string {
	// fdbbackup start command
	var args []string

	// Add cluster file
	if len(db.clusterFile) > 0 {
		args = append(args, "-C", db.clusterFile)
	}

	// Add tag
	if len(db.tag) > 0 {
		args = append(args, "-t", db.tag)
	}

	// Add destination backup URL
	args = append(args, "-d", db._backupURL)

	// Add continuous mode
	if db.continuous {
		args = append(args, "-z")
	}

	// Add snapshot interval
	if db.snapshotInterval > 0 {
		args = append(args, "-s", fmt.Sprintf("%d", db.snapshotInterval))
	}

	// Add partitioned log
	if db.partitionedLog {
		args = append(args, "--partitioned-log-experimental")
	}

	// Add key ranges
	for _, keyRange := range db.keyRanges {
		args = append(args, "-k", fmt.Sprintf("'%s'", keyRange))
	}

	// Add blob credentials file
	if len(db.blobCredentials) > 0 {
		args = append(args, "--blob-credentials", db.blobCredentials)
	}

	// Wait for backup to complete (unless continuous mode)
	// The -w flag makes fdbbackup wait until the backup is restorable before returning
	// This ensures the backup is fully completed before proceeding to next steps
	if !db.continuous {
		args = append(args, "-w")
	}

	// Add additional args
	if len(db.args) > 0 {
		args = append(args, db.args)
	}

	return strings.Join(args, " ")
}

func (db *FoundationDB) perform() (err error) {
	logger := logger.Tag("Database")

	opts := db.build()
	logger.Info("-> Dumping FoundationDB...")

	// Start the backup
	// With -w flag (in non-continuous mode), this will block until backup is complete and restorable
	cmd := fmt.Sprintf("fdbbackup start %s", opts)
	logger.Info("Executing:", cmd)
	
	output, err := helper.Exec(cmd)
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}

	if len(output) > 0 {
		logger.Info("Backup output:", output)
	}

	// For non-continuous backups, verify the backup was created
	if !db.continuous {
		// Check if dumpPath exists and has content
		if !helper.IsExistsPath(db.dumpPath) {
			return fmt.Errorf("backup directory not found: %s", db.dumpPath)
		}

		// Save backup information file in dumpPath
		// This file will be included in archive/compressor processing
		backupInfo := fmt.Sprintf("FoundationDB backup completed\nBackup URL: %s\nTag: %s\nCluster File: %s\nTimestamp: %s\n",
			db._backupURL, db.tag, db.clusterFile, time.Now().Format(time.RFC3339))
		err = os.WriteFile(db._dumpFilePath, []byte(backupInfo), 0644)
		if err != nil {
			logger.Warn("Failed to write backup info file:", err)
		}

		logger.Info("dump path:", db.dumpPath)
		logger.Info("Backup completed successfully")
	} else {
		logger.Info("Continuous backup started, running in background")
	}

	return nil
}

// Status returns the current backup status
func (db *FoundationDB) Status() (string, error) {
	var args []string

	if len(db.clusterFile) > 0 {
		args = append(args, "-C", db.clusterFile)
	}

	if len(db.tag) > 0 {
		args = append(args, "-t", db.tag)
	}

	cmd := fmt.Sprintf("fdbbackup status %s", strings.Join(args, " "))
	output, err := helper.Exec(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to get backup status: %s", err)
	}

	return output, nil
}

// Abort aborts the current backup
func (db *FoundationDB) Abort() error {
	var args []string

	if len(db.clusterFile) > 0 {
		args = append(args, "-C", db.clusterFile)
	}

	if len(db.tag) > 0 {
		args = append(args, "-t", db.tag)
	}

	cmd := fmt.Sprintf("fdbbackup abort %s", strings.Join(args, " "))
	_, err := helper.Exec(cmd)
	if err != nil {
		return fmt.Errorf("failed to abort backup: %s", err)
	}

	return nil
}

// Discontinue discontinues a continuous backup
func (db *FoundationDB) Discontinue() error {
	if !db.continuous {
		return fmt.Errorf("backup is not in continuous mode")
	}

	var args []string

	if len(db.clusterFile) > 0 {
		args = append(args, "-C", db.clusterFile)
	}

	if len(db.tag) > 0 {
		args = append(args, "-t", db.tag)
	}

	cmd := fmt.Sprintf("fdbbackup discontinue %s", strings.Join(args, " "))
	_, err := helper.Exec(cmd)
	if err != nil {
		return fmt.Errorf("failed to discontinue backup: %s", err)
	}

	return nil
}
