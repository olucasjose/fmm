// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package system

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fmm/internal/parser"
)

const (
	SourcesListPath = "/etc/apt/sources.list.d/official-package-repositories.list"
	BackupPath      = "/etc/apt/sources.list.d/official-package-repositories.list.bak"
)

// extractOptionalComponents preserva somente os componentes que o
// mintsources.conf declara explicitamente como opcionais.
func extractOptionalComponents(sourcePath, codename string, allowed []string) (string, error) {
	f, err := os.Open(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer f.Close()

	allowedSet := make(map[string]bool, len(allowed))
	for _, component := range allowed {
		if component != "" {
			allowedSet[component] = true
		}
	}

	found := make(map[string]bool, len(allowedSet))
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || !strings.HasPrefix(line, "deb ") {
			continue
		}

		parts := strings.Fields(line)
		codenameIndex := -1
		for i, part := range parts {
			if part == codename {
				codenameIndex = i
				break
			}
		}
		if codenameIndex == -1 {
			continue
		}

		for _, component := range parts[codenameIndex+1:] {
			if allowedSet[component] {
				found[component] = true
			}
		}
		break
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	selected := make([]string, 0, len(found))
	selectedSet := make(map[string]bool, len(found))
	for _, component := range allowed {
		if found[component] && !selectedSet[component] {
			selected = append(selected, component)
			selectedSet[component] = true
		}
	}
	return strings.Join(selected, " "), nil
}

func renderSourcesTemplate(templateData string, config *parser.MintConfig, mintURL, baseURL, optionalComponents string) string {
	templateData = strings.ReplaceAll(templateData, "$codename", config.Codename)
	templateData = strings.ReplaceAll(templateData, "$basecodename", config.BaseCodename)
	templateData = strings.ReplaceAll(templateData, "$optionalcomponents", optionalComponents)
	templateData = strings.ReplaceAll(templateData, "$mirror", mintURL)
	templateData = strings.ReplaceAll(templateData, "$basemirror", baseURL)
	return templateData
}

// ApplyMirrors realiza a substituição atômica baseada no template oficial do mint.
func ApplyMirrors(ctx context.Context, config *parser.MintConfig, bestMintURL, bestBaseURL string) error {
	// Checa se os diretórios exigidos existem
	targetDir := filepath.Dir(SourcesListPath)
	if _, err := os.Stat(targetDir); err != nil {
		return fmt.Errorf("diretório %s indisponível: %w", targetDir, err)
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
	optionalComponents, err := extractOptionalComponents(SourcesListPath, config.Codename, config.OptionalComponents)
	if err != nil {
		return fmt.Errorf("erro ao ler componentes opcionais: %w", err)
	}

	templateData = renderSourcesTemplate(templateData, config, bestMintURL, bestBaseURL, optionalComponents)

	// O temporário fica no diretório de destino para que o rename continue
	// atômico mesmo quando /tmp estiver em outro sistema de arquivos.
	tmpFile, err := os.CreateTemp(targetDir, ".fmm-sources-*")
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo temporário: %v", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName) // Garante limpeza se algo falhar

	if _, err := tmpFile.WriteString(templateData); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("erro ao escrever no arquivo temporário: %v", err)
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("erro ao sincronizar arquivo temporário: %v", err)
	}
	if err := tmpFile.Chmod(0644); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("erro ao ajustar permissões do arquivo temporário: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("erro ao fechar arquivo temporário: %v", err)
	}

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
