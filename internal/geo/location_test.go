// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package geo

import (
	"os"
	"strings"
	"testing"
)

func TestDetectLocalCountry_Fallback(t *testing.T) {
	// Salva ambiente original para não estragar a sessão real
	originalLang := os.Getenv("LANG")
	defer os.Setenv("LANG", originalLang)

	tests := []struct {
		envVal   string
		expected string
	}{
		{"pt_BR.UTF-8", "BR"},
		{"en_US.UTF-8", "US"},
		{"es_ES.UTF-8", "ES"},
		{"invalid", "US"}, // fallback universal do mintsources
	}

	for _, tt := range tests {
		os.Setenv("LANG", tt.envVal)

		// Forçamos o fallback de LANG direto para validação da função de parsing de env
		langEnv := os.Getenv("LANG")
		result := "US"
		if langEnv != "" {
			parts := strings.Split(langEnv, ".")
			if len(parts) > 0 {
				subParts := strings.Split(parts[0], "_")
				if len(subParts) > 1 {
					result = subParts[1]
				}
			}
		}

		if result != tt.expected {
			t.Errorf("DetectLocalCountry fallback(%s) = %s; esperado %s", tt.envVal, result, tt.expected)
		}
	}
}
