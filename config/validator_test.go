package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateConfig_ValidConfig(t *testing.T) {
	schemaPath := getTestSchemaPath(t)

	tests := []struct {
		name    string
		config  string
	}{
		{
			name: "minimal valid config with models",
			config: `
models:
  my_backup:
    storages:
      local:
        type: local
`,
		},
		{
			name: "config with database",
			config: `
models:
  my_backup:
    databases:
      mysql_db:
        type: mysql
    storages:
      local:
        type: local
`,
		},
		{
			name: "config with schedule",
			config: `
models:
  my_backup:
    schedule:
      cron: "0 2 * * *"
    storages:
      local:
        type: local
`,
		},
		{
			name: "config with compress",
			config: `
models:
  my_backup:
    compress_with:
      type: tgz
    storages:
      local:
        type: local
`,
		},
		{
			name: "config with encrypt",
			config: `
models:
  my_backup:
    encrypt_with:
      type: openssl
    storages:
      local:
        type: local
`,
		},
		{
			name: "config with notifier",
			config: `
models:
  my_backup:
    storages:
      local:
        type: local
    notifiers:
      email:
        type: mail
`,
		},
		{
			name: "config with web",
			config: `
workdir: /tmp/gobackup
web:
  host: "0.0.0.0"
  port: "2703"
models:
  my_backup:
    storages:
      local:
        type: local
`,
		},
		{
			name: "config with multiple storages",
			config: `
models:
  my_backup:
    storages:
      local:
        type: local
      s3_backup:
        type: s3
`,
		},
		{
			name: "config with all database types",
			config: `
models:
  my_backup:
    databases:
      mysql_db:
        type: mysql
      pg_db:
        type: postgresql
      redis_db:
        type: redis
      mongo_db:
        type: mongodb
      sqlite_db:
        type: sqlite
      mssql_db:
        type: mssql
      influx_db:
        type: influxdb
      mariadb_db:
        type: mariadb
      etcd_db:
        type: etcd
      firebird_db:
        type: firebird
      fdb_db:
        type: foundationdb
    storages:
      local:
        type: local
`,
		},
		{
			name: "config with all storage types",
			config: `
models:
  my_backup:
    storages:
      local_store:
        type: local
      ftp_store:
        type: ftp
      sftp_store:
        type: sftp
      scp_store:
        type: scp
      s3_store:
        type: s3
      oss_store:
        type: oss
      gcs_store:
        type: gcs
      azure_store:
        type: azure
      b2_store:
        type: b2
      r2_store:
        type: r2
      spaces_store:
        type: spaces
      cos_store:
        type: cos
      us3_store:
        type: us3
      kodo_store:
        type: kodo
      bos_store:
        type: bos
      minio_store:
        type: minio
      obs_store:
        type: obs
      tos_store:
        type: tos
      upyun_store:
        type: upyun
      webdav_store:
        type: webdav
`,
		},
		{
			name: "config with all notifier types",
			config: `
models:
  my_backup:
    storages:
      local:
        type: local
    notifiers:
      mail_ntf:
        type: mail
      webhook_ntf:
        type: webhook
      discord_ntf:
        type: discord
      slack_ntf:
        type: slack
      feishu_ntf:
        type: feishu
      dingtalk_ntf:
        type: dingtalk
      github_ntf:
        type: github
      telegram_ntf:
        type: telegram
      ses_ntf:
        type: ses
      postmark_ntf:
        type: postmark
      sendgrid_ntf:
        type: sendgrid
      resend_ntf:
        type: resend
      healthchecks_ntf:
        type: healthchecks
      wxwork_ntf:
        type: wxwork
      googlechat_ntf:
        type: googlechat
`,
		},
		{
			name: "config with all compress types",
			config: `
models:
  my_backup:
    compress_with:
      type: tar
    storages:
      local:
        type: local
`,
		},
		{
			name: "config with tgz compress",
			config: `
models:
  my_backup:
    compress_with:
      type: tgz
    storages:
      local:
        type: local
`,
		},
		{
			name: "config with 7z compress",
			config: `
models:
  my_backup:
    compress_with:
      type: 7z
    storages:
      local:
        type: local
`,
		},
		{
			name: "config with description and scripts",
			config: `
models:
  my_backup:
    description: "Daily backup of production database"
    before_script: |
      echo "Starting backup"
    after_script: |
      echo "Backup completed"
    storages:
      local:
        type: local
`,
		},
		{
			name: "config with default_storage",
			config: `
models:
  my_backup:
    default_storage: local
    storages:
      local:
        type: local
      s3_backup:
        type: s3
`,
		},
		{
			name: "config with archive and split_with",
			config: `
models:
  my_backup:
    archive:
      includes:
        - /etc/nginx
    split_with:
      size: 100MB
    storages:
      local:
        type: local
`,
		},
		{
			name: "empty config with models only",
			config: `
models:
  test:
    storages:
      local:
        type: local
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(schemaPath, []byte(tt.config))
			if err != nil {
				t.Errorf("validateConfig() error = %v", err)
			}
		})
	}
}

func TestValidateConfig_InvalidConfig_MissingRequired(t *testing.T) {
	schemaPath := getTestSchemaPath(t)

	tests := []struct {
		name           string
		config         string
		expectedErrMsg string
	}{
		{
			name: "missing models entirely",
			config: `
workdir: /tmp/gobackup
`,
			expectedErrMsg: "models",
		},
		{
			name: "empty models",
			config: `
models:
`,
			expectedErrMsg: "models",
		},
		{
			name: "model without storages",
			config: `
models:
  my_backup:
    description: "Test backup"
`,
			expectedErrMsg: "storages",
		},
		{
			name: "storage without type",
			config: `
models:
  my_backup:
    storages:
      local:
        path: /backups
`,
			expectedErrMsg: "type",
		},
		{
			name: "database without type",
			config: `
models:
  my_backup:
    databases:
      my_db:
        host: localhost
    storages:
      local:
        type: local
`,
			expectedErrMsg: "type",
		},
		{
			name: "compress_without type",
			config: `
models:
  my_backup:
    compress_with:
      level: 9
    storages:
      local:
        type: local
`,
			expectedErrMsg: "type",
		},
		{
			name: "encrypt_without type",
			config: `
models:
  my_backup:
    encrypt_with:
      password: secret
    storages:
      local:
        type: local
`,
			expectedErrMsg: "type",
		},
		{
			name: "notifier without type",
			config: `
models:
  my_backup:
    storages:
      local:
        type: local
    notifiers:
      email:
        smtp_host: smtp.example.com
`,
			expectedErrMsg: "type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(schemaPath, []byte(tt.config))
			if err == nil {
				t.Errorf("validateConfig() expected error but got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.expectedErrMsg) {
				t.Errorf("validateConfig() error = %v, expected to contain %q", err, tt.expectedErrMsg)
			}
		})
	}
}

func TestValidateConfig_InvalidConfig_TypeErrors(t *testing.T) {
	schemaPath := getTestSchemaPath(t)

	tests := []struct {
		name           string
		config         string
		expectedErrMsg string
	}{
		{
			name: "workdir as integer instead of string",
			config: `
workdir: 123
models:
  my_backup:
    storages:
      local:
        type: local
`,
			expectedErrMsg: "workdir",
		},
		{
			name: "web.port as integer instead of string",
			config: `
models:
  my_backup:
    storages:
      local:
        type: local
web:
  port: 2703
`,
			expectedErrMsg: "port",
		},
		{
			name: "description as array instead of string",
			config: `
models:
  my_backup:
    description: ["test"]
    storages:
      local:
        type: local
`,
			expectedErrMsg: "description",
		},
		{
			name: "storages as array instead of object",
			config: `
models:
  my_backup:
    storages:
      - type: local
`,
			expectedErrMsg: "storages",
		},
		{
			name: "databases as string instead of object",
			config: `
models:
  my_backup:
    databases: "mysql"
    storages:
      local:
        type: local
`,
			expectedErrMsg: "databases",
		},
		{
			name: "schedule as string instead of object",
			config: `
models:
  my_backup:
    schedule: "daily"
    storages:
      local:
        type: local
`,
			expectedErrMsg: "schedule",
		},
		{
			name: "compress_with as string instead of object",
			config: `
models:
  my_backup:
    compress_with: "tgz"
    storages:
      local:
        type: local
`,
			expectedErrMsg: "compress_with",
		},
		{
			name: "encrypt_with as boolean instead of object",
			config: `
models:
  my_backup:
    encrypt_with: true
    storages:
      local:
        type: local
`,
			expectedErrMsg: "encrypt_with",
		},
		{
			name: "notifiers as array instead of object",
			config: `
models:
  my_backup:
    storages:
      local:
        type: local
    notifiers:
      - type: mail
`,
			expectedErrMsg: "notifiers",
		},
		{
			name: "storage type as integer instead of string",
			config: `
models:
  my_backup:
    storages:
      local:
        type: 123
`,
			expectedErrMsg: "type",
		},
		{
			name: "database type as boolean instead of string",
			config: `
models:
  my_backup:
    databases:
      my_db:
        type: true
    storages:
      local:
        type: local
`,
			expectedErrMsg: "type",
		},
		{
			name: "compress type as array instead of string",
			config: `
models:
  my_backup:
    compress_with:
      type: ["tar"]
    storages:
      local:
        type: local
`,
			expectedErrMsg: "type",
		},
		{
			name: "encrypt type as object instead of string",
			config: `
models:
  my_backup:
    encrypt_with:
      type:
        name: openssl
    storages:
      local:
        type: local
`,
			expectedErrMsg: "type",
		},
		{
			name: "notifier type as null instead of string",
			config: `
models:
  my_backup:
    storages:
      local:
        type: local
    notifiers:
      email:
        type: null
`,
			expectedErrMsg: "type",
		},
		{
			name: "schedule.cron as integer instead of string",
			config: `
models:
  my_backup:
    schedule:
      cron: 12345
    storages:
      local:
        type: local
`,
			expectedErrMsg: "cron",
		},
		{
			name: "schedule.every as boolean instead of string",
			config: `
models:
  my_backup:
    schedule:
      every: true
    storages:
      local:
        type: local
`,
			expectedErrMsg: "every",
		},
		{
			name: "schedule.at as array instead of string",
			config: `
models:
  my_backup:
    schedule:
      at: ["08:00"]
    storages:
      local:
        type: local
`,
			expectedErrMsg: "at",
		},
		{
			name: "default_storage as integer instead of string",
			config: `
models:
  my_backup:
    default_storage: 123
    storages:
      local:
        type: local
`,
			expectedErrMsg: "default_storage",
		},
		{
			name: "before_script as boolean instead of string",
			config: `
models:
  my_backup:
    before_script: true
    storages:
      local:
        type: local
`,
			expectedErrMsg: "before_script",
		},
		{
			name: "after_script as array instead of string",
			config: `
models:
  my_backup:
    after_script: ["echo done"]
    storages:
      local:
        type: local
`,
			expectedErrMsg: "after_script",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(schemaPath, []byte(tt.config))
			if err == nil {
				t.Errorf("validateConfig() expected error but got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.expectedErrMsg) {
				t.Errorf("validateConfig() error = %v, expected to contain %q", err, tt.expectedErrMsg)
			}
		})
	}
}

func TestValidateConfig_InvalidConfig_EnumErrors(t *testing.T) {
	schemaPath := getTestSchemaPath(t)

	tests := []struct {
		name           string
		config         string
		expectedErrMsg string
	}{
		{
			name: "invalid storage type",
			config: `
models:
  my_backup:
    storages:
      local:
        type: invalid_storage
`,
			expectedErrMsg: "type",
		},
		{
			name: "invalid database type",
			config: `
models:
  my_backup:
    databases:
      my_db:
        type: oracle
    storages:
      local:
        type: local
`,
			expectedErrMsg: "type",
		},
		{
			name: "invalid compress type",
			config: `
models:
  my_backup:
    compress_with:
      type: zip
    storages:
      local:
        type: local
`,
			expectedErrMsg: "type",
		},
		{
			name: "invalid encrypt type",
			config: `
models:
  my_backup:
    encrypt_with:
      type: gpg
    storages:
      local:
        type: local
`,
			expectedErrMsg: "type",
		},
		{
			name: "invalid notifier type",
			config: `
models:
  my_backup:
    storages:
      local:
        type: local
    notifiers:
      email:
        type: sms
`,
			expectedErrMsg: "type",
		},
		{
			name: "empty storage type",
			config: `
models:
  my_backup:
    storages:
      local:
        type: ""
`,
			expectedErrMsg: "type",
		},
		{
			name: "database type with typo",
			config: `
models:
  my_backup:
    databases:
      my_db:
        type: postgressql
    storages:
      local:
        type: local
`,
			expectedErrMsg: "type",
		},
		{
			name: "compress type with typo",
			config: `
models:
  my_backup:
    compress_with:
      type: targz
    storages:
      local:
        type: local
`,
			expectedErrMsg: "type",
		},
		{
			name: "storage type case sensitive",
			config: `
models:
  my_backup:
    storages:
      local:
        type: Local
`,
			expectedErrMsg: "type",
		},
		{
			name: "database type case sensitive",
			config: `
models:
  my_backup:
    databases:
      my_db:
        type: MySQL
    storages:
      local:
        type: local
`,
			expectedErrMsg: "type",
		},
		{
			name: "notifier type case sensitive",
			config: `
models:
  my_backup:
    storages:
      local:
        type: local
    notifiers:
      email:
        type: Mail
`,
			expectedErrMsg: "type",
		},
		{
			name: "multiple invalid types in one config",
			config: `
models:
  my_backup:
    databases:
      db1:
        type: invalid_db
    compress_with:
      type: invalid_compress
    storages:
      local:
        type: invalid_storage
`,
			expectedErrMsg: "type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(schemaPath, []byte(tt.config))
			if err == nil {
				t.Errorf("validateConfig() expected error but got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.expectedErrMsg) {
				t.Errorf("validateConfig() error = %v, expected to contain %q", err, tt.expectedErrMsg)
			}
		})
	}
}

func TestValidateConfig_ComplexScenarios(t *testing.T) {
	schemaPath := getTestSchemaPath(t)

	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{
			name: "multiple models with different configurations",
			config: `
models:
  daily_backup:
    description: "Daily database backup"
    schedule:
      cron: "0 2 * * *"
    databases:
      mysql_db:
        type: mysql
        host: localhost
    compress_with:
      type: tgz
    storages:
      local:
        type: local
        path: /backups/daily
      s3_backup:
        type: s3
  weekly_backup:
    description: "Weekly full backup"
    schedule:
      every: "1week"
      at: "03:00"
    databases:
      postgres_db:
        type: postgresql
      redis_db:
        type: redis
    compress_with:
      type: 7z
    encrypt_with:
      type: openssl
    storages:
      gcs_backup:
        type: gcs
    notifiers:
      slack_ntf:
        type: slack
`,
			wantErr: false,
		},
		{
			name: "model with nested additional properties",
			config: `
models:
  my_backup:
    storages:
      s3:
        type: s3
        bucket: my-bucket
        region: us-east-1
        access_key_id: $AWS_KEY
        secret_access_key: $AWS_SECRET
`,
			wantErr: false,
		},
		{
			name: "completely empty config",
			config: ``,
			wantErr: true,
		},
		{
			name: "config with only comments",
			config: `
# This is a comment
# workdir: /tmp
`,
			wantErr: true,
		},
		{
			name: "config with environment variables",
			config: `
models:
  my_backup:
    storages:
      s3:
        type: s3
        bucket: $BUCKET_NAME
`,
			wantErr: false,
		},
		{
			name: "deeply nested invalid type",
			config: `
models:
  my_backup:
    storages:
      local:
        type: local
    schedule:
      cron:
        minute: "0"
`,
			wantErr: true,
		},
		{
			name: "multiple errors should all be reported",
			config: `
models:
  my_backup:
    databases:
      db1:
        type: invalid
    compress_with:
      type: invalid
    storages:
      local:
        type: invalid
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(schemaPath, []byte(tt.config))
			if tt.wantErr && err == nil {
				t.Errorf("validateConfig() expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateConfig() unexpected error = %v", err)
			}
		})
	}
}

