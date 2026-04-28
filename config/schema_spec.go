//go:generate go run ../cmd/generate-config-schema

package config

// Package config provides schema-only structs that mirror the YAML config shape.

// ConfigSchemaSpec describes the top-level gobackup YAML file.
type ConfigSchemaSpec struct {
	Web    WebSchemaSpec              `json:"web,omitempty" yaml:"web,omitempty"`
	Models map[string]ModelSchemaSpec `json:"models,omitempty" yaml:"models,omitempty"`
}

// WebSchemaSpec describes web auth settings.
type WebSchemaSpec struct {
	Host     string `json:"host,omitempty" yaml:"host,omitempty"`
	Port     string `json:"port,omitempty" yaml:"port,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

// ModelSchemaSpec describes one backup model.
type ModelSchemaSpec struct {
	Description    string                        `json:"description,omitempty" yaml:"description,omitempty"`
	Schedule       *ScheduleSchemaSpec           `json:"schedule,omitempty" yaml:"schedule,omitempty"`
	CompressWith   *SubConfigSchemaSpec          `json:"compress_with,omitempty" yaml:"compress_with,omitempty"`
	EncryptWith    *SubConfigSchemaSpec          `json:"encrypt_with,omitempty" yaml:"encrypt_with,omitempty"`
	Archive        *ArchiveSchemaSpec            `json:"archive,omitempty" yaml:"archive,omitempty"`
	SplitWith      *SplitWithSchemaSpec          `json:"split_with,omitempty" yaml:"split_with,omitempty"`
	Databases      map[string]DatabaseSchemaSpec `json:"databases,omitempty" yaml:"databases,omitempty"`
	Storages       map[string]StorageSchemaSpec  `json:"storages,omitempty" yaml:"storages,omitempty"`
	Notifiers      map[string]NotifierSchemaSpec `json:"notifiers,omitempty" yaml:"notifiers,omitempty"`
	DefaultStorage string                        `json:"default_storage,omitempty" yaml:"default_storage,omitempty"`
	BeforeScript   string                        `json:"before_script,omitempty" yaml:"before_script,omitempty"`
	AfterScript    string                        `json:"after_script,omitempty" yaml:"after_script,omitempty"`
}

// ScheduleSchemaSpec describes model scheduling.
type ScheduleSchemaSpec struct {
	Cron  string `json:"cron,omitempty" yaml:"cron,omitempty"`
	Every string `json:"every,omitempty" yaml:"every,omitempty"`
	At    string `json:"at,omitempty" yaml:"at,omitempty"`
}

// SubConfigSchemaSpec describes the shared common fields for inline provider configs.
type SubConfigSchemaSpec struct {
	Type           string `json:"type,omitempty" yaml:"type,omitempty"`
	FilenameFormat string `json:"filename_format,omitempty" yaml:"filename_format,omitempty"`
	Password       string `json:"password,omitempty" yaml:"password,omitempty"`
	Salt           bool   `json:"salt,omitempty" yaml:"salt,omitempty"`
	OpenSSL        bool   `json:"openssl,omitempty" yaml:"openssl,omitempty"`
}

// ArchiveSchemaSpec describes archive includes/excludes.
type ArchiveSchemaSpec struct {
	Includes []string `json:"includes,omitempty" yaml:"includes,omitempty"`
	Excludes []string `json:"excludes,omitempty" yaml:"excludes,omitempty"`
}

// SplitWithSchemaSpec describes split configuration.
type SplitWithSchemaSpec struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

// DatabaseSchemaSpec describes database-specific inline YAML keys used in the sample config.
type DatabaseSchemaSpec struct {
	Type       string `json:"type,omitempty" yaml:"type,omitempty"`
	Host       string `json:"host,omitempty" yaml:"host,omitempty"`
	Port       int    `json:"port,omitempty" yaml:"port,omitempty"`
	Database   string `json:"database,omitempty" yaml:"database,omitempty"`
	Username   string `json:"username,omitempty" yaml:"username,omitempty"`
	Password   string `json:"password,omitempty" yaml:"password,omitempty"`
	Mode       string `json:"mode,omitempty" yaml:"mode,omitempty"`
	RdbPath    string `json:"rdb_path,omitempty" yaml:"rdb_path,omitempty"`
	InvokeSave bool   `json:"invoke_save,omitempty" yaml:"invoke_save,omitempty"`
}

// StorageSchemaSpec describes storage-specific inline YAML keys used in the sample config.
type StorageSchemaSpec struct {
	Type            string `json:"type,omitempty" yaml:"type,omitempty"`
	Keep            int    `json:"keep,omitempty" yaml:"keep,omitempty"`
	Path            string `json:"path,omitempty" yaml:"path,omitempty"`
	Host            string `json:"host,omitempty" yaml:"host,omitempty"`
	Port            int    `json:"port,omitempty" yaml:"port,omitempty"`
	PrivateKey      string `json:"private_key,omitempty" yaml:"private_key,omitempty"`
	Username        string `json:"username,omitempty" yaml:"username,omitempty"`
	Password        string `json:"password,omitempty" yaml:"password,omitempty"`
	Timeout         int    `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Bucket          string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Region          string `json:"region,omitempty" yaml:"region,omitempty"`
	AccessKeyID     string `json:"access_key_id,omitempty" yaml:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty" yaml:"secret_access_key,omitempty"`
	Account         string `json:"account,omitempty" yaml:"account,omitempty"`
	TenantID        string `json:"tenant_id,omitempty" yaml:"tenant_id,omitempty"`
	ClientID        string `json:"client_id,omitempty" yaml:"client_id,omitempty"`
	ClientSecret    string `json:"client_secret,omitempty" yaml:"client_secret,omitempty"`
}

// NotifierSchemaSpec describes notifier-specific inline YAML keys used in the sample config.
type NotifierSchemaSpec struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}
