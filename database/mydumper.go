package database

import (
	"fmt"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// MyDumper database
//
// type: mydumper
// host: 127.0.0.1
// port: 3306
// socket:
// database:
// username: root
// password:
// tables:
// threads: 4
// compress: false
// args:
// Note: For table exclusion, use the 'args' option with mydumper's regex patterns
type MyDumper struct {
	Base
	host     string
	port     string
	socket   string
	database string
	username string
	password string
	tables   []string
	threads  int
	compress bool
	args     string
}

func (db *MyDumper) init() (err error) {
	viper := db.viper
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("username", "root")
	viper.SetDefault("port", 3306)
	viper.SetDefault("threads", 4)
	viper.SetDefault("compress", false)

	db.host = viper.GetString("host")
	db.port = viper.GetString("port")
	db.socket = viper.GetString("socket")
	db.database = viper.GetString("database")
	db.username = viper.GetString("username")
	db.password = viper.GetString("password")

	db.tables = viper.GetStringSlice("tables")
	db.threads = viper.GetInt("threads")
	db.compress = viper.GetBool("compress")

	if len(viper.GetString("args")) > 0 {
		db.args = viper.GetString("args")
	}

	// socket
	if len(db.socket) != 0 {
		db.host = ""
		db.port = ""
	}

	return nil
}

func (db *MyDumper) build() string {
	dumpArgs := []string{}
	if len(db.host) > 0 {
		dumpArgs = append(dumpArgs, "--host", db.host)
	}
	if len(db.port) > 0 {
		dumpArgs = append(dumpArgs, "--port", db.port)
	}
	if len(db.socket) > 0 {
		dumpArgs = append(dumpArgs, "--socket", db.socket)
	}
	if len(db.username) > 0 {
		dumpArgs = append(dumpArgs, "--user", db.username)
	}
	if len(db.password) > 0 {
		dumpArgs = append(dumpArgs, "--password", db.password)
	}

	if len(db.database) > 0 {
		dumpArgs = append(dumpArgs, "--database", db.database)
	}

	for _, table := range db.tables {
		dumpArgs = append(dumpArgs, "--tables-list", table)
	}

	// Note: mydumper uses --regex as an inclusion pattern.
	// For excluding tables, use the 'regex' config option directly with proper PCRE syntax
	// or use the 'args' config option to pass --ignore-table

	if db.threads > 0 {
		dumpArgs = append(dumpArgs, "--threads", fmt.Sprintf("%d", db.threads))
	}

	if db.compress {
		dumpArgs = append(dumpArgs, "--compress")
	}

	if len(db.args) > 0 {
		dumpArgs = append(dumpArgs, db.args)
	}

	dumpArgs = append(dumpArgs, "--outputdir", db.dumpPath)

	return "mydumper" + " " + strings.Join(dumpArgs, " ")
}

func (db *MyDumper) perform() error {
	logger := logger.Tag("MyDumper")

	logger.Info("-> Dumping MySQL with mydumper...")
	_, err := helper.Exec(db.build())
	if err != nil {
		logger.Errorf("-> Dump error: %s", err)
		return err
	}
	logger.Info("dump path:", db.dumpPath)
	return nil
}
