package system

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"fmm/internal/parser"
)

const (
	SourcesListPath = "/etc/apt/sources.list.d/official-package-repositories.list"
	BackupPath      = "/etc/apt/sources.list.d/official-package-repositories.list.bak"
)

// extractOptionalComponents lê o arquivo atual em disco e resgata componentes extras (ex: romeo, backports) habilitados pelo usuário.
func extractOptionalComponents(filepath string, codename string) string {
	f, err := os.Open(filepath)
	if err != nil {
		return "" // Se não existir ou falhar, segue sem componentes extras
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Isola a linha do repositório Mint (ignora comentários e mirrors base)
		if !strings.HasPrefix(line, "#") && strings.HasPrefix(line, "deb") && strings.Contains(line, " "+codename+" ") {
			parts := strings.Fields(line)
			var optional []string
			foundCodename := false

			for _, part := range parts {
				if !foundCodename {
					if part == codename {
						foundCodename = true
					}
					continue
				}
				// Captura os componentes extras do usuário, ignorando os padrões core do Mint
				if part != "main" && part != "upstream" && part != "import" {
					optional = append(optional, part)
				}
			}

			if len(optional) > 0 {
				return strings.Join(optional, " ")
			}
			break
		}
	}
	return ""
}

// ApplyMirrors realiza a substituição atômica baseada no template oficial do mint.
func ApplyMirrors(ctx context.Context, config *parser.MintConfig, bestMintURL, bestBaseURL string) error {
	// Checa se os diretórios exigidos existem
	if _, err := os.Stat("/etc/apt/sources.list.d"); os.IsNotExist(err) {
		return fmt.Errorf("diretório /etc/apt/sources.list.d não encontrado")
	}

	templatePath := "/usr/share/mintsources/" + config.Codename + "/official-package-repositories.list"
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("arquivo de template ausente: %s", templatePath)
	}

	// Lê template original
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("erro ao ler template: %v", err)
	}
	templateData := string(data)

	// Recupera componentes opcionais antes de aplicar modificações
	optionalComponents := extractOptionalComponents(SourcesListPath, config.Codename)

	// Substituições idênticas ao código oficial (mintsources.py)
	templateData = strings.ReplaceAll(templateData, "$codename", config.Codename)
	templateData = strings.ReplaceAll(templateData, "$basecodename", config.BaseCodename)
	templateData = strings.ReplaceAll(templateData, "$optionalcomponents", optionalComponents)
	templateData = strings.ReplaceAll(templateData, "$mirror", bestMintURL)
	templateData = strings.ReplaceAll(templateData, "$basemirror", bestBaseURL)

	// Criação de arquivo temporário
	tmpFile, err := os.CreateTemp("/tmp", "fmm-sources-*")
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo temporário: %v", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName) // Garante limpeza se algo falhar

	if _, err := tmpFile.WriteString(templateData); err != nil {
		tmpFile.Close()
		return fmt.Errorf("erro ao escrever no arquivo temporário: %v", err)
	}
	tmpFile.Close()

	// Checa ctx para não quebrar nada se o usuário cancelou
	if ctx.Err() != nil {
		return fmt.Errorf("processo abortado. Nenhuma alteração feita")
	}

	// Backup do atual
	if _, err := os.Stat(SourcesListPath); err == nil {
		if err := copyFile(SourcesListPath, BackupPath); err != nil {
			return fmt.Errorf("falha ao criar backup (%s): %v", BackupPath, err)
		}
	}

	// Rename atômico do POSIX
	if err := os.Rename(tmpName, SourcesListPath); err != nil {
		return fmt.Errorf("falha crítica ao aplicar (os.Rename): %v", err)
	}

	// Permissão estrita pro root gerenciar o sources list
	os.Chmod(SourcesListPath, 0644)

	return nil
}

// UpdateCache invoca 'apt-get update' de forma transparente (ligando stdout).
func UpdateCache(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "apt-get", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Sync()
}
