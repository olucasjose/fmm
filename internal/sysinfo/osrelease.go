package sysinfo

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// ParseOSRelease lê um io.Reader e extrai o VERSION_CODENAME.
func ParseOSRelease(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "VERSION_CODENAME=") {
			return strings.TrimPrefix(line, "VERSION_CODENAME="), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}

// GetCodename é o helper que executa a leitura direta do disco.
func GetCodename() (string, error) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return "", err
	}
	defer f.Close()
	return ParseOSRelease(f)
}
