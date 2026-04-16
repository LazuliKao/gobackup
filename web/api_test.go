package web

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestAPIStatus(t *testing.T) {
	code, body := invokeHttp("GET", "/status", nil, nil)

	assert.Equal(t, 200, code)
	assertMatchJSON(t, gin.H{"message": "GoBackup is running.", "version": "master"}, body)
}

func TestAPIGetModels(t *testing.T) {
	code, _ := invokeHttp("GET", "/api/config", nil, nil)

	assert.Equal(t, 200, code)
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
