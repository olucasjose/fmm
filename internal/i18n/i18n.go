package i18n

import (
	"os"
	"strings"
)

type Language string

const (
	EN Language = "en"
	PT Language = "pt"
	ES Language = "es"
)

var currentLang = EN

var dictionary = map[Language]map[string]string{
	EN: {
		"root_desc":    "Fastest Mint Mirror - Benchmark and apply Linux Mint mirrors.",
		"run_desc":     "Run benchmark tests on mirrors.",
		"list_desc":    "List available mirrors.",
		"flag_limit":   "Limit the number of mirrors tested",
		"flag_mirrors": "Select specific mirrors by URL/Name",
		"flag_country": "Filter by one or more countries",
		"flag_cont":    "Filter by one or more continents",
		"flag_apply":   "Apply the fastest mirrors to the sources list",
		"flag_update":  "Run apt-get update after applying (requires --apply)",
		"flag_target":  "Set a target speed (e.g., 1mb/s, 500kb/s) and stop when met",
		"flag_errs":    "Show exact reasons for mirror failures",
		"flag_quiet":   "Run silently",
		"interrupted":  "Process interrupted by user. Cleaning up...",
		"unreachable":  "Unreachable",
		"obsolete":     "Obsolete/Stale",
		"testing":      "Testing",
	},
	PT: {
		"root_desc":    "Fastest Mint Mirror - Teste e aplique mirrors do Linux Mint.",
		"run_desc":     "Executa testes de velocidade nos mirrors.",
		"list_desc":    "Lista os mirrors disponíveis.",
		"flag_limit":   "Limita a quantidade de mirrors testados",
		"flag_mirrors": "Seleciona mirrors específicos por URL/Nome",
		"flag_country": "Filtra por um ou mais países",
		"flag_cont":    "Filtra por um ou mais continentes",
		"flag_apply":   "Aplica os mirrors mais rápidos no sources.list",
		"flag_update":  "Executa apt-get update após aplicar (exige --apply)",
		"flag_target":  "Define uma meta de velocidade (ex: 1mb/s, 500kb/s) e encerra ao atingir",
		"flag_errs":    "Mostra motivos exatos de erros em mirrors",
		"flag_quiet":   "Roda os testes de forma silenciosa",
		"interrupted":  "Processo interrompido pelo usuário. Limpando dados seguros...",
		"unreachable":  "Inacessível",
		"obsolete":     "Desatualizado",
		"testing":      "Testando",
	},
	ES: {
		"root_desc":    "Fastest Mint Mirror - Pruebe y aplique mirrors de Linux Mint.",
		"run_desc":     "Ejecuta pruebas de velocidad en mirrors.",
		"list_desc":    "Enumera los mirrors disponibles.",
		"flag_limit":   "Limita el número de mirrors probados",
		"flag_mirrors": "Selecciona mirrors específicos por URL/Nombre",
		"flag_country": "Filtra por uno o más países",
		"flag_cont":    "Filtra por uno o más continentes",
		"flag_apply":   "Aplica los mirrors más rápidos al sources.list",
		"flag_update":  "Ejecuta apt-get update después de aplicar (requiere --apply)",
		"flag_target":  "Establece una velocidad objetivo (ej: 1mb/s) y se detiene al alcanzarla",
		"flag_errs":    "Muestra motivos exactos de fallas en mirrors",
		"flag_quiet":   "Ejecuta silenciosamente",
		"interrupted":  "Proceso interrumpido por el usuario. Limpiando...",
		"unreachable":  "Inaccesible",
		"obsolete":     "Obsoleto",
		"testing":      "Probando",
	},
}

// Init detecta a linguagem através da variável de ambiente LANG.
func Init() {
	langEnv := os.Getenv("LANG")
	langEnv = strings.ToLower(langEnv)

	if strings.HasPrefix(langEnv, "pt") {
		currentLang = PT
	} else if strings.HasPrefix(langEnv, "es") {
		currentLang = ES
	} else {
		currentLang = EN
	}
}

// T retorna a string traduzida para a chave especificada.
func T(key string) string {
	if val, ok := dictionary[currentLang][key]; ok {
		return val
	}
	// Fallback para EN caso não exista tradução e, por fim, a própria chave.
	if val, ok := dictionary[EN][key]; ok {
		return val
	}
	return key
}
