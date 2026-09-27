// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package system

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fmm/internal/parser"
)

func TestExtractOptionalComponentsOnlyPreservesDeclaredNames(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "official-package-repositories.list")
	content := "deb https://fastly.linuxmint.io zara main upstream import backport backport romeo romeo\n"
	if err := os.WriteFile(sourcePath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	components, err := extractOptionalComponents(sourcePath, "zara", []string{"romeo"})
	if err != nil {
		t.Fatalf("extractOptionalComponents retornou erro: %v", err)
	}
	if components != "romeo" {
		t.Errorf("componentes = %q; esperado %q", components, "romeo")
	}
}

func TestRenderSourcesTemplateNormalizesBackport(t *testing.T) {
	config := &parser.MintConfig{Codename: "zara", BaseCodename: "noble"}
	template := "deb $mirror $codename main upstream import backport $optionalcomponents\n" +
		"deb $basemirror $basecodename main restricted universe multiverse\n"

	result := renderSourcesTemplate(template, config, "https://mint.example", "https://base.example", "romeo")
	if strings.Count(result, "backport") != 1 {
		t.Errorf("backport não foi normalizado: %q", result)
	}
	if strings.Count(result, "romeo") != 1 {
		t.Errorf("romeo não foi preservado uma única vez: %q", result)
	}
	if strings.Contains(result, "$") {
		t.Errorf("template ainda contém placeholders: %q", result)
	}
}
