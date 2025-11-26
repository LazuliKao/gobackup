package database

import (
	"github.com/gobackup/gobackup/config"
	"github.com/spf13/viper"

	"testing"

	"github.com/longbridgeapp/assert"
)

func TestMyDumper_init(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("threads", 8)
	viper.Set("compress", true)
	viper.Set("args", "--a1 --a2 --a3")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		config.SubConfig{
			Type:  "mydumper",
			Name:  "mydumper1",
			Viper: viper,
		},
	)

	db := &MyDumper{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	script := db.build()
	assert.Equal(t, script, "mydumper --host 1.2.3.4 --port 1234 --user user1 --password pass1 --database my_db --threads 8 --compress --a1 --a2 --a3 --outputdir /data/backups/mydumper/mydumper1")
}

func TestMyDumper_dumpArgsWithAdditionalOptions(t *testing.T) {
	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		config.SubConfig{
			Type: "mydumper",
			Name: "mydumper1",
		},
	)
	db := &MyDumper{
		Base:     base,
		database: "dummy_test",
		host:     "127.0.0.2",
		port:     "6378",
		password: "*&^92'",
		threads:  4,
		compress: false,
		args:     "--long-query-guard 300",
	}

	assert.Equal(t, db.build(), "mydumper --host 127.0.0.2 --port 6378 --password *&^92' --database dummy_test --threads 4 --long-query-guard 300 --outputdir /data/backups/mydumper/mydumper1")
}
