// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package engine

import (
	"testing"
)

func TestFormatSpeed(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"Bytes", 500, "500 B/s"},
		{"KiloBytes exact", 1024, "1.0 KB/s"},
		{"KiloBytes fractional", 1536, "1.5 KB/s"},
		{"MegaBytes exact", 1048576, "1.0 MB/s"},
		{"MegaBytes fractional", 2621440, "2.5 MB/s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Sem t.Parallel() para garantir concorrência zero exigida no escopo
			got := FormatSpeed(tt.input)
			if got != tt.expected {
				t.Errorf("FormatSpeed(%v) = %v; esperado %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseTargetSpeed(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"Empty", "", 0},
		{"MegaBytes abbreviation", "1.5mb", 1.5 * 1024 * 1024},
		{"MegaBytes full", "2mb/s", 2 * 1024 * 1024},
		{"KiloBytes abbreviation", "500kb", 500 * 1024},
		{"KiloBytes full", "800kb/s", 800 * 1024},
		{"Invalid String", "fast", 0},
		{"Missing metric", "100", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTargetSpeed(tt.input)
			if got != tt.expected {
				t.Errorf("ParseTargetSpeed(%v) = %v; esperado %v", tt.input, got, tt.expected)
			}
		})
	}
}
