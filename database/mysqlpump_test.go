package database

import (
	"github.com/gobackup/gobackup/config"
	"github.com/spf13/viper"

	"testing"

	"github.com/longbridgeapp/assert"
)

func TestMySQLPump_init(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("tables", []string{"foo", "bar"})
	viper.Set("exclude_tables", []string{"aa", "bb"})
	viper.Set("parallel", 4)
	viper.Set("args", "--a1 --a2 --a3")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		config.SubConfig{
			Type:  "mysqlpump",
			Name:  "mysqlpump1",
			Viper: viper,
		},
	)

	db := &MySQLPump{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	script := db.build()
	assert.Equal(t, script, "mysqlpump --host=1.2.3.4 --port=1234 --user=user1 --password=pass1 --exclude-tables=aa --exclude-tables=bb --default-parallelism=4 --a1 --a2 --a3 --databases my_db --include-tables=foo,bar --result-file=/data/backups/mysqlpump/mysqlpump1/my_db.sql")
}

func TestMySQLPump_dumpArgsWithAdditionalOptions(t *testing.T) {
	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		config.SubConfig{
			Type: "mysqlpump",
			Name: "mysqlpump1",
		},
	)
	db := &MySQLPump{
		Base:     base,
		database: "dummy_test",
		host:     "127.0.0.2",
		port:     "6378",
		password: "*&^92'",
		parallel: 2,
		args:     "--single-transaction",
	}

	assert.Equal(t, db.build(), "mysqlpump --host=127.0.0.2 --port=6378 --password=*&^92' --default-parallelism=2 --single-transaction --databases dummy_test --result-file=/data/backups/mysqlpump/mysqlpump1/dummy_test.sql")
}
