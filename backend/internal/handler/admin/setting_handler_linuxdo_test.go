package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newLinuxDoSettingsHandler() (*SettingHandler, *settingHandlerRepoStub) {
	repo := &settingHandlerRepoStub{values: map[string]string{}}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)
	return handler, repo
}

func baseValidLinuxDoBody() map[string]any {
	return map[string]any{
		"linuxdo_connect_enabled":       true,
		"linuxdo_connect_client_id":     "linuxdo-client",
		"linuxdo_connect_client_secret": "linuxdo-secret",
		"linuxdo_connect_redirect_url":  "https://example.com/api/v1/auth/oauth/linuxdo/callback",
	}
}

// TestSettingsPUT_LinuxDo_BypassRegistration_RoundTrip verifies save+load for bypass_registration.
func TestSettingsPUT_LinuxDo_BypassRegistration_RoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _ := newLinuxDoSettingsHandler()

	body := baseValidLinuxDoBody()
	body["linuxdo_connect_bypass_registration"] = true

	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, data["linuxdo_connect_bypass_registration"])
}

func TestSettingsPUT_LinuxDo_BypassRegistration_DefaultFalse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _ := newLinuxDoSettingsHandler()

	body := baseValidLinuxDoBody()
	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, false, data["linuxdo_connect_bypass_registration"])
}
