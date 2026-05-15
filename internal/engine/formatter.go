// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package engine

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// FormatSpeed converte a taxa de bytes/sec para um formato legível.
func FormatSpeed(bytesPerSec float64) string {
	mb := 1024.0 * 1024.0
	kb := 1024.0

	if bytesPerSec >= mb {
		return fmt.Sprintf("%.1f MB/s", bytesPerSec/mb)
	}
	if bytesPerSec >= kb {
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/kb)
	}
	return fmt.Sprintf("%.0f B/s", bytesPerSec)
}

// ParseTargetSpeed converte strings como "1.5mb/s", "500kb" em bytes por segundo.
func ParseTargetSpeed(target string) float64 {
	if target == "" {
		return 0
	}
	target = strings.ToLower(strings.TrimSpace(target))

	re := regexp.MustCompile(`([0-9.]+)\s*([a-z]+)`)
	matches := re.FindStringSubmatch(target)

	if len(matches) < 3 {
		return 0
	}

	val, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	unit := matches[2]
	if strings.Contains(unit, "mb") || strings.Contains(unit, "m") {
		return val * 1024 * 1024
	}
	if strings.Contains(unit, "kb") || strings.Contains(unit, "k") {
		return val * 1024
	}

	return val
}
