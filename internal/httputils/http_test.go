package httputils

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHeaderToString(t *testing.T) {
	expected := `Header1: One
Header2: 2
Header2: Two
Header3: Three
`
	header := http.Header{
		"Header2": []string{"2", "Two"},
		"Header1": []string{"One"},
		"Header3": []string{"Three"},
	}

	assert.Equal(t, expected, HeadersToSortedString(header))
}

func TestGetLogFormatter(t *testing.T) {
	expected := "127.0.0.1 - [01/Nov/0001:00:00:00 +0000] \"GET / HTTP/1.1\" 200 100 1000 local \"golang test 1.0\" \"1.1.1.1, 2.2.2.2\" \"-\"\n"

	h := http.Header{}
	h.Set("User-Agent", "golang test 1.0")
	h.Set("Referer", "local")
	h.Set("X-Forwarded-For", "1.1.1.1, 2.2.2.2")

	r := http.Request{
		Proto:  "HTTP/1.1",
		Header: h,
	}

	p := gin.LogFormatterParams{
		ClientIP:     "127.0.0.1",
		TimeStamp:    time.Time{},
		Method:       "GET",
		Path:         "/",
		StatusCode:   200,
		BodySize:     100,
		Latency:      1000,
		ErrorMessage: "",
		Request:      &r,
	}

	assert.Equal(t, expected, GetLogFormatter(p))
}

func TestNormalizeLog(t *testing.T) {
	assert.Equal(t, "-", normalizeLog(""))
	assert.Equal(t, "string", normalizeLog("string"))
	assert.Equal(t, "-", normalizeLog([]string{}))
	assert.Equal(t, "1.1.1.1", normalizeLog([]string{"1.1.1.1"}))
	assert.Equal(t, "1.1.1.1, 2.2.2.2", normalizeLog([]string{"1.1.1.1", "2.2.2.2"}))
}
