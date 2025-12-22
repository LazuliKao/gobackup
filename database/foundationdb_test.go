package database

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestFoundationDB_build(t *testing.T) {
	base := newBase(config.ModelConfig{
		DumpPath: "/tmp/gobackup",
	}, config.SubConfig{
		Name: "test-foundationdb",
		Type: "foundationdb",
	})

	// Test basic configuration
	v := viper.New()
	v.Set("cluster_file", "/etc/foundationdb/fdb.cluster")
	v.Set("tag", "daily")
	v.Set("backup_url", "file:///backup/fdb")

	db := &FoundationDB{
		Base: base,
	}
	db.viper = v
	db.init()

	opts := db.build()
	assert.Contains(t, opts, "-C /etc/foundationdb/fdb.cluster")
	assert.Contains(t, opts, "-t daily")
	assert.Contains(t, opts, "-d file:///backup/fdb")
	assert.Contains(t, opts, "-w")
}

func TestFoundationDB_buildContinuous(t *testing.T) {
	base := newBase(config.ModelConfig{
		DumpPath: "/tmp/gobackup",
	}, config.SubConfig{
		Name: "test-foundationdb",
		Type: "foundationdb",
	})

	v := viper.New()
	v.Set("cluster_file", "/etc/foundationdb/fdb.cluster")
	v.Set("tag", "continuous")
	v.Set("continuous", true)
	v.Set("snapshot_interval", 3600)
	v.Set("backup_url", "file:///backup/fdb")

	db := &FoundationDB{
		Base: base,
	}
	db.viper = v
	db.init()

	opts := db.build()
	assert.Contains(t, opts, "-C /etc/foundationdb/fdb.cluster")
	assert.Contains(t, opts, "-t continuous")
	assert.Contains(t, opts, "-d file:///backup/fdb")
	assert.Contains(t, opts, "-z")
	assert.Contains(t, opts, "-s 3600")
	assert.NotContains(t, opts, "-w")
}

func TestFoundationDB_buildWithKeyRanges(t *testing.T) {
	base := newBase(config.ModelConfig{
		DumpPath: "/tmp/gobackup",
	}, config.SubConfig{
		Name: "test-foundationdb",
		Type: "foundationdb",
	})

	v := viper.New()
	v.Set("cluster_file", "/etc/foundationdb/fdb.cluster")
	v.Set("tag", "partial")
	v.Set("backup_url", "file:///backup/fdb")
	v.Set("key_ranges", []string{"apple banana", "mango pineapple"})

	db := &FoundationDB{
		Base: base,
	}
	db.viper = v
	db.init()

	opts := db.build()
	assert.Contains(t, opts, "-C /etc/foundationdb/fdb.cluster")
	assert.Contains(t, opts, "-t partial")
	assert.Contains(t, opts, "-d file:///backup/fdb")
	assert.Contains(t, opts, "-k 'apple banana'")
	assert.Contains(t, opts, "-k 'mango pineapple'")
}

func TestFoundationDB_buildWithPartitionedLog(t *testing.T) {
	base := newBase(config.ModelConfig{
		DumpPath: "/tmp/gobackup",
	}, config.SubConfig{
		Name: "test-foundationdb",
		Type: "foundationdb",
	})

	v := viper.New()
	v.Set("cluster_file", "/etc/foundationdb/fdb.cluster")
	v.Set("partitioned_log", true)
	v.Set("backup_url", "file:///backup/fdb")

	db := &FoundationDB{
		Base: base,
	}
	db.viper = v
	db.init()

	opts := db.build()
	assert.Contains(t, opts, "--partitioned-log-experimental")
}

func TestFoundationDB_buildWithBlobStore(t *testing.T) {
	base := newBase(config.ModelConfig{
		DumpPath: "/tmp/gobackup",
	}, config.SubConfig{
		Name: "test-foundationdb",
		Type: "foundationdb",
	})

	v := viper.New()
	v.Set("cluster_file", "/etc/foundationdb/fdb.cluster")
	v.Set("tag", "s3-backup")
	v.Set("backup_url", "blobstore://key:secret@s3.amazonaws.com/backup?bucket=my-backup&region=us-west-2")
	v.Set("blob_credentials", "/etc/foundationdb/blob_credentials.json")

	db := &FoundationDB{
		Base: base,
	}
	db.viper = v
	db.init()

	opts := db.build()
	assert.Contains(t, opts, "-C /etc/foundationdb/fdb.cluster")
	assert.Contains(t, opts, "-t s3-backup")
	assert.Contains(t, opts, "-d blobstore://key:secret@s3.amazonaws.com/backup?bucket=my-backup&region=us-west-2")
	assert.Contains(t, opts, "--blob-credentials /etc/foundationdb/blob_credentials.json")
}

func TestFoundationDB_defaultBackupURL(t *testing.T) {
	base := newBase(config.ModelConfig{
		DumpPath: "/tmp/gobackup",
	}, config.SubConfig{
		Name: "test-foundationdb",
		Type: "foundationdb",
	})

	v := viper.New()
	v.Set("cluster_file", "/etc/foundationdb/fdb.cluster")
	// Don't set backup_url - it should use default

	db := &FoundationDB{
		Base: base,
	}
	db.viper = v
	db.init()

	// Should default to file://<dumpPath>
	assert.Contains(t, db._backupURL, "file://")
	assert.Contains(t, db._backupURL, "/tmp/gobackup")
	assert.Contains(t, db._backupURL, "foundationdb")
}

func TestFoundationDB_minimalConfig(t *testing.T) {
	base := newBase(config.ModelConfig{
		DumpPath: "/tmp/gobackup",
	}, config.SubConfig{
		Name: "fdb",
		Type: "foundationdb",
	})

	// Minimal config - only cluster_file is set
	v := viper.New()
	v.Set("cluster_file", "/etc/foundationdb/fdb.cluster")

	db := &FoundationDB{
		Base: base,
	}
	db.viper = v
	err := db.init()

	assert.Nil(t, err)
	assert.Equal(t, "/etc/foundationdb/fdb.cluster", db.clusterFile)
	assert.Equal(t, "default", db.tag)
	assert.Equal(t, false, db.continuous)
	assert.Contains(t, db._backupURL, "file://")
	
	// Build should work with minimal config
	opts := db.build()
	assert.Contains(t, opts, "-C /etc/foundationdb/fdb.cluster")
	assert.Contains(t, opts, "-t default")
	assert.Contains(t, opts, "-w")
}
