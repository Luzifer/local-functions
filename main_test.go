package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleScriptCallEcho(t *testing.T) {
	configureTestHandler(t, time.Second)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/echo", strings.NewReader("hello\nthere\n"))
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req = mux.SetURLVars(req, map[string]string{"script": "echo"})

	rec := httptest.NewRecorder()
	handleScriptCall(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ACCEPT=text/plain\n")
	assert.Contains(t, rec.Body.String(), "CONTENT_TYPE=text/plain; charset=utf-8\n")
	assert.Contains(t, rec.Body.String(), "METHOD=POST\n")
	assert.Contains(t, rec.Body.String(), "=====\n\nhello\nthere\n")
}

func TestHandleScriptCallLongrunTimeout(t *testing.T) {
	configureTestHandler(t, 10*time.Millisecond)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/longrun", nil)
	req = mux.SetURLVars(req, map[string]string{"script": "longrun"})

	rec := httptest.NewRecorder()
	handleScriptCall(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Script execution failed")
}

func TestHandleScriptCallRejectsOutsideScriptDir(t *testing.T) {
	configureTestHandler(t, time.Second)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/../main.go", nil)
	req = mux.SetURLVars(req, map[string]string{"script": "../main.go"})

	rec := httptest.NewRecorder()
	handleScriptCall(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "Not found\n", rec.Body.String())
}

func TestHandleScriptCallReturnsNotFoundForMissingScript(t *testing.T) {
	configureTestHandler(t, time.Second)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/missing", nil)
	req = mux.SetURLVars(req, map[string]string{"script": "missing"})

	rec := httptest.NewRecorder()
	handleScriptCall(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "Not found\n", rec.Body.String())
}

func TestHandleScriptCallTest(t *testing.T) {
	configureTestHandler(t, time.Second)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil)
	req = mux.SetURLVars(req, map[string]string{"script": "test"})

	rec := httptest.NewRecorder()
	handleScriptCall(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Test success\n", rec.Body.String())
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Method"))
}

func configureTestHandler(t *testing.T, timeout time.Duration) {
	t.Helper()

	cfg.CommandTimeout = timeout
	cfg.ScriptDir = "./scripts"
}
