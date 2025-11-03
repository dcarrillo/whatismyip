package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDisabledMetrics_Middleware(t *testing.T) {
	if enabled {
		t.Skip("Skipping disabled test - metrics already enabled")
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/test", nil)

	middleware := GinMiddleware()

	assert.NotPanics(t, func() {
		middleware(c)
	})
}

func TestDisabledMetrics_GeoLookup(t *testing.T) {
	if enabled {
		t.Skip("Skipping disabled test - metrics already enabled")
	}

	assert.NotPanics(t, func() {
		RecordGeoLookup("city")
	})
}

func TestDisabledMetrics_PortScan(t *testing.T) {
	if enabled {
		t.Skip("Skipping disabled test - metrics already enabled")
	}

	assert.NotPanics(t, func() {
		RecordPortScan()
	})
}

func TestDisabledMetrics_DNSQuery(t *testing.T) {
	if enabled {
		t.Skip("Skipping disabled test - metrics already enabled")
	}

	assert.NotPanics(t, func() {
		RecordDNSQuery("A", "NOERROR")
	})
}

func TestEnable(t *testing.T) {
	Enable()

	assert.True(t, enabled, "Enable() should set enabled to true")
	assert.NotNil(t, requestsTotal, "requestsTotal should be initialized")
	assert.NotNil(t, requestDuration, "requestDuration should be initialized")
	assert.NotNil(t, requestsInFlight, "requestsInFlight should be initialized")
	assert.NotNil(t, geoLookups, "geoLookups should be initialized")
	assert.NotNil(t, portScans, "portScans should be initialized")
	assert.NotNil(t, dnsQueries, "dnsQueries should be initialized")
}

func TestEnableIdempotent(t *testing.T) {
	Enable()
	firstRequestsTotal := requestsTotal

	Enable()
	Enable()

	assert.Equal(t, firstRequestsTotal, requestsTotal, "Enable() should be idempotent")
}

func TestGinMiddleware_StatusCategories(t *testing.T) {
	Enable()

	testCases := []struct {
		status   int
		category string
	}{
		{200, "2xx"},
		{201, "2xx"},
		{301, "3xx"},
		{404, "4xx"},
		{500, "5xx"},
	}

	for _, tc := range testCases {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(GinMiddleware())
		router.GET("/test-status", func(c *gin.Context) {
			c.Status(tc.status)
		})

		initialCount := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/test-status", tc.category))

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test-status", nil)
		router.ServeHTTP(w, req)

		count := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/test-status", tc.category))
		assert.Equal(t, initialCount+1, count, "Expected count for category %s to increase by 1", tc.category)
	}
}

func TestGinMiddleware_404(t *testing.T) {
	Enable()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(GinMiddleware())

	initialCount := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/404", "4xx"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nonexistent-path", nil)
	router.ServeHTTP(w, req)

	count := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/404", "4xx"))
	assert.Equal(t, initialCount+1, count, "Expected count to increase by 1 for empty path (404)")
}

func TestGinMiddleware_RecordsDuration(t *testing.T) {
	Enable()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(GinMiddleware())
	router.GET("/test-duration", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test-duration", nil)
	router.ServeHTTP(w, req)

	metric := requestDuration.WithLabelValues("GET", "/test-duration")
	assert.NotNil(t, metric, "Expected histogram metric to exist")
}

func TestRecordGeoLookup(t *testing.T) {
	Enable()

	initialCityCount := testutil.ToFloat64(geoLookups.WithLabelValues("city"))
	initialCountryCount := testutil.ToFloat64(geoLookups.WithLabelValues("asn"))

	RecordGeoLookup("city")
	RecordGeoLookup("city")
	RecordGeoLookup("asn")

	cityCount := testutil.ToFloat64(geoLookups.WithLabelValues("city"))
	assert.Equal(t, initialCityCount+2, cityCount, "Expected city lookups to increase by 2")

	countryCount := testutil.ToFloat64(geoLookups.WithLabelValues("asn"))
	assert.Equal(t, initialCountryCount+1, countryCount, "Expected country lookups to increase by 1")
}

func TestRecordPortScan(t *testing.T) {
	Enable()

	initialCount := testutil.ToFloat64(portScans)

	RecordPortScan()
	RecordPortScan()

	count := testutil.ToFloat64(portScans)
	assert.Equal(t, initialCount+2, count, "Expected port scans to increase by 2")
}

func TestRecordDNSQuery(t *testing.T) {
	Enable()

	initialACount := testutil.ToFloat64(dnsQueries.WithLabelValues("A", "NOERROR"))
	initialAAAACount := testutil.ToFloat64(dnsQueries.WithLabelValues("AAAA", "NOERROR"))
	initialNXDOMAINCount := testutil.ToFloat64(dnsQueries.WithLabelValues("A", "NXDOMAIN"))

	RecordDNSQuery("A", "NOERROR")
	RecordDNSQuery("A", "NOERROR")
	RecordDNSQuery("AAAA", "NOERROR")
	RecordDNSQuery("A", "NXDOMAIN")

	aCount := testutil.ToFloat64(dnsQueries.WithLabelValues("A", "NOERROR"))
	assert.Equal(t, initialACount+2, aCount, "Expected A NOERROR queries to increase by 2")

	aaaaCount := testutil.ToFloat64(dnsQueries.WithLabelValues("AAAA", "NOERROR"))
	assert.Equal(t, initialAAAACount+1, aaaaCount, "Expected AAAA NOERROR queries to increase by 1")

	nxdomainCount := testutil.ToFloat64(dnsQueries.WithLabelValues("A", "NXDOMAIN"))
	assert.Equal(t, initialNXDOMAINCount+1, nxdomainCount, "Expected A NXDOMAIN queries to increase by 1")
}
