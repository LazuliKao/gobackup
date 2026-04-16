package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

var (
	missingPropertiesPattern = regexp.MustCompile(`missing properties?: (.+)$`)
	typeErrorPattern         = regexp.MustCompile(`expected ([^,]+), but got ([^\s]+)`)

	// Exist Is config file exist
	Exist bool
	// Models configs
	Models []ModelConfig
	// gobackup base dir
	GoBackupDir string = getGoBackupDir()

	PidFilePath string = filepath.Join(GoBackupDir, "gobackup.pid")
	LogFilePath string = filepath.Join(GoBackupDir, "gobackup.log")
	Web         WebConfig

	wLock = sync.Mutex{}

	// The config file loaded at
	UpdatedAt time.Time

	onConfigChanges = make([]func(fsnotify.Event), 0)
)

type WebConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

type ScheduleConfig struct {
	Enabled bool `json:"enabled,omitempty"`
	// Cron expression
	Cron string `json:"cron,omitempty"`
	// Every
	Every string `json:"every,omitempty"`
	// At time
	At string `json:"at,omitempty"`
}

func (sc ScheduleConfig) String() string {
	if sc.Enabled {
		if len(sc.Cron) > 0 {
			return fmt.Sprintf("cron %s", sc.Cron)
		} else {
			if len(sc.At) > 0 {
				return fmt.Sprintf("every %s at %s", sc.Every, sc.At)
			} else {
				return fmt.Sprintf("every %s", sc.Every)
			}
		}
	}

	return "disabled"
}

// ModelConfig for special case
type ModelConfig struct {
	Name        string
	Description string
	// WorkDir of the gobackup started
	WorkDir        string
	TempPath       string
	DumpPath       string
	Schedule       ScheduleConfig
	CompressWith   SubConfig
	EncryptWith    SubConfig
	Archive        *viper.Viper
	Splitter       *viper.Viper
	Databases      map[string]SubConfig
	Storages       map[string]SubConfig
	DefaultStorage string
	Notifiers      map[string]SubConfig
	Viper          *viper.Viper
	BeforeScript   string
	AfterScript    string
}

func getGoBackupDir() string {
	dir := os.Getenv("GOBACKUP_DIR")
	if len(dir) == 0 {
		dir = filepath.Join(os.Getenv("HOME"), ".gobackup")
	}
	return dir
}

// SubConfig sub config info
type SubConfig struct {
	Name  string
	Type  string
	Viper *viper.Viper
}

// Init
// loadConfig from:
// - ./gobackup.yml
// - ~/.gobackup/gobackup.yml
// - /etc/gobackup/gobackup.yml
func Init(configFile string) error {
	logger := logger.Tag("Config")

	viper.SetConfigType("yaml")

	// set config file directly
	if len(configFile) > 0 {
		configFile = helper.AbsolutePath(configFile)
		logger.Info("Load config:", configFile)

		viper.SetConfigFile(configFile)
	} else {
		logger.Info("Load config from default path.")
		viper.SetConfigName("gobackup")

		// ./gobackup.yml
		viper.AddConfigPath(".")
		// ~/.gobackup/gobackup.yml
		viper.AddConfigPath("$HOME/.gobackup") // call multiple times to add many search paths
		// /etc/gobackup/gobackup.yml
		viper.AddConfigPath("/etc/gobackup/") // path to look for the config file in
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		logger.Info("Config file changed:", in.Name)
		defer onConfigChanged(in)
		if err := loadConfig(); err != nil {
			logger.Error(err.Error())
		}
	})

	return loadConfig()
}

// OnConfigChange add callback when config changed
func OnConfigChange(run func(in fsnotify.Event)) {
	onConfigChanges = append(onConfigChanges, run)
}

// Invoke callbacks when config changed
func onConfigChanged(in fsnotify.Event) {
	for _, fn := range onConfigChanges {
		fn(in)
	}
}

func validateConfig(schemaPath string, configData []byte) error {
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(schemaPath)
	if err != nil {
		return err
	}

	var data interface{}
	if err := yaml.Unmarshal(configData, &data); err != nil {
		return err
	}

	if err := schema.Validate(data); err != nil {
		if validationErr, ok := err.(*jsonschema.ValidationError); ok {
			return formatValidationError(validationErr)
		}

		return err
	}

	return nil
}

func formatValidationError(err *jsonschema.ValidationError) error {
	validationErrors := flattenValidationErrors(err)
	if len(validationErrors) == 0 {
		validationErrors = []*jsonschema.ValidationError{err}
	}

	formattedErrors := make([]string, 0, len(validationErrors))
	for _, validationErr := range validationErrors {
		formattedErrors = append(formattedErrors, formatSingleValidationError(validationErr))
	}

	sort.Strings(formattedErrors)

	return fmt.Errorf("Configuration validation failed:\n- %s", strings.Join(formattedErrors, "\n- "))
}

