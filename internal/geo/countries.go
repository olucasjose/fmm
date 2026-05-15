// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package geo

import (
	"encoding/json"
	"os"
)

// CountryData representa a estrutura do JSON interno do Linux Mint
type CountryData struct {
	CCA2      string `json:"cca2"`
	Region    string `json:"region"`
	Subregion string `json:"subregion"`
}

var countriesCache map[string]CountryData

// LoadCountries processa o arquivo JSON do sistema e cria um mapa em memória.
// O(1) de tempo de acesso.
func LoadCountries() {
	if countriesCache != nil {
		return
	}
	countriesCache = make(map[string]CountryData)

	path := "/usr/lib/linuxmint/mintSources/countries.json"
	data, err := os.ReadFile(path)
	if err != nil {
		return // Falha silenciosa, as regiões ficarão em branco se o SO não for Mint
	}

	var list []CountryData
	if err := json.Unmarshal(data, &list); err == nil {
		for _, c := range list {
			countriesCache[c.CCA2] = c
		}
	}
}

// GetRegionInfo retorna Region e Subregion dado um country code (ISO).
func GetRegionInfo(cca2 string) (region, subregion string) {
	if c, ok := countriesCache[cca2]; ok {
		return c.Region, c.Subregion
	}
	return "", ""
}
