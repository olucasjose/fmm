package parser

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// MintConfig guarda configurações base extraídas.
type MintConfig struct {
	Codename        string
	BaseCodename    string
	BaseDefault     string
	MintDefault     string
	MirrorsPath     string
	BaseMirrorsPath string
}

// ParseMintConfig lê as configurações cruas e retorna o struct de config.
func ParseMintConfig(r io.Reader) (*MintConfig, error) {
	config := &MintConfig{}
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			switch key {
			case "codename":
				config.Codename = val
			case "base_codename":
				config.BaseCodename = val
			case "default":
				config.MintDefault = val
			case "base_default":
				config.BaseDefault = val
			case "mirrors":
				config.MirrorsPath = val
			case "base_mirrors":
				config.BaseMirrorsPath = val
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return config, nil
}

// LoadConfig carrega as configurações direto do disco.
func LoadConfig(codename string) (*MintConfig, error) {
	path := "/usr/share/mintsources/" + codename + "/mintsources.conf"
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseMintConfig(f)
}
