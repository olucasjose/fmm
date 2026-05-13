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
	// Reproduz FOLLOWLOCATION do pycurl
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("too many redirects")
		}
		return nil
	},
}

// DefaultMirrorInfo guarda a data e idade do mirror default oficial.
// Deve ser obtido UMA VEZ antes de iniciar os benchmarks, via FetchDefaultMirrorInfo.
type DefaultMirrorInfo struct {
	MintDate *time.Time // timestamp do /db/version no mirror default Mint
	MintAge  *int       // dias desde agora até MintDate
	BaseDate *time.Time // timestamp do /ls-lR.gz no mirror default Base
	BaseAge  *int       // dias desde agora até BaseDate
}

// FetchDefaultMirrorInfo consulta a data dos mirrors default (oficial) para
// usar como referência no check de staleness, reproduzindo o comportamento do mintsources.
func FetchDefaultMirrorInfo(ctx context.Context, config *parser.MintConfig) *DefaultMirrorInfo {
	info := &DefaultMirrorInfo{}

	// Mint default
	mintURL := config.MintDefault + "/db/version"
	if t := getURLLastModified(ctx, mintURL); t != nil {
		info.MintDate = t
		age := int(time.Since(*t).Hours() / 24)
		info.MintAge = &age
	}

	// Base default
	baseURL := config.BaseDefault + "/ls-lR.gz"
	if t := getURLLastModified(ctx, baseURL); t != nil {
		info.BaseDate = t
		age := int(time.Since(*t).Hours() / 24)
		info.BaseAge = &age
	}

	return info
}

// TestMirror orquestra a checagem de staleness e de download de velocidade para um mirror.
func TestMirror(ctx context.Context, m domain.Mirror, config *parser.MintConfig, defaultInfo *DefaultMirrorInfo) Result {
	var checkURL, downloadURL string
	var maxAgeDays int
	var defaultDate *time.Time
	var skipStalenessCheck bool

	if m.Type == domain.TypeMint {
		checkURL = m.URL + "/db/version"
		downloadURL = m.URL + "/dists/" + config.Codename + "/main/Contents-amd64.gz"
		maxAgeDays = MintMaxAgeDays
		defaultDate = defaultInfo.MintDate

		// Reproduz check_mint_mirror_up_to_date do mintsources:
		// "If the default server was updated recently, the age is irrelevant"
		if defaultInfo.MintAge == nil || *defaultInfo.MintAge < MintMaxAgeDays {
			skipStalenessCheck = true
		}
	} else {
		checkURL = m.URL + "/ls-lR.gz"
		downloadURL = m.URL + "/dists/" + config.BaseCodename + "/main/binary-amd64/Packages.gz"
		maxAgeDays = BaseMaxAgeDays
		defaultDate = defaultInfo.BaseDate

		// Reproduz check_base_mirror_up_to_date do mintsources:
		// "If the default mirror is unavailable it's likely temporary, assume the mirror is ok."
		if defaultInfo.BaseAge == nil {
			skipStalenessCheck = true
		}
	}

	// 1. Checa se o mirror está atualizado (Timeout total de 30s conforme pycurl)
	if !skipStalenessCheck {
		staleCtx, staleCancel := context.WithTimeout(ctx, 30*time.Second)
		defer staleCancel()
		if err := checkStaleness(staleCtx, checkURL, maxAgeDays, defaultDate); err != nil {
			return Result{Mirror: m, Err: err}
		}
	}

	// 2. Faz o teste real de velocidade (Timeout total de 20s conforme pycurl)
	speedCtx, speedCancel := context.WithTimeout(ctx, 20*time.Second)
	defer speedCancel()
	speed, err := measureSpeed(speedCtx, downloadURL)
	return Result{Mirror: m, Speed: speed, Err: err}
}

// getURLLastModified reproduz get_url_last_modified do mintsources:
// usa OPT_FILETIME/INFO_FILETIME do pycurl (equivale ao header Last-Modified).
func getURLLastModified(ctx context.Context, url string) *time.Time {
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "HEAD", url, nil)
	if err != nil {
		return nil
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil
	}

	lastMod := resp.Header.Get("Last-Modified")
	if lastMod == "" {
		return nil
	}

	t, err := time.Parse(http.TimeFormat, lastMod)
	if err != nil {
		return nil
	}

	return &t
}

// checkStaleness compara a idade do mirror com a data do mirror default,
// reproduzindo check_mirror_up_to_date do mintsources:
//
//	mirror_age = (default_mirror_date - mirror_date).days
//	if mirror_age > max_age: obsolete
func checkStaleness(ctx context.Context, url string, maxAge int, defaultDate *time.Time) error {
	mirrorDate := getURLLastModified(ctx, url)
	if mirrorDate == nil {
		// mintsources: "Error: Can't find the age of url" → return False
		return errors.New("unreachable")
	}

	if defaultDate == nil {
		// Sem referência, não podemos comparar — assume ok
		return nil
	}

	// Atraso de sincronização: quantos dias o mirror está atrás do default
	mirrorAge := int(defaultDate.Sub(*mirrorDate).Hours() / 24)
	if mirrorAge > maxAge {
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
