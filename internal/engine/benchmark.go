package engine

import (
	"context"
	"errors"
	"io"
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

// client global reutilizável para aproveitar Keep-Alive internamente.
var httpClient = &http.Client{
	Timeout: 15 * time.Second,
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

	// 1. Checa se o mirror está atualizado (Staleness)
	if err := checkStaleness(ctx, checkURL, maxAgeDays); err != nil {
		return Result{Mirror: m, Err: err}
	}

	// 2. Faz o teste real de velocidade
	speed, err := measureSpeed(ctx, downloadURL)
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
		// Se não tem header, não conseguimos garantir a idade. Assumimos obsoleto (mesmo default do Mint).
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
	// Lemos o body redirecionando para descarte (não escrevemos em disco) para medir o tráfego em RAM.
	// O buffer interno de io.Copy se encarrega do chunking.
	bytesRead, err := io.Copy(io.Discard, resp.Body)

	// Mesmo que haja erro de timeout no final, medimos a velocidade do que foi baixado.
	elapsed := time.Since(start).Seconds()

	if bytesRead == 0 || elapsed == 0 {
		return 0, errors.New("unreachable")
	}

	return float64(bytesRead) / elapsed, nil
}
