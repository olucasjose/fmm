// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package parser

import (
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
