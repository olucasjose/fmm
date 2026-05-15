package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//go:embed locales/*.json
var localesFS embed.FS

type Language string

const (
	EN Language = "en"
	PT Language = "pt"
	ES Language = "es"
)

var currentLang = EN

var dictionary = map[Language]map[string]string{}

func Init() {
	// Detecta idioma do sistema
	langEnv := strings.ToLower(os.Getenv("LANG"))

	if strings.HasPrefix(langEnv, "pt") {
		currentLang = PT
	} else if strings.HasPrefix(langEnv, "es") {
		currentLang = ES
	} else {
		currentLang = EN
	}

	// Carrega todos os arquivos de tradução embutidos
	for _, lang := range []Language{EN, PT, ES} {
		data, err := localesFS.ReadFile(fmt.Sprintf("locales/%s.json", lang))
		if err != nil {
			continue
		}

		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			continue
		}

		dictionary[lang] = translations
	}
}

// T retorna a tradução para a chave informada no idioma atual.
// Aceita argumentos variádicos para interpolação via fmt.Sprintf.
// Se a chave não existir no idioma atual, faz fallback para EN.
// Se não existir em nenhum idioma, retorna a própria chave.
func T(key string, args ...interface{}) string {
	template := key

	if val, ok := dictionary[currentLang][key]; ok {
		template = val
	} else if val, ok := dictionary[EN][key]; ok {
		template = val
	}

	if len(args) > 0 {
		return fmt.Sprintf(template, args...)
	}

	return template
}
