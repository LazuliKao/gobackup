package database

import (
	"github.com/gobackup/gobackup/config"
	"github.com/spf13/viper"

	"testing"

	"github.com/longbridgeapp/assert"
)

func TestMariaDBDump_init(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("tables", []string{"foo", "bar"})
	viper.Set("exclude_tables", []string{"aa", "bb"})
	viper.Set("args", "--a1 --a2 --a3")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		config.SubConfig{
			Type:  "mariadb-dump",
			Name:  "mariadb_dump1",
			Viper: viper,
		},
	)

	db := &MariaDBDump{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	script := db.build()
	assert.Equal(t, script, "mariadb-dump --host 1.2.3.4 --port 1234 -u user1 -ppass1 --ignore-table=my_db.aa --ignore-table=my_db.bb --a1 --a2 --a3 my_db foo bar --result-file=/data/backups/mariadb-dump/mariadb_dump1/my_db.sql")
}

func TestMariaDBDump_dumpArgsWithAdditionalOptions(t *testing.T) {
	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		config.SubConfig{
			Type: "mariadb-dump",
			Name: "mariadb_dump1",
		},
	)
	db := &MariaDBDump{
		Base:     base,
		database: "dummy_test",
		host:     "127.0.0.2",
		port:     "6378",
		password: "*&^92'",
		args:     "--single-transaction --quick",
	}

	assert.Equal(t, db.build(), "mariadb-dump --host 127.0.0.2 --port 6378 -p*&^92' --single-transaction --quick dummy_test --result-file=/data/backups/mariadb-dump/mariadb_dump1/dummy_test.sql")
}
