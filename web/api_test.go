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

func invokeConfigEditorAuth(method string, path string, headers map[string][]string, remoteAddr string) (statusCode int, body string) {
	r := setupRouter("master")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	if remoteAddr != "" {
		req.RemoteAddr = remoteAddr
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	r.ServeHTTP(w, req)
	return w.Code, w.Body.String()
}

func basicAuthHeader(username string, password string) string {
	credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return "Basic " + credentials
}

func withWebConfig(webConfig config.WebConfig, run func()) {
	originalWeb := config.Web
	defer func() {
		config.Web = originalWeb
	}()

	config.Web = webConfig
	run()
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

func TestConfigEditorAuthRequiresBasicConfigByDefault(t *testing.T) {
	withWebConfig(config.WebConfig{}, func() {
		code, body := invokeConfigEditorAuth("GET", "/api/config/paths", nil, "192.0.2.10:12345")

		assert.Equal(t, http.StatusForbidden, code)
		assert.Contains(t, body, "Config editor requires API authentication")
	})
}

func TestConfigEditorAuthAcceptsBasicAuth(t *testing.T) {
	withWebConfig(config.WebConfig{Username: "gobackup", Password: "123456"}, func() {
		code, _ := invokeConfigEditorAuth("GET", "/api/config/paths", map[string][]string{
			"Authorization": {basicAuthHeader("gobackup", "123456")},
		}, "192.0.2.10:12345")

		assert.Equal(t, http.StatusOK, code)
	})
}

func TestConfigEditorProxyAuthAcceptsAllowedUserFromTrustedProxy(t *testing.T) {
	withWebConfig(config.WebConfig{
		AuthMode: "proxy",
		ProxyAuth: config.ProxyAuthConfig{
			TrustedProxies: []string{"127.0.0.1"},
			UserHeader:     "Remote-User",
			GroupHeader:    "Remote-Groups",
			AllowedUsers:   []string{"alice"},
		},
	}, func() {
		code, _ := invokeConfigEditorAuth("GET", "/api/config/paths", map[string][]string{
			"Remote-User": {"alice"},
		}, "127.0.0.1:12345")

		assert.Equal(t, http.StatusOK, code)
	})
}

func TestConfigEditorProxyAuthRejectsUntrustedRemote(t *testing.T) {
	withWebConfig(config.WebConfig{
		AuthMode: "proxy",
		ProxyAuth: config.ProxyAuthConfig{
			TrustedProxies: []string{"127.0.0.1"},
			UserHeader:     "Remote-User",
			AllowedUsers:   []string{"alice"},
		},
	}, func() {
		code, body := invokeConfigEditorAuth("GET", "/api/config/paths", map[string][]string{
			"Remote-User": {"alice"},
		}, "192.0.2.10:12345")

		assert.Equal(t, http.StatusUnauthorized, code)
		assert.Contains(t, body, "Proxy authentication required")
	})
}

func TestConfigEditorProxyAuthRejectsDuplicateUserHeader(t *testing.T) {
	withWebConfig(config.WebConfig{
		AuthMode: "proxy",
		ProxyAuth: config.ProxyAuthConfig{
			TrustedProxies: []string{"127.0.0.0/8"},
			UserHeader:     "Remote-User",
			AllowedUsers:   []string{"alice"},
		},
	}, func() {
		code, _ := invokeConfigEditorAuth("GET", "/api/config/paths", map[string][]string{
			"Remote-User": {"alice", "mallory"},
		}, "127.0.0.1:12345")

		assert.Equal(t, http.StatusUnauthorized, code)
	})
}

func TestConfigEditorProxyAuthAcceptsAllowedGroup(t *testing.T) {
	withWebConfig(config.WebConfig{
		AuthMode: "proxy",
		ProxyAuth: config.ProxyAuthConfig{
			TrustedProxies: []string{"127.0.0.0/8"},
			UserHeader:     "Remote-User",
			GroupHeader:    "Remote-Groups",
			AllowedGroups:  []string{"admins"},
		},
	}, func() {
		code, _ := invokeConfigEditorAuth("GET", "/api/config/paths", map[string][]string{
			"Remote-User":   {"bob"},
			"Remote-Groups": {"users, admins"},
		}, "127.0.0.1:12345")

		assert.Equal(t, http.StatusOK, code)
	})
}

func TestConfigEditorBasicOrProxyAuthAcceptsEitherMethod(t *testing.T) {
	withWebConfig(config.WebConfig{
		Username: "gobackup",
		Password: "123456",
		AuthMode: "basic_or_proxy",
		ProxyAuth: config.ProxyAuthConfig{
			TrustedProxies: []string{"127.0.0.1"},
			UserHeader:     "Remote-User",
			AllowedUsers:   []string{"alice"},
		},
	}, func() {
		code, _ := invokeConfigEditorAuth("GET", "/api/config/paths", map[string][]string{
			"Authorization": {basicAuthHeader("gobackup", "123456")},
		}, "192.0.2.10:12345")
		assert.Equal(t, http.StatusOK, code)

		code, _ = invokeConfigEditorAuth("GET", "/api/config/paths", map[string][]string{
			"Remote-User": {"alice"},
		}, "127.0.0.1:12345")
		assert.Equal(t, http.StatusOK, code)
	})
}

func TestAPIProxyAuthProtectsNonConfigEndpoints(t *testing.T) {
	withWebConfig(config.WebConfig{
		AuthMode: "proxy",
		ProxyAuth: config.ProxyAuthConfig{
			TrustedProxies: []string{"127.0.0.1"},
			UserHeader:     "Remote-User",
			AllowedUsers:   []string{"alice"},
		},
	}, func() {
		code, body := invokeConfigEditorAuth("GET", "/api/list?model=test_model", map[string][]string{
			"Remote-User": {"alice"},
		}, "192.0.2.10:12345")

		assert.Equal(t, http.StatusUnauthorized, code)
		assert.Contains(t, body, "Proxy authentication required")
	})
}

func TestAPIBasicOrProxyAuthAcceptsProxyOnNonConfigEndpoints(t *testing.T) {
	withWebConfig(config.WebConfig{
		Username: "gobackup",
		Password: "123456",
		AuthMode: "basic_or_proxy",
		ProxyAuth: config.ProxyAuthConfig{
			TrustedProxies: []string{"127.0.0.1"},
			UserHeader:     "Remote-User",
			AllowedUsers:   []string{"alice"},
		},
	}, func() {
		code, body := invokeConfigEditorAuth("GET", "/api/list?model=test_model", map[string][]string{
			"Remote-User": {"alice"},
		}, "127.0.0.1:12345")

		assert.NotEqual(t, http.StatusUnauthorized, code)
		assert.NotEqual(t, http.StatusForbidden, code)
		assert.NotContains(t, body, "authentication required")
	})
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
