package router

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	validator "github.com/dcarrillo/whatismyip/internal/validator/uuid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestGetDNSDiscoveryHandler(t *testing.T) {
	store := cache.New(cache.NoExpiration, cache.NoExpiration)
	handler := GetDNSDiscoveryHandler(store, domain, "")

	t.Run("calls next if host does not have domain suffix", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(trustedHeader, testIP.ipv4)
		req.Host = "example.com"

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler(c)
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, testIP.ipv4+"\n", w.Body.String())
	})

	t.Run("redirects if host is domain", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Host = domain

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler(c)

		assert.Equal(t, http.StatusFound, w.Code)
		r, err := url.Parse(w.Header().Get("Location"))
		assert.NoError(t, err)
		assert.True(t, validator.IsValid(strings.Split(r.Host, ".")[0]))
		assert.Equal(t, domain, strings.Join(strings.Split(r.Host, ".")[1:], "."))
	})
}

func TestHandleDNS(t *testing.T) {
	store := cache.New(cache.NoExpiration, cache.NoExpiration)
	u := uuid.New().String()

	tests := []struct {
		name      string
		subDomain string
		stored    any
	}{
		{
			name:      "not found if the subdomain is not a valid uuid",
			subDomain: "not-uuid",
			stored:    "",
		},
		{
			name:      "not found if the ip is not found in the store",
			subDomain: u,
			stored:    "",
		},
		{
			name:      "not found if the ip is in store but is not valid",
			subDomain: u,
			stored:    "bogus",
		},
		{
			name:      "not found if the store contains no string",
			subDomain: u,
			stored:    20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Host = tt.subDomain + "." + domain

			if tt.stored != "" {
				store.Add(tt.subDomain, tt.stored, cache.DefaultExpiration)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			handleDNS(c, store)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	}
}

func TestAcceptDNSRequest(t *testing.T) {
	store := cache.New(cache.NoExpiration, cache.NoExpiration)

	tests := []struct {
		name   string
		accept string
		want   string
	}{
		{
			name:   "returns json dns data",
			accept: "application/json",
			want:   jsonDNSIPv4,
		},
		{
			name:   "return plan text dns data",
			accept: "text/plain",
			want:   plainDNSIPv4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			u := uuid.New().String()
			req.Host = u + "." + domain
			req.Header.Add("Accept", tt.accept)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			store.Add(u, testIP.ipv4, cache.DefaultExpiration)
			handleDNS(c, store)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.want, w.Body.String())
		})
	}
}
