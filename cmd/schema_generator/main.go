package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
)

type ConfigSchema struct {
	WorkDir string                 `json:"workdir,omitempty" jsonschema:"title=WorkDir,description=Base working directory for temporary backup files."`
	Web     WebConfig              `json:"web,omitempty" jsonschema:"title=WebConfig,description=Web UI and API server configuration."`
	Models  map[string]ModelConfig `json:"models" jsonschema:"title=Models,description=Backup models keyed by model name."`
}

type WebConfig struct {
	Host     string `json:"host,omitempty" jsonschema:"title=Host,description=Web server bind host."`
	Port     string `json:"port,omitempty" jsonschema:"title=Port,description=Web server port."`
	Username string `json:"username,omitempty" jsonschema:"title=Username,description=Web UI username."`
	Password string `json:"password,omitempty" jsonschema:"title=Password,description=Web UI password."`
}

type ScheduleConfig struct {
	Cron  string `json:"cron,omitempty" jsonschema:"title=Cron,description=Cron expression for scheduled backups."`
	Every string `json:"every,omitempty" jsonschema:"title=Every,description=Interval expression such as 1day or 12h."`
	At    string `json:"at,omitempty" jsonschema:"title=At,description=Time of day used together with every."`
}

type ModelConfig struct {
	Name           string                         `json:"name,omitempty" jsonschema:"title=Name,description=Model name."`
	Description    string                         `json:"description,omitempty" jsonschema:"title=Description,description=Human readable description for the backup model."`
	Schedule       ScheduleConfig                 `json:"schedule,omitempty" jsonschema:"title=Schedule,description=Backup schedule configuration."`
	CompressWith   CompressSubConfig              `json:"compress_with,omitempty" jsonschema:"title=CompressWith,description=Compression configuration."`
	EncryptWith    EncryptSubConfig               `json:"encrypt_with,omitempty" jsonschema:"title=EncryptWith,description=Encryption configuration."`
	Archive        map[string]any                 `json:"archive,omitempty" jsonschema:"title=Archive,description=Archive configuration."`
	Splitter       map[string]any                 `json:"split_with,omitempty" jsonschema:"title=Splitter,description=Split output configuration."`
	Databases      map[string]DatabaseSubConfig   `json:"databases,omitempty" jsonschema:"title=Databases,description=Database sources keyed by name."`
	Storages       map[string]StorageSubConfig    `json:"storages,omitempty" jsonschema:"title=Storages,description=Storage destinations keyed by name."`
	DefaultStorage string                         `json:"default_storage,omitempty" jsonschema:"title=DefaultStorage,description=Default storage name."`
	Notifiers      map[string]NotifierSubConfig   `json:"notifiers,omitempty" jsonschema:"title=Notifiers,description=Notification providers keyed by name."`
	BeforeScript   string                         `json:"before_script,omitempty" jsonschema:"title=BeforeScript,description=Script executed before backup."`
	AfterScript    string                         `json:"after_script,omitempty" jsonschema:"title=AfterScript,description=Script executed after backup."`
}

type SubConfig struct {
	Name string `json:"name,omitempty" jsonschema:"title=Name,description=Sub-configuration name."`
}

type DatabaseSubConfig struct {
	SubConfig
	Type string `json:"type" jsonschema:"title=Type,description=Database type,enum=mysql,enum=postgresql,enum=redis,enum=mongodb,enum=sqlite,enum=mssql,enum=influxdb,enum=mariadb,enum=etcd,enum=firebird,enum=foundationdb"`
}

type StorageSubConfig struct {
	SubConfig
	Type string `json:"type" jsonschema:"title=Type,description=Storage type,enum=local,enum=ftp,enum=sftp,enum=scp,enum=s3,enum=oss,enum=gcs,enum=azure,enum=b2,enum=r2,enum=spaces,enum=cos,enum=us3,enum=kodo,enum=bos,enum=minio,enum=obs,enum=tos,enum=upyun,enum=webdav"`
}

type CompressSubConfig struct {
	SubConfig
	Type string `json:"type,omitempty" jsonschema:"title=Type,description=Compression type,enum=tar,enum=tgz,enum=7z"`
}

type EncryptSubConfig struct {
	SubConfig
	Type string `json:"type,omitempty" jsonschema:"title=Type,description=Encryption type,enum=openssl"`
}

type NotifierSubConfig struct {
	SubConfig
	Type string `json:"type" jsonschema:"title=Type,description=Notifier type,enum=mail,enum=webhook,enum=discord,enum=slack,enum=feishu,enum=dingtalk,enum=github,enum=telegram,enum=ses,enum=postmark,enum=sendgrid,enum=resend,enum=healthchecks,enum=wxwork,enum=googlechat"`
}

func main() {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: true,
		RequiredFromJSONSchemaTags: true,
	}

	schema := reflector.Reflect(&ConfigSchema{})
	schema.Version = "https://json-schema.org/draft/2020-12/schema"
	schema.ID = "https://gobackup.github.io/schema/config.schema.json"
	schema.Title = "GoBackup Config Schema"

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		panic(err)
	}

	outputPath := filepath.Join("config", "schema.json")
	if err := os.WriteFile(outputPath, append(data, '\n'), 0644); err != nil {
		panic(err)
	}
}
