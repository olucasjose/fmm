// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fmm/internal/domain"
)

func TestParseMirrorsFile(t *testing.T) {
	mockFile := `
#LOC:BR
http://mirror.ufscar.br/mint packages
http://mirror.unesp.br/mint

#LOC:US
http://linuxmint.com
http://ubuntu-ports.com
`
	reader := strings.NewReader(mockFile)

	mirrors, err := ParseMirrorsFile(reader, domain.TypeMint)
	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	// Deve ter ignorado as linhas vazias e o ubuntu-ports. Restam 3.
	if len(mirrors) != 3 {
		t.Fatalf("Esperava 3 mirrors, obteve %d", len(mirrors))
	}

	// Verifica se a junção URL + Nome funcionou
	if mirrors[0].Name != "packages" {
		t.Errorf("Esperava nome 'packages', obteve '%s'", mirrors[0].Name)
	}

	// Verifica propagação do LOC
	if mirrors[2].Country != "US" {
		t.Errorf("Esperava país US, obteve '%s'", mirrors[2].Country)
	}

	// Verifica o Injetor de Tipo
	if mirrors[0].Type != domain.TypeMint {
		t.Errorf("Esperava type mint, obteve '%s'", mirrors[0].Type)
	}
}

func TestOpenFirstAvailableUsesDeclaredPathFirst(t *testing.T) {
	tempDir := t.TempDir()
	configured := filepath.Join(tempDir, "configured.list")
	fallback := filepath.Join(tempDir, "fallback.list")
	if err := os.WriteFile(configured, []byte("configured"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fallback, []byte("fallback"), 0600); err != nil {
		t.Fatal(err)
	}

	file, selected, err := openFirstAvailable(uniquePaths(configured, fallback))
	if err != nil {
		t.Fatalf("openFirstAvailable retornou erro: %v", err)
	}
	defer file.Close()
	if selected != configured {
		t.Errorf("selecionou %q; esperado %q", selected, configured)
	}
}

func TestOpenFirstAvailableFallsBackWhenConfiguredIsMissing(t *testing.T) {
	tempDir := t.TempDir()
	missing := filepath.Join(tempDir, "missing.list")
	fallback := filepath.Join(tempDir, "fallback.list")
	if err := os.WriteFile(fallback, []byte("fallback"), 0600); err != nil {
		t.Fatal(err)
	}

	file, selected, err := openFirstAvailable(uniquePaths(missing, fallback))
	if err != nil {
		t.Fatalf("openFirstAvailable retornou erro: %v", err)
	}
	defer file.Close()
	if selected != fallback {
		t.Errorf("selecionou %q; esperado fallback %q", selected, fallback)
	}
}

func TestOpenFirstAvailableReportsAllAttemptedPaths(t *testing.T) {
	tempDir := t.TempDir()
	first := filepath.Join(tempDir, "first.list")
	second := filepath.Join(tempDir, "second.list")

	_, _, err := openFirstAvailable([]string{first, second})
	if err == nil {
		t.Fatal("esperava erro quando nenhum catálogo existe")
	}
	if !strings.Contains(err.Error(), first) || !strings.Contains(err.Error(), second) {
		t.Errorf("erro não lista todos os caminhos tentados: %v", err)
	}
}

func TestEnsureDefaultDebianMirror(t *testing.T) {
	mirrors := []domain.Mirror{{URL: "https://example.org/debian", Type: domain.TypeBase}}
	result := ensureDefaultDebianMirror(mirrors, "/templates/debian.mirrors", "https://deb.debian.org/debian/")

	if len(result) != 2 {
		t.Fatalf("esperava 2 mirrors, obteve %d", len(result))
	}
	defaultMirror := result[1]
	if defaultMirror.URL != "https://deb.debian.org/debian" || defaultMirror.Country != "WD" {
		t.Errorf("mirror Debian inesperado: %#v", defaultMirror)
	}

	result = ensureDefaultDebianMirror(result, "/templates/Debian.mirrors", "https://deb.debian.org/debian")
	if len(result) != 2 {
		t.Errorf("mirror padrão Debian foi duplicado: %#v", result)
	}
}

func TestEnsureDefaultDebianMirrorIgnoresUbuntuCatalog(t *testing.T) {
	mirrors := []domain.Mirror{{URL: "https://archive.ubuntu.com/ubuntu", Type: domain.TypeBase}}
	result := ensureDefaultDebianMirror(mirrors, "/templates/Ubuntu.mirrors", "https://deb.debian.org/debian")
	if len(result) != 1 {
		t.Errorf("mirror Debian foi adicionado a um catálogo Ubuntu: %#v", result)
	}
}
