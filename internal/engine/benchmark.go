package engine

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"fmm/internal/domain"
	"fmm/internal/parser"
)

type Result struct {
	Mirror domain.Mirror
	Speed  float64
	Err    error
}

const (
	MintMaxAgeDays = 2
	BaseMaxAgeDays = 14
)

// client configurado exatamente como o pycurl do mintsources.py
var httpClient = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second, // pycurl CONNECTTIMEOUT
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

// TestMirror orquestra a checagem de staleness e de download de velocidade para um mirror.
func TestMirror(ctx context.Context, m domain.Mirror, config *parser.MintConfig) Result {
	var checkURL, downloadURL string
	var maxAgeDays int

	if m.Type == domain.TypeMint {
		checkURL = m.URL + "/db/version"
		downloadURL = m.URL + "/dists/" + config.Codename + "/main/Contents-amd64.gz"
		maxAgeDays = MintMaxAgeDays
	} else {
		checkURL = m.URL + "/ls-lR.gz"
		downloadURL = m.URL + "/dists/" + config.BaseCodename + "/main/binary-amd64/Packages.gz"
		maxAgeDays = BaseMaxAgeDays
	}

	// 1. Checa se o mirror está atualizado (Timeout total de 30s conforme pycurl)
	staleCtx, staleCancel := context.WithTimeout(ctx, 30*time.Second)
	defer staleCancel()
	if err := checkStaleness(staleCtx, checkURL, maxAgeDays); err != nil {
		return Result{Mirror: m, Err: err}
	}

	// 2. Faz o teste real de velocidade (Timeout total de 20s conforme pycurl)
	speedCtx, speedCancel := context.WithTimeout(ctx, 20*time.Second)
	defer speedCancel()
	speed, err := measureSpeed(speedCtx, downloadURL)
	return Result{Mirror: m, Speed: speed, Err: err}
}

func checkStaleness(ctx context.Context, url string, maxAge int) error {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.New("unreachable")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.New("unreachable")
	}

	lastMod := resp.Header.Get("Last-Modified")
	if lastMod == "" {
		return errors.New("obsolete")
	}

	t, err := time.Parse(http.TimeFormat, lastMod)
	if err != nil {
		return errors.New("obsolete")
	}

	ageDays := time.Since(t).Hours() / 24
	if ageDays > float64(maxAge) {
		return errors.New("obsolete")
	}

	return nil
}

func measureSpeed(ctx context.Context, url string) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, errors.New("unreachable")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, errors.New("unreachable")
	}

	start := time.Now()
	bytesRead, err := io.Copy(io.Discard, resp.Body)

	elapsed := time.Since(start).Seconds()

	if bytesRead == 0 || elapsed == 0 {
		return 0, errors.New("unreachable")
	}

	return float64(bytesRead) / elapsed, nil
}
