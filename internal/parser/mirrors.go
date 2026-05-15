// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package parser

import (
	"bufio"
	"io"
	"os"
	"strings"

	"fmm/internal/domain"
	"fmm/internal/geo"
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

// LoadMirrors carrega as listas de base e mint a partir dos paths especificados.
func LoadMirrors(mintPath, basePath string) ([]domain.Mirror, []domain.Mirror, error) {
	fMint, err := os.Open(mintPath)
	if err != nil {
		return nil, nil, err
	}
	defer fMint.Close()
	mintMirrors, err := ParseMirrorsFile(fMint, domain.TypeMint)
	if err != nil {
		return nil, nil, err
	}

	fBase, err := os.Open(basePath)
	if err != nil {
		return nil, nil, err
	}
	defer fBase.Close()
	baseMirrors, err := ParseMirrorsFile(fBase, domain.TypeBase)
	if err != nil {
		return nil, nil, err
	}

	return mintMirrors, baseMirrors, nil
}
