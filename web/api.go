package web

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/logger"
	"github.com/gobackup/gobackup/model"
	"github.com/gobackup/gobackup/storage"
	"github.com/stoicperlman/fls"
)

//go:embed dist
var staticFS embed.FS
var logFile *os.File
var storageDownload = storage.Download

var errConfigPathNotFound = errors.New("config file not found")

type embedFileSystem struct {
	http.FileSystem
	indexes bool
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	f, err := e.Open(path)
	if err != nil {
		return false
	}

	// check if indexing is allowed
	s, _ := f.Stat()
	if s.IsDir() && !e.indexes {
		return false
	}

	return true
}

// StartHTTP run API server
func StartHTTP(version string) (err error) {
	logger := logger.Tag("API")

	if len(config.Web.Password) == 0 {
		logger.Warn("You are running with insecure API server. Please don't forget setup `web.password` in config file for more safety.")
	}

	logFile, err = os.Open(config.LogFilePath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	logger.Infof("Starting API server on port http://%s:%s", config.Web.Host, config.Web.Port)

	if os.Getenv("GO_ENV") == "dev" {
		go func() {
			for {
				time.Sleep(5 * time.Second)
				logger.Info("Ping", time.Now())
			}
		}()
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := setupRouter(version)

	// Enable baseAuth
	if len(config.Web.Username) > 0 && len(config.Web.Password) > 0 {
		r.Use(gin.BasicAuth(gin.Accounts{
			config.Web.Username: config.Web.Password,
		}))
	}

	fe, _ := fs.Sub(staticFS, "dist")
	embedFs := embedFileSystem{http.FS(fe), true}
	r.Use(static.Serve("/", embedFs))
	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS("/", embedFs)
	})

	return r.Run(config.Web.Host + ":" + config.Web.Port)
}

func setupRouter(version string) *gin.Engine {
	r := gin.Default()

	r.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "GoBackup is running.",
			"version": version,
		})
	})

	r.Use(func(c *gin.Context) {
		c.Next()

		// Skip if no errors
		if len(c.Errors) == 0 {
			return
		}

		c.AbortWithStatusJSON(c.Writer.Status(), gin.H{
			"message": c.Errors.String(),
		})

	})

	group := r.Group("/api")
	group.GET("/config", getConfig)
	configGroup := group.Group("/config")
	configGroup.Use(requireConfigEditorAuth)
	configGroup.GET("/paths", getConfigPaths)
	configGroup.GET("/raw", getConfigRaw)
	configGroup.POST("/save", saveConfig)
	configGroup.POST("/validate", validateConfig)
	group.GET("/list", list)
	group.GET("/download", download)
	group.POST("/perform", perform)
	group.GET("/log", log)
	return r
}

func requireConfigEditorAuth(c *gin.Context) {
	if len(config.Web.Username) == 0 || len(config.Web.Password) == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "Config editor requires API authentication.",
		})
		return
	}

	username, password, ok := c.Request.BasicAuth()
	if !ok || username != config.Web.Username || password != config.Web.Password {
		c.Header("WWW-Authenticate", `Basic realm="Authorization Required"`)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Authentication required.",
		})
		return
	}

	c.Next()
}

// GET /api/config
func getConfig(c *gin.Context) {
	models := map[string]any{}
	for _, m := range model.GetModels() {
		models[m.Config.Name] = gin.H{
			"description":   m.Config.Description,
			"schedule":      m.Config.Schedule,
			"schedule_info": m.Config.Schedule.String(),
		}
	}

	c.JSON(200, gin.H{
		"models": models,
	})
}

type configPathResponse struct {
	Paths        []string                `json:"paths"`
	AllowedPaths []configPathStatusEntry `json:"allowed_paths,omitempty"`
	CurrentPath  string                  `json:"current_path,omitempty"`
}

type configPathStatus struct {
	Path   string
	Exists bool
	Err    error
}

type configPathStatusEntry struct {
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
}


func configPathStatuses() []configPathStatus {
	knownPaths := config.KnownConfigFilePaths()
	statuses := make([]configPathStatus, 0, len(knownPaths))
	for _, candidate := range knownPaths {
		_, err := os.Stat(candidate)
		exists := err == nil
		if err != nil && errors.Is(err, os.ErrNotExist) {
			err = nil
		}
		statuses = append(statuses, configPathStatus{
			Path:   candidate,
			Exists: exists,
			Err:    err,
		})
	}

	return statuses
}

func allowedConfigPaths() []configPathStatusEntry {
	statuses := configPathStatuses()
	allowedPaths := make([]configPathStatusEntry, 0, len(statuses))
	for _, status := range statuses {
		allowedPaths = append(allowedPaths, configPathStatusEntry{
			Path:   status.Path,
			Exists: status.Exists,
		})
	}

	return allowedPaths
}

