package location_utils

import (
	"net/http"
	"strings"
)

// Location HTTP 请求的 location 信息
type Location struct {
	Scheme string
	Host   string
}

// GetLocation 获取 HTTP 请求的 location 信息
func GetLocation(r *http.Request) (location Location) {
	location.Scheme = resolveScheme(r)
	location.Host = resolveHost(r)
	return
}

func resolveScheme(r *http.Request) string {
	switch {
	case r.Header.Get("X-Forwarded-Proto") == "https":
		return "https"
	case r.URL.Scheme == "https":
		return "https"
	case r.TLS != nil:
		return "https"
	case strings.HasPrefix(r.Proto, "HTTPS"):
		return "https"
	default:
		return "http"
	}
}

func resolveHost(r *http.Request) (host string) {
	switch {
	case r.Header.Get("X-Forwarded-Host") != "":
		return r.Header.Get("X-Forwarded-Host")
	case r.Header.Get("X-Host") != "":
		return r.Header.Get("X-Host")
	case r.Host != "":
		return r.Host
	case r.URL.Host != "":
		return r.URL.Host
	default:
		return "localhost"
	}
}