func TestFormatValidationError(t *testing.T) {
	schemaPath := getTestSchemaPath(t)

	tests := []struct {
		name              string
		config            string
		expectedInErrMsg  string
		shouldContainPath bool
	}{
		{
			name: "error message contains field path",
			config: `
models:
  my_backup:
    storages:
      local:
        type: invalid
`,
			expectedInErrMsg:  "type",
			shouldContainPath: true,
		},
		{
			name: "missing property shows field is required",
			config: `
models:
  my_backup:
    storages:
      local:
        path: /backups
`,
			expectedInErrMsg:  "required",
			shouldContainPath: true,
		},
		{
			name: "type error shows expected type",
			config: `
models:
  my_backup:
    storages:
      local:
        type: 123
`,
			expectedInErrMsg:  "type",
			shouldContainPath: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(schemaPath, []byte(tt.config))
			if err == nil {
				t.Errorf("validateConfig() expected error but got nil")
				return
			}
			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.expectedInErrMsg) {
				t.Errorf("error message %q should contain %q", errMsg, tt.expectedInErrMsg)
			}
			if tt.shouldContainPath && !strings.Contains(errMsg, "Configuration validation failed:") {
				t.Errorf("error message should start with 'Configuration validation failed:'")
			}
		})
	}
}

func getTestSchemaPath(t *testing.T) string {
	t.Helper()

	// Try multiple locations for the schema file
	possiblePaths := []string{
		"schema.json",
		filepath.Join("config", "schema.json"),
		filepath.Join("..", "config", "schema.json"),
		filepath.Join("/repos/gobackup", "config", "schema.json"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	t.Fatalf("Could not find schema.json in any of the expected locations: %v", possiblePaths)
	return ""
}
