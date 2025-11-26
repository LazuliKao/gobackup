package database

import (
	"github.com/gobackup/gobackup/config"
	"github.com/spf13/viper"

	"testing"

	"github.com/longbridgeapp/assert"
)

func TestXtraBackup_init(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("parallel", 4)
	viper.Set("compress", true)
	viper.Set("args", "--a1 --a2 --a3")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		config.SubConfig{
			Type:  "xtrabackup",
			Name:  "xtrabackup1",
			Viper: viper,
		},
	)

	db := &XtraBackup{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	script := db.build()
	assert.Equal(t, script, "xtrabackup --backup --host=1.2.3.4 --port=1234 --user=user1 --password=pass1 --databases=my_db --parallel=4 --compress --a1 --a2 --a3 --target-dir=/data/backups/xtrabackup/xtrabackup1")
}

func TestXtraBackup_dumpArgsWithAdditionalOptions(t *testing.T) {
	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		config.SubConfig{
			Type: "xtrabackup",
			Name: "xtrabackup1",
		},
	)
	db := &XtraBackup{
		Base:     base,
		database: "dummy_test",
		host:     "127.0.0.2",
		port:     "6378",
		password: "*&^92'",
		parallel: 2,
		compress: false,
		args:     "--no-lock",
	}

	assert.Equal(t, db.build(), "xtrabackup --backup --host=127.0.0.2 --port=6378 --password=*&^92' --databases=dummy_test --parallel=2 --no-lock --target-dir=/data/backups/xtrabackup/xtrabackup1")
}