func existingConfigPaths() []string {
	statuses := configPathStatuses()
	existingPaths := make([]string, 0, len(statuses))
	for _, status := range statuses {
		if status.Exists {
			existingPaths = append(existingPaths, status.Path)
		}
	}

	return existingPaths
}

func resolveConfigPath(requestedPath string, allowMissing bool) (string, error) {
	statuses := configPathStatuses()

	if requestedPath != "" {
		resolvedPath, ok := config.ResolveKnownConfigFilePath(requestedPath)
		if !ok {
			return "", errConfigPathNotFound
		}

		for _, status := range statuses {
			if status.Path == resolvedPath {
				if status.Err != nil {
					return "", status.Err
				}
				if status.Exists || allowMissing {
					return status.Path, nil
				}

				return "", errConfigPathNotFound
			}
		}

		return "", errConfigPathNotFound
	}

	currentPath := config.CurrentConfigFilePath()
	for _, status := range statuses {
		if status.Path == currentPath {
			if status.Err != nil {
				return "", status.Err
			}
			if status.Exists || allowMissing {
				return status.Path, nil
			}

			break
		}
	}

	existingPaths := make([]string, 0, len(statuses))
	for _, status := range statuses {
		if status.Err != nil {
			return "", status.Err
		}
		if status.Exists {
			existingPaths = append(existingPaths, status.Path)
		}
	}

	if len(existingPaths) == 0 {
		if allowMissing && len(statuses) > 0 {
			return statuses[0].Path, nil
		}

		return "", errConfigPathNotFound
	}

	return existingPaths[0], nil
}

// GET /api/config/paths
func getConfigPaths(c *gin.Context) {
	c.JSON(200, configPathResponse{
		Paths:        existingConfigPaths(),
		AllowedPaths: allowedConfigPaths(),
		CurrentPath:  config.CurrentConfigFilePath(),
	})
}

// GET /api/config/raw
func getConfigRaw(c *gin.Context) {
	configPath, err := resolveConfigPath(c.Query("path"), false)
	if err != nil {
		statusCode := 500
		if errors.Is(err, errConfigPathNotFound) {
			statusCode = 404
		}
		c.AbortWithError(statusCode, err)
		return
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.Data(200, "text/plain; charset=utf-8", content)
}

type saveConfigParam struct {
	Path            *string `json:"path" binding:"required"`
	Content         *string `json:"content" binding:"required"`
	CreateIfMissing bool    `json:"create_if_missing"`
}

type validateConfigParam struct {
	Path    *string `json:"path"`
	Content *string `json:"content" binding:"required"`
}

func writeConfigAtomically(configPath string, content []byte) error {
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return writeNewConfigAtomically(configPath, content)
		}
		return err
	}

	if fileInfo.IsDir() {
		return errConfigPathNotFound
	}

	tempFile, err := os.CreateTemp(filepath.Dir(configPath), filepath.Base(configPath)+".*.tmp")
	if err != nil {
		return err
	}

	tempPath := tempFile.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()

	if _, err := tempFile.Write(content); err != nil {
		tempFile.Close()
		return err
	}

	if err := tempFile.Chmod(fileInfo.Mode()); err != nil {
		tempFile.Close()
		return err
	}

	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return err
	}

	if err := tempFile.Close(); err != nil {
		return err
	}

	return os.Rename(tempPath, configPath)
}

func writeNewConfigAtomically(configPath string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(filepath.Dir(configPath), filepath.Base(configPath)+".*.tmp")
	if err != nil {
		return err
	}

	tempPath := tempFile.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()

	if _, err := tempFile.Write(content); err != nil {
		tempFile.Close()
		return err
	}

	if err := tempFile.Chmod(0600); err != nil {
		tempFile.Close()
		return err
	}

	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return err
	}

	if err := tempFile.Close(); err != nil {
		return err
	}

	return os.Rename(tempPath, configPath)
}

// POST /api/config/save
func saveConfig(c *gin.Context) {
	var param saveConfigParam
	if err := c.ShouldBindJSON(&param); err != nil {
		c.AbortWithError(400, err)
		return
	}

	if *param.Path == "" {
		c.AbortWithError(400, fmt.Errorf("Path is required"))
		return
	}

	configPath, err := resolveConfigPath(*param.Path, param.CreateIfMissing)
	if err != nil {
		statusCode := 500
		if errors.Is(err, errConfigPathNotFound) {
			statusCode = 404
		}
		c.AbortWithError(statusCode, err)
		return
	}

	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if !param.CreateIfMissing {
				c.AbortWithError(404, errConfigPathNotFound)
				return
			}
		} else {
			c.AbortWithError(500, err)
			return
		}
	}

	if err := config.ValidateRuntimeConfig(configPath, []byte(*param.Content)); err != nil {
		c.AbortWithError(400, err)
		return
	}

	if err := writeConfigAtomically(configPath, []byte(*param.Content)); err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, gin.H{
		"message": "Config saved successfully.",
	})
}