func flattenValidationErrors(err *jsonschema.ValidationError) []*jsonschema.ValidationError {
	if err == nil {
		return nil
	}

	if len(err.Causes) == 0 {
		return []*jsonschema.ValidationError{err}
	}

	var result []*jsonschema.ValidationError
	for _, cause := range err.Causes {
		result = append(result, flattenValidationErrors(cause)...)
	}

	return result
}

func formatSingleValidationError(err *jsonschema.ValidationError) string {
	path := jsonPointerToFieldPath(err.InstanceLocation)
	message := strings.TrimSpace(err.Message)

	if requiredField := missingRequiredField(message); requiredField != "" {
		path = joinFieldPath(path, requiredField)
		message = "field is required"
	} else if expectedType, actualType, ok := parseTypeError(message); ok {
		message = fmt.Sprintf("Invalid type. Expected: %s, got: %s", expectedType, actualType)
	}

	if path == "" {
		return message
	}

	return fmt.Sprintf("%s: %s", path, message)
}

func missingRequiredField(message string) string {
	matches := missingPropertiesPattern.FindStringSubmatch(message)
	if len(matches) != 2 {
		return ""
	}

	field := strings.TrimSpace(matches[1])
	field = strings.Trim(field, "'\"")
	field = strings.TrimSuffix(field, ",")

	if idx := strings.Index(field, ","); idx >= 0 {
		field = field[:idx]
	}
	if idx := strings.Index(field, " or "); idx >= 0 {
		field = field[:idx]
	}

	return strings.TrimSpace(strings.Trim(field, "'\""))
}

func parseTypeError(message string) (string, string, bool) {
	matches := typeErrorPattern.FindStringSubmatch(message)
	if len(matches) != 3 {
		return "", "", false
	}

	return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2]), true
}

func jsonPointerToFieldPath(pointer string) string {
	if pointer == "" || pointer == "/" {
		return ""
	}

	parts := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	for i, part := range parts {
		part = strings.ReplaceAll(part, "~1", "/")
		part = strings.ReplaceAll(part, "~0", "~")
		parts[i] = part
	}

	return strings.Join(parts, ".")
}

func joinFieldPath(base string, part string) string {
	if base == "" {
		return part
	}
	if part == "" {
		return base
	}

	return base + "." + part
}

func loadConfig() error {
	wLock.Lock()
	defer wLock.Unlock()

	logger := logger.Tag("Config")

	err := viper.ReadInConfig()
	if err != nil {
		logger.Error("Load gobackup config failed: ", err)
		return err
	}

	viperConfigFile := viper.ConfigFileUsed()
	if info, err := os.Stat(viperConfigFile); err == nil {
		// max permission: 0770
		if info.Mode()&(1<<2) != 0 {
			logger.Warnf("Other users are able to access %s with mode %v", viperConfigFile, info.Mode())
		}
	}

	logger.Info("Config file:", viperConfigFile)

	// load .env if exists in the same directory of used config file and expand variables in the config
	dotEnv := filepath.Join(filepath.Dir(viperConfigFile), ".env")
	if _, err := os.Stat(dotEnv); err == nil {
		if err := godotenv.Load(dotEnv); err != nil {
			logger.Errorf("Load %s failed: %v", dotEnv, err)
			return err
		}
	}

	cfg, _ := os.ReadFile(viperConfigFile)
	expandedCfg := []byte(os.ExpandEnv(string(cfg)))
	schemaPath := filepath.Join(filepath.Dir(viperConfigFile), "schema.json")
	if _, err := os.Stat(schemaPath); err != nil {
		schemaPath = filepath.Join(GoBackupDir, "config", "schema.json")
		if _, err := os.Stat(schemaPath); err != nil {
			schemaPath = filepath.Join("config", "schema.json")
		}
	}

	if err := validateConfig(schemaPath, expandedCfg); err != nil {
		logger.Errorf("Validate config failed: %v", err)
		return err
	}

	if err := viper.ReadConfig(strings.NewReader(string(expandedCfg))); err != nil {
		logger.Errorf("Load expanded config failed: %v", err)
		return err
	}

	// TODO: Here the `useTempWorkDir` and `workdir`, is not in config document. We need removed it.
	viper.Set("useTempWorkDir", false)
	if workdir := viper.GetString("workdir"); len(workdir) == 0 {
		// use temp dir as workdir
		dir, err := os.MkdirTemp("", "gobackup")
		if err != nil {
			return err
		}

		viper.Set("workdir", dir)
		viper.Set("useTempWorkDir", true)
	}

	Exist = true
	Models = []ModelConfig{}
	for key := range viper.GetStringMap("models") {
		model, err := loadModel(key)
		if err != nil {
			return fmt.Errorf("load model %s: %v", key, err)
		}

		Models = append(Models, model)
	}

	if len(Models) == 0 {
		return fmt.Errorf("no model found in %s", viperConfigFile)
	}

	// Load web config
	Web = WebConfig{}
	viper.SetDefault("web.host", "0.0.0.0")
	viper.SetDefault("web.port", 2703)
	Web.Host = viper.GetString("web.host")
	Web.Port = viper.GetString("web.port")
	Web.Username = viper.GetString("web.username")
	Web.Password = viper.GetString("web.password")

	UpdatedAt = time.Now()
	logger.Infof("Config loaded, found %d models.", len(Models))

	return nil
}

