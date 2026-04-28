package web

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/storage"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func init() {
	if err := config.Init("../gobackup_test.yml"); err != nil {
		panic(err.Error())
	}
}

func assertMatchJSON(t *testing.T, expected map[string]any, actual string) {
	t.Helper()

	expectedJSON, err := json.Marshal(expected)
	assert.NoError(t, err)
	assert.Equal(t, string(expectedJSON), actual)
}

func invokeHttp(method string, path string, headers map[string]string, data map[string]any) (statusCode int, body string) {
	r := setupRouter("master")
	w := httptest.NewRecorder()

	bodyBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	for key := range headers {
		req.Header.Add(key, headers[key])
	}

	if len(data) > 0 {
		req.Header.Add("Content-Type", "application/json")
	}

	r.ServeHTTP(w, req)

	return w.Code, w.Body.String()
}

func invokeHttpBody(method string, path string, headers map[string]string, body []byte) (statusCode int, responseBody string) {
	r := setupRouter("master")
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(method, path, bytes.NewBuffer(body))
	for key := range headers {
		req.Header.Add(key, headers[key])
	}

	r.ServeHTTP(w, req)

	return w.Code, w.Body.String()
}

func configEditorHeaders(extra map[string]string) map[string]string {
	headers := map[string]string{}
	for key, value := range extra {
		headers[key] = value
	}

	credentials := config.Web.Username + ":" + config.Web.Password
	headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials))

	return headers
}

func withTempConfigFile(t *testing.T) string {
	t.Helper()

	originalConfigFile := viper.ConfigFileUsed()
	originalContent, err := os.ReadFile(originalConfigFile)
	assert.NoError(t, err)

	tempConfigFile := filepath.Join(t.TempDir(), filepath.Base(originalConfigFile))
	err = os.WriteFile(tempConfigFile, originalContent, 0o600)
	assert.NoError(t, err)

	err = config.Init(tempConfigFile)
	assert.NoError(t, err)

	t.Cleanup(func() {
		err := config.Init(originalConfigFile)
		assert.NoError(t, err)
	})

	return tempConfigFile
}

func TestAPIStatus(t *testing.T) {
	code, body := invokeHttp("GET", "/status", nil, nil)

	assert.Equal(t, 200, code)
	assertMatchJSON(t, gin.H{"message": "GoBackup is running.", "version": "master"}, body)
}

func TestAPIGetModels(t *testing.T) {
	code, _ := invokeHttp("GET", "/api/config", nil, nil)

	assert.Equal(t, 200, code)
}

func TestAPIGetConfigFile(t *testing.T) {
	code, body := invokeHttp("GET", "/api/config/file", configEditorHeaders(nil), nil)

	assert.Equal(t, 200, code)

	expected, err := os.ReadFile(viper.ConfigFileUsed())
	assert.NoError(t, err)
	assert.Equal(t, string(expected), body)
}

func TestAPIPostConfigFileSavesValidYAML(t *testing.T) {
	tempConfigFile := withTempConfigFile(t)
	originalContent, err := os.ReadFile(tempConfigFile)
	assert.NoError(t, err)

	updatedContent := strings.Replace(string(originalContent), "description: \"This is base test.\"", "description: \"Updated from API save test.\"", 1)
	code, body := invokeHttpBody("POST", "/api/config/file", configEditorHeaders(map[string]string{"Content-Type": "text/yaml"}), []byte(updatedContent))

	assert.Equal(t, 200, code)
	assertMatchJSON(t, gin.H{"message": "config file saved"}, body)

	savedContent, err := os.ReadFile(tempConfigFile)
	assert.NoError(t, err)
	assert.Equal(t, updatedContent, string(savedContent))

	err = os.WriteFile(tempConfigFile, originalContent, 0o600)
	assert.NoError(t, err)

	restoredContent, err := os.ReadFile(tempConfigFile)
	assert.NoError(t, err)
	assert.Equal(t, string(originalContent), string(restoredContent))
}

func TestAPIPostConfigFileRejectsInvalidYAML(t *testing.T) {
	tempConfigFile := withTempConfigFile(t)
	originalContent, err := os.ReadFile(tempConfigFile)
	assert.NoError(t, err)

	invalidContent := []byte("models:\n  broken: [\n")
	code, body := invokeHttpBody("POST", "/api/config/file", configEditorHeaders(map[string]string{"Content-Type": "text/yaml"}), invalidContent)

	assert.Equal(t, 400, code)
	assert.Equal(t, true, strings.Contains(body, "invalid config file:"))

	currentContent, err := os.ReadFile(tempConfigFile)
	assert.NoError(t, err)
	assert.Equal(t, string(originalContent), string(currentContent))
}

