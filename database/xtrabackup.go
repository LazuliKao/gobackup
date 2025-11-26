package database

import (
	"fmt"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// XtraBackup database
//
// type: xtrabackup
// host: 127.0.0.1
// port: 3306
// socket:
// database:
// username: root
// password:
// parallel: 1
// compress: false
// args:
type XtraBackup struct {
	Base
	host     string
	port     string
	socket   string
	database string
	username string
	password string
	parallel int
	compress bool
	args     string
}

func (db *XtraBackup) init() (err error) {
	viper := db.viper
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("username", "root")
	viper.SetDefault("port", 3306)
	viper.SetDefault("parallel", 1)
	viper.SetDefault("compress", false)

	db.host = viper.GetString("host")
	db.port = viper.GetString("port")
	db.socket = viper.GetString("socket")
	db.database = viper.GetString("database")
	db.username = viper.GetString("username")
	db.password = viper.GetString("password")
	db.parallel = viper.GetInt("parallel")
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

func (db *XtraBackup) build() string {
	dumpArgs := []string{"--backup"}
	if len(db.host) > 0 {
		dumpArgs = append(dumpArgs, "--host="+db.host)
	}
	if len(db.port) > 0 {
		dumpArgs = append(dumpArgs, "--port="+db.port)
	}
	if len(db.socket) > 0 {
		dumpArgs = append(dumpArgs, "--socket="+db.socket)
	}
	if len(db.username) > 0 {
		dumpArgs = append(dumpArgs, "--user="+db.username)
	}
	if len(db.password) > 0 {
		dumpArgs = append(dumpArgs, "--password="+db.password)
	}

	if len(db.database) > 0 {
		dumpArgs = append(dumpArgs, "--databases="+db.database)
	}

	if db.parallel > 0 {
		dumpArgs = append(dumpArgs, fmt.Sprintf("--parallel=%d", db.parallel))
	}

	if db.compress {
		dumpArgs = append(dumpArgs, "--compress")
	}

	if len(db.args) > 0 {
		dumpArgs = append(dumpArgs, db.args)
	}

	dumpArgs = append(dumpArgs, "--target-dir="+db.dumpPath)

	return "xtrabackup" + " " + strings.Join(dumpArgs, " ")
}

func (db *XtraBackup) perform() error {
	logger := logger.Tag("XtraBackup")

	logger.Info("-> Dumping MySQL with xtrabackup...")
	_, err := helper.Exec(db.build())
	if err != nil {
		logger.Errorf("-> Dump error: %s", err)
		return err
	}
	logger.Info("dump path:", db.dumpPath)
	return nil
}
