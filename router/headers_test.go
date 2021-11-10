package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "/user-agent", nil)
	req.Header.Set("User-Agent", "test")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
	assert.Equal(t, "test", w.Body.String())
}

func TestHeaders(t *testing.T) {
	expected := `Header1: value1
Header2: value21
Header2: value22
Header3: value3
Host: 
`

	req, _ := http.NewRequest("GET", "/headers", nil)
	req.Header["Header1"] = []string{"value1"}
	req.Header["Header2"] = []string{"value21", "value22"}
	req.Header["Header3"] = []string{"value3"}

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
	assert.Equal(t, expected, w.Body.String())
}