// POST /api/config/validate
func validateConfig(c *gin.Context) {
	var param validateConfigParam
	if err := c.ShouldBindJSON(&param); err != nil {
		c.AbortWithError(400, err)
		return
	}

	requestedPath := ""
	if param.Path != nil {
		requestedPath = *param.Path
	}

	configPath, err := resolveConfigPath(requestedPath, requestedPath != "")
	if err != nil {
		statusCode := 500
		if errors.Is(err, errConfigPathNotFound) {
			statusCode = 404
		}
		c.AbortWithError(statusCode, err)
		return
	}

	if err := config.ValidateRuntimeConfig(configPath, []byte(*param.Content)); err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.JSON(200, gin.H{
		"message": "Config is valid.",
		"valid":   true,
	})
}

// POST /api/perform
func perform(c *gin.Context) {
	type performParam struct {
		Model string `form:"model" json:"model" binding:"required"`
	}

	var param performParam
	if err := c.Bind(&param); err != nil {
		logger.Errorf("Bind error: %v", err)
	}

	m := model.GetModelByName(param.Model)
	if m == nil {
		c.AbortWithError(404, fmt.Errorf("Model: \"%s\" not found", param.Model))
		return
	}

	go func() {
		if err := m.Perform(); err != nil {
			logger.Errorf("Perform error: %v", err)
		}
	}()
	c.JSON(200, gin.H{"message": fmt.Sprintf("Backup: %s performed in background.", param.Model)})
}

// GET /api/list?model=xxx&parent=
func list(c *gin.Context) {
	modelName := c.Query("model")
	m := model.GetModelByName(modelName)
	if m == nil {
		c.AbortWithError(404, fmt.Errorf("Model: \"%s\" not found", modelName))
		return
	}

	parent := c.Query("parent")
	if parent == "" {
		parent = "/"
	}

	files, err := storage.List(m.Config, parent)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, gin.H{"files": files})
}

// GET /api/download?model=xxx&path=
func download(c *gin.Context) {
	modelName := c.Query("model")
	m := model.GetModelByName(modelName)
	if m == nil {
		c.AbortWithError(404, fmt.Errorf("Model: \"%s\" not found", modelName))
		return
	}

	file := c.Query("path")
	if file == "" {
		c.AbortWithError(404, fmt.Errorf("File not found"))
		return
	}

	downloadResult, err := storageDownload(m.Config, file)
	if err != nil || downloadResult == nil {
		c.AbortWithError(500, err)
		return
	}

	defer func() {
		if err := downloadResult.Close(); err != nil {
			logger.Errorf("Failed to close download result: %v", err)
		}
	}()

	if len(downloadResult.RedirectURL) > 0 {
		c.Redirect(302, downloadResult.RedirectURL)
		return
	}

	if downloadResult.Reader == nil {
		c.AbortWithError(500, fmt.Errorf("download is not available for file: %s", file))
		return
	}

	filename := downloadResult.Filename
	if filename == "" {
		filename = path.Base(file)
	}

	contentType := downloadResult.ContentType
	if contentType == "" {
		contentType = mime.TypeByExtension(path.Ext(filename))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if downloadResult.Size > 0 {
		c.Header("Content-Length", fmt.Sprintf("%d", downloadResult.Size))
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Status(http.StatusOK)

	if _, err := io.Copy(c.Writer, downloadResult.Reader); err != nil {
		logger.Errorf("Failed to stream download %s: %v", file, err)
	}
}

// GET /api/log
func log(c *gin.Context) {
	// https://github.com/gin-gonic/examples/blob/master/realtime-chat/main.go#L27
	chanStream := tailFile()
	clientGone := c.Request.Context().Done()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			println("Client gone, close stream.")
			return false
		case msg := <-chanStream:
			if os.Getenv("GO_ENV") == "dev" {
				println(msg)
			}

			if _, err := c.Writer.WriteString(msg + "\n"); err != nil {
				logger.Errorf("Failed to write to stream: %v", err)
			}
			c.Writer.Flush()
			return true
		}
	})
}

// tailFile tail the log file and make a chain to stream output log
func tailFile() chan string {
	out_chan := make(chan string)

	file := fls.LineFile(logFile)
	if _, err := file.SeekLine(-50, io.SeekEnd); err != nil {
		logger.Errorf("Failed to seek log file: %v", err)
	}
	bf := bufio.NewReader(file)

	go func() {
		for {
			line, _, _ := bf.ReadLine()

			if len(line) == 0 {
				time.Sleep(50 * time.Millisecond)
			} else {
				out_chan <- string(line)
			}
		}
	}()

	return out_chan
}
