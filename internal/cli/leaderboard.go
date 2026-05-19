// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"fmt"
	"sort"
	"strings"

	"fmm/internal/engine"
	"fmm/internal/i18n"
	"github.com/pterm/pterm"
)

const leaderboardSize = 10

// leaderboardEntry holds the data for a single row in the live leaderboard.
type leaderboardEntry struct {
	name  string
	speed float64
}

// leaderboard manages a live-updating list of the fastest mirrors seen so far
// during a benchmark run. It renders itself inside a pterm.Area so it updates
// in-place without scrolling the terminal.
type leaderboard struct {
	area       *pterm.AreaPrinter
	entries    []leaderboardEntry
	tested     int
	total      int
	current    string // non-empty while a test is in progress
	mirrorType string
}

// newLeaderboard creates and starts the live area printer.
// The area is configured with RemoveWhenDone so Stop() clears all dynamic
// content, and the final state is printed as static text separately.
func newLeaderboard(mirrorType string, total int) (*leaderboard, error) {
	area, err := pterm.DefaultArea.WithRemoveWhenDone(true).Start()
	if err != nil {
		return nil, err
	}
	lb := &leaderboard{
		area:       area,
		total:      total,
		mirrorType: mirrorType,
	}
	lb.render()
	return lb, nil
}

// setTesting updates the progress counter and redraws.
func (lb *leaderboard) setTesting(name string) {
	lb.current = name
	lb.tested++
	lb.render()
}

// addResult registers a successful result, re-sorts the list, and redraws.
func (lb *leaderboard) addResult(name string, speed float64) {
	lb.entries = append(lb.entries, leaderboardEntry{name: name, speed: speed})
	sort.Slice(lb.entries, func(i, j int) bool {
		return lb.entries[i].speed > lb.entries[j].speed
	})
	lb.render()
}

// stop clears the dynamic area and reprints the final state as static text.
// This two-step approach prevents duplication caused by Enter keypresses
// that shift the area content and confuse pterm's cursor tracking.
func (lb *leaderboard) stop() {
	_ = lb.area.Stop()
	lb.printFinal()
}

// render builds the text frame and updates the area in-place.
func (lb *leaderboard) render() {
	var sb strings.Builder

	// ── Section title with live progress ─────────────────────────────────────
	progress := fmt.Sprintf("(%d/%d)", lb.tested, lb.total)
	title := fmt.Sprintf("# %s %s", i18n.T("benchmarking_section", lb.mirrorType), progress)
	sb.WriteString(pterm.Yellow(title))
	sb.WriteString("\n\n")

	// ── Header ───────────────────────────────────────────────────────────────
	sb.WriteString(fmt.Sprintf(" %-3s  %-38s  %s\n", "#", i18n.T("table_header_name"), i18n.T("table_header_speed")))
	sb.WriteString(pterm.Gray(" " + strings.Repeat("-", 56) + "\n"))

	// ── Rows ─────────────────────────────────────────────────────────────────
	top := lb.entries
	if len(top) > leaderboardSize {
		top = top[:leaderboardSize]
	}

	for i, e := range top {
		rank := fmt.Sprintf("%d.", i+1)
		name := truncate(e.name, 38)
		speed := engine.FormatSpeed(e.speed)

		var row string
		if i == 0 {
			row = pterm.Green(fmt.Sprintf(" %-3s  %-38s  %s", rank, name, speed))
		} else {
			row = fmt.Sprintf(" %-3s  %-38s  %s", rank, name, speed)
		}
		sb.WriteString(row + "\n")
	}

	// Pad empty rows while testing so the area height stays stable.
	if lb.current != "" {
		for i := len(top); i < leaderboardSize; i++ {
			sb.WriteString(pterm.Gray(fmt.Sprintf(" %-3s  %-38s  %s\n", "-", "-", "-")))
		}
	}

	lb.area.Update(sb.String())
}

// printFinal prints the final leaderboard state as static text (not managed
// by pterm.Area), so it cannot be affected by cursor tracking issues.
func (lb *leaderboard) printFinal() {
	progress := fmt.Sprintf("(%d/%d)", lb.tested, lb.total)
	title := fmt.Sprintf("# %s %s", i18n.T("benchmarking_section", lb.mirrorType), progress)
	pterm.Println(pterm.Yellow(title))
	pterm.Println()

	pterm.Println(fmt.Sprintf(" %-3s  %-38s  %s", "#", i18n.T("table_header_name"), i18n.T("table_header_speed")))
	pterm.Println(pterm.Gray(" " + strings.Repeat("-", 56)))

	top := lb.entries
	if len(top) > leaderboardSize {
		top = top[:leaderboardSize]
	}

	for i, e := range top {
		rank := fmt.Sprintf("%d.", i+1)
		name := truncate(e.name, 38)
		speed := engine.FormatSpeed(e.speed)

		if i == 0 {
			pterm.Println(pterm.Green(fmt.Sprintf(" %-3s  %-38s  %s", rank, name, speed)))
		} else {
			pterm.Println(fmt.Sprintf(" %-3s  %-38s  %s", rank, name, speed))
		}
	}
}

// truncate shortens s to at most max runes, appending "…" if needed.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