func loadModel(key string) (ModelConfig, error) {
	var model ModelConfig
	model.Name = key

	workdir, _ := os.Getwd()

	model.WorkDir = workdir
	model.TempPath = filepath.Join(viper.GetString("workdir"), fmt.Sprintf("%d", time.Now().UnixNano()))
	model.DumpPath = filepath.Join(model.TempPath, key)
	model.Viper = viper.Sub("models." + key)

	model.Description = model.Viper.GetString("description")
	model.Schedule = ScheduleConfig{Enabled: false}

    compressViper := model.Viper.Sub("compress_with")
    if compressViper == nil {
        compressViper = viper.New()
    }
    compressViper.SetDefault("type", "tar")
    compressViper.SetDefault("filename_format", "2006.01.02.15.04.05")
	model.CompressWith = SubConfig{
		Type:  compressViper.GetString("type"),
		Viper: compressViper,
	}

	model.EncryptWith = SubConfig{
		Type:  model.Viper.GetString("encrypt_with.type"),
		Viper: model.Viper.Sub("encrypt_with"),
	}

	model.Archive = model.Viper.Sub("archive")
	model.Splitter = model.Viper.Sub("split_with")

	model.BeforeScript = model.Viper.GetString("before_script")
	model.AfterScript = model.Viper.GetString("after_script")

	loadScheduleConfig(&model)
	loadDatabasesConfig(&model)
	loadStoragesConfig(&model)

	if len(model.Storages) == 0 {
		return ModelConfig{}, fmt.Errorf("no storage found in model %s", model.Name)
	}

	loadNotifiersConfig(&model)

	return model, nil
}

func loadScheduleConfig(model *ModelConfig) {
	subViper := model.Viper.Sub("schedule")
	model.Schedule = ScheduleConfig{Enabled: false}
	if subViper == nil {
		return
	}

	model.Schedule = ScheduleConfig{
		Enabled: true,
		Cron:    subViper.GetString("cron"),
		Every:   subViper.GetString("every"),
		At:      subViper.GetString("at"),
	}
}

func loadDatabasesConfig(model *ModelConfig) {
	subViper := model.Viper.Sub("databases")
	model.Databases = map[string]SubConfig{}
	for key := range model.Viper.GetStringMap("databases") {
		dbViper := subViper.Sub(key)
		model.Databases[key] = SubConfig{
			Name:  key,
			Type:  dbViper.GetString("type"),
			Viper: dbViper,
		}
	}
}

func loadStoragesConfig(model *ModelConfig) {
	storageConfigs := map[string]SubConfig{}

	model.DefaultStorage = model.Viper.GetString("default_storage")

	subViper := model.Viper.Sub("storages")
	for key := range model.Viper.GetStringMap("storages") {
		storageViper := subViper.Sub(key)
		storageConfigs[key] = SubConfig{
			Name:  key,
			Type:  storageViper.GetString("type"),
			Viper: storageViper,
		}

		// Set default storage
		if len(model.DefaultStorage) == 0 {
			model.DefaultStorage = key
		}
	}
	model.Storages = storageConfigs

}

func loadNotifiersConfig(model *ModelConfig) {
	subViper := model.Viper.Sub("notifiers")
	model.Notifiers = map[string]SubConfig{}
	for key := range model.Viper.GetStringMap("notifiers") {
		dbViper := subViper.Sub(key)
		model.Notifiers[key] = SubConfig{
			Name:  key,
			Type:  dbViper.GetString("type"),
			Viper: dbViper,
		}
	}
}

// GetModelConfigByName get model config by name
func GetModelConfigByName(name string) (model *ModelConfig) {
	for _, m := range Models {
		if m.Name == name {
			model = &m
			return
		}
	}
	return
}

// GetDatabaseByName get database config by name
func (model *ModelConfig) GetDatabaseByName(name string) (subConfig *SubConfig) {
	for _, m := range model.Databases {
		if m.Name == name {
			subConfig = &m
			return
		}
	}
	return
}
