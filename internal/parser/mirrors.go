// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fmm/internal/domain"
	"fmm/internal/geo"
)

const (
	NewMintMirrorsPath    = "/usr/share/mint-mirrors/linuxmint.list"
	LegacyMintMirrorsPath = "/usr/share/python-apt/templates/LinuxMint.mirrors"
)

// ParseMirrorsFile interpreta os arquivos '.mirrors' do Mint e injeta a tag de tipo.
func ParseMirrorsFile(r io.Reader, mType domain.MirrorType) ([]domain.Mirror, error) {
	var mirrors []domain.Mirror
	var currentCountry string

	geo.LoadCountries() // Garante que a base do SO está indexada

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#LOC:") {
			currentCountry = strings.TrimPrefix(line, "#LOC:")
			continue
		}

		if currentCountry != "" && !strings.Contains(line, "ubuntu-ports") {
			elements := strings.Fields(line)
			if len(elements) == 0 {
				continue
			}

			url := elements[0]
			if strings.HasSuffix(url, "/") {
				url = url[:len(url)-1]
			}

			name := url
			if len(elements) > 1 {
				name = strings.Join(elements[1:], " ")
			}

			region, subregion := geo.GetRegionInfo(currentCountry)

			mirrors = append(mirrors, domain.Mirror{
				URL:       url,
				Country:   currentCountry,
				Region:    region,
				Subregion: subregion,
				Name:      name,
				Type:      mType,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return mirrors, nil
}

func uniquePaths(paths ...string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		result = append(result, path)
	}
	return result
}

func openFirstAvailable(paths []string) (*os.File, string, error) {
	for _, path := range paths {
		file, err := os.Open(path)
		if err == nil {
			return file, path, nil
		}
		if !os.IsNotExist(err) {
			return nil, "", fmt.Errorf("não foi possível abrir a lista de mirrors %s: %w", path, err)
		}
	}
	return nil, "", fmt.Errorf("nenhuma lista de mirrors encontrada; caminhos tentados: %s", strings.Join(paths, ", "))
}

func normalizeMirrorURL(url string) string {
	return strings.TrimRight(strings.TrimSpace(url), "/")
}

func ensureDefaultDebianMirror(mirrors []domain.Mirror, basePath, baseDefault string) []domain.Mirror {
	if !strings.EqualFold(filepath.Base(basePath), "debian.mirrors") {
		return mirrors
	}

	defaultURL := normalizeMirrorURL(baseDefault)
	if defaultURL == "" {
		return mirrors
	}
	for _, mirror := range mirrors {
		if normalizeMirrorURL(mirror.URL) == defaultURL {
			return mirrors
		}
	}

	return append(mirrors, domain.Mirror{
		URL:     defaultURL,
		Country: "WD",
		Name:    baseDefault,
		Type:    domain.TypeBase,
	})
}

// LoadMirrors carrega os catálogos indicados pela configuração do sistema.
// Para mirrors Mint, usa os caminhos novo e legado apenas como fallback quando
// o arquivo configurado não existe. O catálogo base nunca é substituído por um
// fallback de outra distribuição.
func LoadMirrors(config *MintConfig) ([]domain.Mirror, []domain.Mirror, error) {
	mintPaths := uniquePaths(config.MirrorsPath, NewMintMirrorsPath, LegacyMintMirrorsPath)
	fMint, _, err := openFirstAvailable(mintPaths)
	if err != nil {
		return nil, nil, err
	}
	defer fMint.Close()
	mintMirrors, err := ParseMirrorsFile(fMint, domain.TypeMint)
	if err != nil {
		return nil, nil, err
	}

	fBase, basePath, err := openFirstAvailable(uniquePaths(config.BaseMirrorsPath))
	if err != nil {
		return nil, nil, err
	}
	defer fBase.Close()
	baseMirrors, err := ParseMirrorsFile(fBase, domain.TypeBase)
	if err != nil {
		return nil, nil, err
	}
	baseMirrors = ensureDefaultDebianMirror(baseMirrors, basePath, config.BaseDefault)

	return mintMirrors, baseMirrors, nil
}
