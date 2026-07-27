package middleware

import (
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", redactIP(c.ClientIP())),
		}

		if query != "" {
			attrs = append(attrs, slog.String("query", query))
		}
		if rid, ok := c.Get("request_id"); ok {
			attrs = append(attrs, slog.String("request_id", rid.(string)))
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		log.LogAttrs(c.Request.Context(), level, "request", attrs...)
	}
}

func redactIP(addr string) string {
	host := addr
	port := ""

	if h, p, err := net.SplitHostPort(addr); err == nil {
		host = h
		port = p
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return addr
	}

	ip4 := ip.To4()
	if ip4 == nil {
		//return as it is for non-IPv4 addresses
		return addr
	}

	parts := strings.Split(ip4.String(), ".")
	parts[2] = "X"
	parts[3] = "X"
	redacted := strings.Join(parts, ".")

	if port != "" {
		return net.JoinHostPort(redacted, port)
	}
	return redacted
}
