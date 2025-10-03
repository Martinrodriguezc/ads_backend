package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"ads_backend/internal/ads_service"
	"ads_backend/internal/domain"
	http_server "ads_backend/internal/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const baseURL = "http://localhost:8081"

func TestE2E_AdLifecycle(t *testing.T) {
	server := startTestServer(t)
	defer server.stop()

	waitForServer(t, baseURL)

	reqBody := map[string]interface{}{
		"title":      "E2E Test Ad",
		"imageUrl":   "http://example.com/e2e-ad.jpg",
		"placement":  "home_screen",
		"ttlMinutes": 60,
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/adposts", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.NotEmpty(t, resp.Header.Get("X-Correlation-ID"))

	var createdAd domain.Ad
	err = json.NewDecoder(resp.Body).Decode(&createdAd)
	require.NoError(t, err)

	assert.NotEmpty(t, createdAd.ID)
	assert.Equal(t, "E2E Test Ad", createdAd.Title)
	assert.Equal(t, "http://example.com/e2e-ad.jpg", createdAd.ImageUrl)
	assert.Equal(t, domain.HomeScreen, createdAd.Placement)
	assert.Equal(t, domain.StatusActive, createdAd.Status)
	assert.Equal(t, 60, createdAd.TTLMinutes)

	resp, err = http.Get(baseURL + "/adposts/" + createdAd.ID)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var fetchedAd domain.Ad
	err = json.NewDecoder(resp.Body).Decode(&fetchedAd)
	require.NoError(t, err)

	assert.Equal(t, createdAd.ID, fetchedAd.ID)
	assert.Equal(t, "E2E Test Ad", fetchedAd.Title)

	resp, err = http.Get(baseURL + "/adspots?placement=home_screen&status=active")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var ads []domain.Ad
	err = json.NewDecoder(resp.Body).Decode(&ads)
	require.NoError(t, err)

	assert.NotEmpty(t, ads)
	found := false
	for _, ad := range ads {
		if ad.ID == createdAd.ID {
			found = true
			assert.Equal(t, "E2E Test Ad", ad.Title)
			break
		}
	}
	assert.True(t, found, "Created ad should be in the list")

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/adposts/"+createdAd.ID+"/deactivate", nil)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var deactivatedAd domain.Ad
	err = json.NewDecoder(resp.Body).Decode(&deactivatedAd)
	require.NoError(t, err)

	assert.Equal(t, createdAd.ID, deactivatedAd.ID)
	assert.Equal(t, domain.StatusInactive, deactivatedAd.Status)

	resp, err = http.Get(baseURL + "/adspots?placement=home_screen&status=active")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&ads)
	require.NoError(t, err)

	for _, ad := range ads {
		assert.NotEqual(t, createdAd.ID, ad.ID, "Deactivated ad should not be in active list")
	}
}

type testServer struct {
	mux    *http.ServeMux
	server *http.Server
}

func startTestServer(t *testing.T) *testServer {
	t.Helper()

	os.Setenv("HTTP_PORT", "8081")
	defer os.Unsetenv("HTTP_PORT")

	log := zap.NewNop()
	service := ads_service.NewService()
	handler := http_server.NewAdsHandler(log, service)
	rateLimiter := http_server.NewRateLimiter(100, 200)
	requestLogger := http_server.NewRequestLogger(log)

	routes := []http_server.Route{handler}
	mux := http_server.NewServeMux(routes)

	var finalHandler http.Handler = mux
	finalHandler = rateLimiter.Middleware(finalHandler)
	finalHandler = requestLogger.Middleware(finalHandler)

	server := &http.Server{
		Addr:         ":8081",
		Handler:      finalHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	ts := &testServer{
		mux:    mux,
		server: server,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	return ts
}

func (ts *testServer) stop() {
	if ts.server != nil {
		ts.server.Close()
	}
}

func waitForServer(t *testing.T, url string) {
	t.Helper()

	for i := 0; i < 30; i++ {
		resp, err := http.Get(url + "/adspots?placement=home_screen")
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("Server failed to start within timeout")
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