func TestAPIPostConfigFileRequiresActivePath(t *testing.T) {
	originalConfigFile := viper.ConfigFileUsed()
	t.Cleanup(func() {
		err := config.Init(originalConfigFile)
		assert.NoError(t, err)
	})

	viper.Reset()
	viper.SetConfigType("yaml")

	code, body := invokeHttpBody("POST", "/api/config/file", configEditorHeaders(map[string]string{"Content-Type": "text/yaml"}), []byte("models: {}\n"))

	assert.Equal(t, 404, code)
	assertMatchJSON(t, gin.H{"message": "config file not found"}, body)
}

func TestAPIPostConfigFileRejectsMissingActiveFile(t *testing.T) {
	tempConfigFile := withTempConfigFile(t)
	originalContent, err := os.ReadFile(tempConfigFile)
	assert.NoError(t, err)

	err = os.Remove(tempConfigFile)
	assert.NoError(t, err)

	code, body := invokeHttpBody("POST", "/api/config/file", configEditorHeaders(map[string]string{"Content-Type": "text/yaml"}), originalContent)

	assert.Equal(t, 404, code)
	assert.Equal(t, true, strings.Contains(body, "config file not found:"))

	_, err = os.Stat(tempConfigFile)
	assert.Equal(t, true, os.IsNotExist(err))
}

func TestAPIPostConfigFileRejectsConfigWithoutModels(t *testing.T) {
	tempConfigFile := withTempConfigFile(t)
	originalContent, err := os.ReadFile(tempConfigFile)
	assert.NoError(t, err)

	invalidContent := []byte("web:\n  username: gobackup\n")
	code, body := invokeHttpBody("POST", "/api/config/file", configEditorHeaders(map[string]string{"Content-Type": "text/yaml"}), invalidContent)

	assert.Equal(t, 400, code)
	assert.Equal(t, true, strings.Contains(body, "invalid config file: no model found in config"))

	currentContent, err := os.ReadFile(tempConfigFile)
	assert.NoError(t, err)
	assert.Equal(t, string(originalContent), string(currentContent))
}

func TestAPIPostPeform(t *testing.T) {
	code, body := invokeHttp("POST", "/api/perform", nil, gin.H{"model": "test_model"})

	assert.Equal(t, 200, code)
	assertMatchJSON(t, gin.H{"message": "Backup: test_model performed in background."}, body)
}

func TestAPIDownloadStreamsLocalFile(t *testing.T) {
	tempDir := t.TempDir()
	fileName := "backup.tar.gz"
	fileContent := "streamed backup payload"
	filePath := filepath.Join(tempDir, fileName)
	assert.NoError(t, os.WriteFile(filePath, []byte(fileContent), 0644))

	originalModels := config.Models
	defer func() {
		config.Models = originalModels
	}()

	config.Models = []config.ModelConfig{{
		Name:           "download_test",
		DefaultStorage: "local",
		Storages: map[string]config.SubConfig{
			"local": {
				Name: "local",
				Type: "local",
				Viper: func() *viper.Viper {
					vp := viper.New()
					vp.Set("path", tempDir)
					return vp
				}(),
			},
		},
	}}

	r := setupRouter("master")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/download?model=download_test&path="+fileName, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
	assert.Contains(t, w.Body.String(), "Local is not support download")
}

func TestAPIDownloadStreamsReaderResult(t *testing.T) {
	body := "streamed bytes"
	r := setupRouter("master")
	w := httptest.NewRecorder()

	originalDownload := storageDownload
	defer func() {
		storageDownload = originalDownload
	}()

	storageDownload = func(model config.ModelConfig, fileKey string) (*storage.DownloadResult, error) {
		return &storage.DownloadResult{
			Reader:      io.NopCloser(bytes.NewBufferString(body)),
			Filename:    "backup.tar.gz",
			Size:        int64(len(body)),
			ContentType: "application/gzip",
		}, nil
	}

	originalModels := config.Models
	defer func() {
		config.Models = originalModels
	}()

	config.Models = []config.ModelConfig{{
		Name:           "download_test",
		DefaultStorage: "local",
		Storages: map[string]config.SubConfig{
			"local": {Name: "local", Type: "local"},
		},
	}}

	req, _ := http.NewRequest("GET", "/api/download?model=download_test&path=backup.tar.gz", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, body, w.Body.String())
	assert.Equal(t, "application/gzip", w.Header().Get("Content-Type"))
	assert.Equal(t, "attachment; filename=\"backup.tar.gz\"", w.Header().Get("Content-Disposition"))
	assert.Equal(t, "14", w.Header().Get("Content-Length"))
}
