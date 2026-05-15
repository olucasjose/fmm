// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package geo

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

type geoResponse struct {
	CountryCode string `json:"country_code"`
}

// DetectLocalCountry tenta buscar o ISO do país por GeoIP. Se falhar, usa a var env LANG.
func DetectLocalCountry() string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.ip2location.io")
	if err == nil {
		defer resp.Body.Close()
		var geo geoResponse
		if err := json.NewDecoder(resp.Body).Decode(&geo); err == nil {
			if geo.CountryCode != "" && geo.CountryCode != "None" {
				return geo.CountryCode
			}
		}
	}

	// Fallback para env LANG (Ex: pt_BR.UTF-8 -> BR)
	langEnv := os.Getenv("LANG")
	if langEnv != "" {
		parts := strings.Split(langEnv, ".")
		if len(parts) > 0 {
			subParts := strings.Split(parts[0], "_")
			if len(subParts) > 1 {
				return strings.ToUpper(subParts[1])
			}
		}
	}

	return "US" // Fallback universal da base original do mintsources
}
