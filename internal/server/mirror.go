package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/pikocloud/pikomirror/internal/dbo"
)

func NewMirror(client *dbo.Queries, bufferSize int, timeout time.Duration) *Mirror {
	return &Mirror{
		client:        client,
		buffer:        make(chan *dbo.RequestCreateParams, bufferSize),
		bufferTimeout: timeout,
	}
}

type Mirror struct {
	client        *dbo.Queries
	buffer        chan *dbo.RequestCreateParams
	bufferTimeout time.Duration
}

func (srv *Mirror) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	remoteIP := realIP(r)
	remoteURL := realURL(r)
	requestID := r.Header.Get("X-Request-Id")
	logger := slog.With("method", r.Method, "path", r.URL.Path, "request_id", requestID, "remote_ip", remoteIP)

	var partialContent bool
	content, err := io.ReadAll(r.Body)
	if errors.Is(err, io.ErrUnexpectedEOF) {
		partialContent = true
		err = nil
		logger.Warn("partial content")
	}
	if err != nil {
		logger.Error("Failed to read body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()

	inputHeaders, err := json.Marshal(r.Header)
	if err != nil {
		logger.Error("failed to marshal input headers", "error", err)
	}

	t := time.NewTimer(srv.bufferTimeout)
	defer t.Stop()

	select {
	case srv.buffer <- &dbo.RequestCreateParams{
		CreatedAt:      time.Now(),
		Domain:         remoteURL.Hostname(),
		Method:         strings.ToUpper(realMethod(r)),
		Path:           remoteURL.Path,
		URL:            remoteURL.String(),
		RequestID:      requestID,
		IP:             remoteIP,
		Headers:        inputHeaders,
		Content:        content,
		PartialContent: partialContent,
	}:
	case <-t.C:
		logger.Error("failed save request into buffer - timeout")
	}
}

func (srv *Mirror) Stream(ctx context.Context, aggregation time.Duration) error {
	ticker := time.NewTicker(aggregation)
	defer ticker.Stop()

	var batch []dbo.RequestCreateParams
	for {
		select {
		case m := <-srv.buffer:
			batch = append(batch, *m)
			continue
		case <-ticker.C:
		case <-ctx.Done():
			return ctx.Err()
		}

		if len(batch) == 0 {
			continue
		}

		started := time.Now()

		res := srv.client.RequestCreate(ctx, batch)
		var errs []error
		res.Exec(func(i int, err error) {
			if err != nil {
				errs = append(errs, fmt.Errorf("insert webhook %s: %w", batch[i].RequestID, err))
			}
		})
		errs = append(errs, res.Close())

		if err := errors.Join(errs...); err != nil {
			slog.Error("failed to save webhook", "error", err, "batch", len(batch))
			continue
		}

		dbDuration := time.Since(started)
		slog.Info("batch saved", "duration", dbDuration, "batch", len(batch))

		batch = nil
	}
}

func realMethod(r *http.Request) string {
	if h := r.Header.Get("X-Forwarded-Method"); h != "" {
		return h
	}
	return r.Method
}

func realURL(r *http.Request) *url.URL {
	var u *url.URL
	if h := r.Header.Get("X-Original-Uri"); h != "" {
		if uri, err := url.Parse(h); err != nil || uri == nil {
			return r.URL
		} else {
			u = uri
		}
	} else {
		u = r.URL
	}

	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		u.Scheme = proto
	} else {
		u.Scheme = "http"
	}

	if host := r.Header.Get("X-Forwarded-Host"); host != "" {
		u.Host = host
	} else {
		u.Host = r.Host
	}

	if port := r.Header.Get("X-Forwarded-Port"); port != "" {
		host, _, _ := net.SplitHostPort(u.Host)
		u.Host = host + ":" + port
	}

	return u
}

// RealIP from request.
// Adapted from https://github.com/tomasen/realip/blob/master/realip.go
func realIP(r *http.Request) netip.Addr {
	xForwardedFor := r.Header.Get("X-Forwarded-For")

	for _, address := range strings.Split(xForwardedFor, ",") {
		address = strings.TrimSpace(address)
		if addr, err := netip.ParseAddr(r.RemoteAddr); err == nil {
			return addr
		}
	}

	if !strings.ContainsRune(r.RemoteAddr, ':') {
		if addr, err := netip.ParseAddr(r.RemoteAddr); err == nil {
			return addr
		}
		return netip.AddrFrom4([4]byte{0, 0, 0, 0})
	}

	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	if addr, err := netip.ParseAddr(remoteIP); err == nil {
		return addr
	}
	return netip.AddrFrom4([4]byte{0, 0, 0, 0})
}
