// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package parser

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseMintConfig(t *testing.T) {
	input := `
[general]
codename=gigi
base_codename=trixie

[mirrors]
default=http://packages.linuxmint.com
base_default=https://deb.debian.org/debian
mirrors=/usr/share/mint-mirrors/linuxmint.list
base_mirrors=/usr/share/python-apt/templates/debian.mirrors

[optional_component_1]
name=romeo

[optional_component_2]
name=romeo

[unrelated]
name=ignored
`

	config, err := ParseMintConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseMintConfig retornou erro: %v", err)
	}

	if config.Codename != "gigi" || config.BaseCodename != "trixie" {
		t.Fatalf("codenames inesperados: %#v", config)
	}
	if config.MirrorsPath != NewMintMirrorsPath {
		t.Errorf("MirrorsPath = %q; esperado %q", config.MirrorsPath, NewMintMirrorsPath)
	}
	if config.BaseMirrorsPath != "/usr/share/python-apt/templates/debian.mirrors" {
		t.Errorf("BaseMirrorsPath inesperado: %q", config.BaseMirrorsPath)
	}
	if !reflect.DeepEqual(config.OptionalComponents, []string{"romeo"}) {
		t.Errorf("OptionalComponents = %#v; esperado [romeo]", config.OptionalComponents)
	}
}

func TestParseMintConfigWithoutOptionalComponents(t *testing.T) {
	input := `[general]
codename=legacy
base_codename=focal

[mirrors]
default=http://packages.linuxmint.com
base_default=http://archive.ubuntu.com/ubuntu
mirrors=/usr/share/python-apt/templates/LinuxMint.mirrors
base_mirrors=/usr/share/python-apt/templates/Ubuntu.mirrors
`

	config, err := ParseMintConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseMintConfig retornou erro: %v", err)
	}
	if len(config.OptionalComponents) != 0 {
		t.Errorf("esperava nenhum componente opcional, obteve %#v", config.OptionalComponents)
	}
}
