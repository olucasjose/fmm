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
		"flag_limit":   "Limit the number of mirrors tested (e.g. 5 or 2,5 for mint,base)",
		"flag_mirrors": "Select specific mirrors by URL/Name",
		"flag_country": "Filter by one or more countries",
		"flag_cont":    "Filter by one or more regions (e.g. Americas, Europe)",
		"flag_apply":   "Apply the fastest mirrors to the sources list",
		"flag_update":  "Run apt-get update after applying (requires --apply)",
		"flag_target":  "Set a target speed (e.g. 1mb/s or 1mb/s,500kb/s for mint,base)",
		"flag_errs":    "Show exact reasons for mirror failures",
		"flag_quiet":          "Run silently",
		"flag_viable":         "Number of viable mirrors to find (default 5, e.g. 3 or 3,7 for mint,base)",
		"flag_ranking":        "Show the current mirror ranking",
		"interrupted":         "Process interrupted by user. Cleaning up...",
		"unreachable":         "Unreachable",
		"obsolete":            "Obsolete/Stale",
		"testing":             "Testing",
		"ranking_country":     "Detected country",
		"target_reached":      "Target reached",
		"rehab_testing":       "Rehabilitation test:",
		"rehab_result":        "Rehab result",
		"final_results":       "Final Results",
		"best_mint":           "Best Mint",
		"best_base":           "Best Base",
		"no_mint_found":       "No valid Mint mirror found.",
		"no_base_found":       "No valid Base mirror found.",
		"viable_summary":      "Viable mirrors found",
		"ranking_desc":        "Manage the mirror ranking.",
		"ranking_reset_desc":  "Reset the ranking data.",
		"ranking_reset_done":  "Ranking reset successfully.",
		"ranking_not_found":   "No ranking file found. Nothing to reset.",
		"ranking_header":      "Mirror Ranking",
		"ranking_empty":       "No ranking data yet. Run 'fmm run' first.",
		"ranking_updated":     "Updated",
		"press_enter_stop":   "Press [Enter] to stop and use the best mirrors found.",
		"stopped_by_user":    "Stopped. Using best mirrors found.",
	},
	PT: {
		"root_desc":    "Fastest Mint Mirror - Teste e aplique mirrors do Linux Mint.",
		"run_desc":     "Executa testes de velocidade nos mirrors.",
		"list_desc":    "Lista os mirrors disponíveis.",
		"flag_limit":   "Limita a quantidade de mirrors testados (ex: 5 ou 2,5 para mint,base)",
		"flag_mirrors": "Seleciona mirrors específicos por URL/Nome",
		"flag_country": "Filtra por um ou mais países",
		"flag_cont":    "Filtra por uma ou mais regiões (ex: Americas, Europe)",
		"flag_apply":   "Aplica os mirrors mais rápidos no sources.list",
		"flag_update":  "Executa apt-get update após aplicar (exige --apply)",
		"flag_target":  "Define uma meta de velocidade (ex: 1mb/s ou 1mb/s,500kb/s)",
		"flag_errs":    "Mostra motivos exatos de erros em mirrors",
		"flag_quiet":          "Roda os testes de forma silenciosa",
		"flag_viable":         "Quantidade de mirrors viáveis a encontrar (padrão 5, ex: 3 ou 3,7 para mint,base)",
		"flag_ranking":        "Exibe o ranking atual dos mirrors",
		"interrupted":         "Processo interrompido pelo usuário. Limpando dados seguros...",
		"unreachable":         "Inacessível",
		"obsolete":            "Desatualizado",
		"testing":             "Testando",
		"ranking_country":     "País detectado",
		"target_reached":      "Meta atingida",
		"rehab_testing":       "Teste de reabilitação:",
		"rehab_result":        "Resultado rehab",
		"final_results":       "Resultados Finais",
		"best_mint":           "Melhor Mint",
		"best_base":           "Melhor Base",
		"no_mint_found":       "Nenhum mirror Mint válido encontrado.",
		"no_base_found":       "Nenhum mirror Base válido encontrado.",
		"viable_summary":      "Mirrors viáveis encontrados",
		"ranking_desc":        "Gerencia o ranking de mirrors.",
		"ranking_reset_desc":  "Reseta os dados do ranking.",
		"ranking_reset_done":  "Ranking resetado com sucesso.",
		"ranking_not_found":   "Nenhum arquivo de ranking encontrado. Nada a resetar.",
		"ranking_header":      "Ranking de Mirrors",
		"ranking_empty":       "Sem dados de ranking ainda. Execute 'fmm run' primeiro.",
		"ranking_updated":     "Atualizado em",
		"press_enter_stop":   "Pressione [Enter] para parar e usar os melhores mirrors encontrados.",
		"stopped_by_user":    "Parado. Usando melhores mirrors encontrados.",
	},
	ES: {
		"root_desc":    "Fastest Mint Mirror - Pruebe y aplique mirrors de Linux Mint.",
		"run_desc":     "Ejecuta pruebas de velocidad en mirrors.",
		"list_desc":    "Enumera los mirrors disponibles.",
		"flag_limit":   "Limita el número de mirrors probados (ej: 5 o 2,5 para mint,base)",
		"flag_mirrors": "Selecciona mirrors específicos por URL/Nombre",
		"flag_country": "Filtra por uno o más países",
		"flag_cont":    "Filtra por una o más regiones (ej: Americas, Europe)",
		"flag_apply":   "Aplica los mirrors más rápidos al sources.list",
		"flag_update":  "Ejecuta apt-get update después de aplicar (requiere --apply)",
		"flag_target":  "Establece una velocidad objetivo (ej: 1mb/s o 1mb/s,500kb/s)",
		"flag_errs":    "Muestra motivos exactos de fallas en mirrors",
		"flag_quiet":          "Ejecuta silenciosamente",
		"flag_viable":         "Cantidad de mirrors viables a encontrar (por defecto 5, ej: 3 o 3,7 para mint,base)",
		"flag_ranking":        "Muestra el ranking actual de mirrors",
		"interrupted":         "Proceso interrumpido por el usuario. Limpiando...",
		"unreachable":         "Inaccesible",
		"obsolete":            "Obsoleto",
		"testing":             "Probando",
		"ranking_country":     "País detectado",
		"target_reached":      "Meta alcanzada",
		"rehab_testing":       "Prueba de rehabilitación:",
		"rehab_result":        "Resultado rehab",
		"final_results":       "Resultados Finales",
		"best_mint":           "Mejor Mint",
		"best_base":           "Mejor Base",
		"no_mint_found":       "Ningún mirror Mint válido encontrado.",
		"no_base_found":       "Ningún mirror Base válido encontrado.",
		"viable_summary":      "Mirrors viables encontrados",
		"ranking_desc":        "Administra el ranking de mirrors.",
		"ranking_reset_desc":  "Resetea los datos del ranking.",
		"ranking_reset_done":  "Ranking reseteado exitosamente.",
		"ranking_not_found":   "Ningún archivo de ranking encontrado. Nada que resetear.",
		"ranking_header":      "Ranking de Mirrors",
		"ranking_empty":       "Sin datos de ranking aún. Ejecute 'fmm run' primero.",
		"ranking_updated":     "Actualizado",
		"press_enter_stop":   "Presione [Enter] para detener y usar los mejores mirrors encontrados.",
		"stopped_by_user":    "Detenido. Usando mejores mirrors encontrados.",
	},
}

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

func T(key string) string {
	if val, ok := dictionary[currentLang][key]; ok {
		return val
	}
	if val, ok := dictionary[EN][key]; ok {
		return val
	}
	return key
}
