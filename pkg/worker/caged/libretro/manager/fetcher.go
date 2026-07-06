package manager

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/giongto35/cloud-game/v3/pkg/logger"
)

const (
	maxRetries   = 3
	retryBackoff = 500 * time.Millisecond
)

type byteRange struct{ start, end int64 }

type Fetcher struct {
	Client      *http.Client
	MinPartSize int64
	MaxParts    int
	UserAgent   string
	Parallelism int
	log         *logger.Logger
}

func NewFetcher(log *logger.Logger) Fetcher {
	return Fetcher{
		Client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				Proxy:           http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		Parallelism: 3,
		MinPartSize: 10 << 20,
		MaxParts:    3,
		UserAgent:   "Cloud-Game/3.1",
		log:         log,
	}
}

func (f Fetcher) Request(ctx context.Context, dest string, urls ...Download) <-chan Result {
	ch := make(chan Result)
	go func() {
		defer close(ch)
		var wg sync.WaitGroup
		sema := make(chan struct{}, f.Parallelism)
		for _, dl := range urls {
			if ctx.Err() != nil {
				break
			}
			wg.Go(func() {
				select {
				case sema <- none:
				case <-ctx.Done():
					return
				}
				defer func() { <-sema }()
				path, status, err := f.download(ctx, dest, dl)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					f.log.Error().Err(err).Msgf("download [%s] %s has failed", dl.Key, dl.Address)
				} else {
					f.log.Info().Msgf("Downloaded [%s] -> %s", dl.Key, path)
				}
				ch <- Result{Key: dl.Key, Filename: path, Err: err, StatusCode: status}
			})
		}
		wg.Wait()
	}()
	return ch
}

func (f Fetcher) download(ctx context.Context, dest string, dl Download) (string, int, error) {
	var path string
	var status int
	_, err := retry(ctx, f.log, dl.Key, func() (struct{}, error) {
		p, s, err := f.try(ctx, dest, dl)
		status = s
		if err == nil {
			path = p
			return none, nil
		}
		if s == http.StatusNotFound || s == http.StatusGone {
			return none, nonRetryable{err}
		}
		return none, err
	})
	if err != nil {
		return "", status, err
	}
	return path, status, nil
}

func (f Fetcher) try(ctx context.Context, dest string, dl Download) (string, int, error) {
	size, ranges, err := f.head(ctx, dl)
	if err != nil {
		return f.downloadSingle(ctx, dest, dl, -1)
	}
	outPath := filepath.Join(dest, filename(dl))
	if size > 0 {
		if fi, err := os.Stat(outPath); err == nil && fi.Size() == size {
			return outPath, http.StatusOK, nil
		}
	}
	if ranges && size > int64(f.MinPartSize) {
		return f.downloadMulti(ctx, dest, dl, size)
	}
	return f.downloadSingle(ctx, dest, dl, size)
}

func (f Fetcher) head(ctx context.Context, dl Download) (size int64, ranges bool, err error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodHead, dl.Address, nil)
	req.Header.Set("User-Agent", "Cloud-Game/3.0")
	resp, err := f.Client.Do(req)
	if err != nil {
		return 0, false, err
	}
	resp.Body.Close()
	return resp.ContentLength, resp.Header.Get("Accept-Ranges") == "bytes", nil
}

func (f Fetcher) downloadSingle(ctx context.Context, dest string, dl Download, knownSize int64) (string, int, error) {
	outPath := filepath.Join(dest, filename(dl))
	var resumeAt int64
	if fi, err := os.Stat(outPath); err == nil {
		if knownSize > 0 && fi.Size() == knownSize {
			return outPath, http.StatusOK, nil
		}
		resumeAt = fi.Size()
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, dl.Address, nil)
	req.Header.Set("User-Agent", f.UserAgent)
	if resumeAt > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeAt))
	}
	resp, err := f.Client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return "", resp.StatusCode, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	flag := os.O_CREATE | os.O_WRONLY
	if resumeAt > 0 && resp.StatusCode == http.StatusPartialContent {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}
	out, err := os.OpenFile(outPath, flag, 0644)
	if err != nil {
		return "", resp.StatusCode, fmt.Errorf("open file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", resp.StatusCode, fmt.Errorf("write file: %w", err)
	}
	return outPath, resp.StatusCode, nil
}

func (f Fetcher) downloadMulti(ctx context.Context, dest string, dl Download, size int64) (string, int, error) {
	partSize := max(size/int64(f.MaxParts), f.MinPartSize)
	var ranges []byteRange
	for start := int64(0); start < size; start += partSize {
		end := start + partSize - 1
		if end >= size {
			end = size - 1
		}
		ranges = append(ranges, byteRange{start, end})
		if end == size-1 {
			break
		}
	}
	if len(ranges) <= 1 {
		return f.downloadSingle(ctx, dest, dl, size)
	}

	parts := make([][]byte, len(ranges))
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		firstErr error
	)
	for i, r := range ranges {
		wg.Go(func() {
			data, err := retry(ctx, f.log, "", func() ([]byte, error) {
				return f.fetchRange(ctx, dl, r.start, r.end)
			})
			mu.Lock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
			} else {
				parts[i] = data
			}
			mu.Unlock()
		})
	}
	wg.Wait()
	if firstErr != nil {
		return "", 0, fmt.Errorf("multipart: %w", firstErr)
	}
	if err := writeFile(dest, dl, parts); err != nil {
		return "", 0, err
	}
	return filepath.Join(dest, filename(dl)), http.StatusOK, nil
}

func (f Fetcher) fetchRange(ctx context.Context, dl Download, start, end int64) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, dl.Address, nil)
	req.Header.Set("User-Agent", "Cloud-Game/3.0")
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	return buf.Bytes(), nil
}

func writeFile(dest string, dl Download, parts [][]byte) error {
	out, err := os.Create(filepath.Join(dest, filename(dl)))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()
	for _, p := range parts {
		if _, err := out.Write(p); err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}
	return nil
}

func filename(dl Download) string {
	name := filepath.Base(dl.Address)
	if name == "" || name == "." || name == "/" {
		return dl.Key
	}
	return name
}

func retry[T any](ctx context.Context, log *logger.Logger, label string, fn func() (T, error)) (T, error) {
	var (
		lastErr error
		zero    T
	)
	for attempt := range maxRetries {
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}
		if attempt > 0 {
			select {
			case <-time.After(retryBackoff * time.Duration(attempt)):
			case <-ctx.Done():
				return zero, ctx.Err()
			}
		}
		val, err := fn()
		if err == nil {
			return val, nil
		}
		if nr, ok := err.(nonRetryable); ok {
			return zero, nr.err
		}
		lastErr = err
		if label != "" {
			log.Warn().Msgf("%s attempt %d/%d failed: %v", label, attempt+1, maxRetries, err)
		}
	}
	return zero, fmt.Errorf("after %d retries: %w", maxRetries, lastErr)
}

var none struct{}

type nonRetryable struct{ err error }

func (nr nonRetryable) Error() string { return nr.err.Error() }
