package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/marcus/avatars/internal/library"
)

func (a *app) request(ctx context.Context, method, path string, body any, dest any) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		b, e := json.Marshal(body)
		if e != nil {
			return nil, e
		}
		reader = bytes.NewReader(b)
	}
	req, e := http.NewRequestWithContext(ctx, method, a.baseURL+path, reader)
	if e != nil {
		return nil, e
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, e := client.Do(req)
	if e != nil {
		return nil, fmt.Errorf("connect to studio: %w", e)
	}
	defer resp.Body.Close()
	b, e := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if e != nil {
		return nil, e
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var envelope struct {
			Error *library.Error `json:"error"`
		}
		if json.Unmarshal(b, &envelope) == nil && envelope.Error != nil {
			return nil, envelope.Error
		}
		return nil, fmt.Errorf("service returned HTTP %d", resp.StatusCode)
	}
	if dest != nil {
		if e := json.Unmarshal(b, dest); e != nil {
			return nil, fmt.Errorf("invalid service response: %w", e)
		}
	}
	return b, nil
}
func (a *app) link(path string) string { return a.baseURL + path }
func (a *app) linkedAvatar(v library.Avatar) library.Avatar {
	v.URL = a.link(v.URL)
	v.SVGURL = a.link(v.SVGURL)
	v.PNGURL = a.link(v.PNGURL)
	return v
}
func (a *app) linkedCollection(v library.Collection) library.Collection {
	v.URL = a.link(v.URL)
	v.Avatars = append([]library.Avatar(nil), v.Avatars...)
	for i := range v.Avatars {
		v.Avatars[i] = a.linkedAvatar(v.Avatars[i])
	}
	return v
}
