package domain

import "strings"

// FilterMirrors processa as listas originais de mirrors e aplica os filtros.
func FilterMirrors(mirrors []Mirror, selectedCountries []string, selectedMirrors []string, limit int) []Mirror {
	var filtered []Mirror

	// Otimização: mapas para lookups O(1) ignorando case sensitive.
	countriesMap := make(map[string]bool)
	for _, c := range selectedCountries {
		countriesMap[strings.ToUpper(c)] = true
	}

	mirrorsMap := make(map[string]bool)
	for _, m := range selectedMirrors {
		mirrorsMap[strings.ToLower(m)] = true
	}

	for _, m := range mirrors {
		// Se flags específicas foram passadas
		if len(countriesMap) > 0 || len(mirrorsMap) > 0 {
			countryMatch := len(countriesMap) > 0 && countriesMap[strings.ToUpper(m.Country)]
			urlMatch := len(mirrorsMap) > 0 && mirrorsMap[strings.ToLower(m.URL)]
			nameMatch := len(mirrorsMap) > 0 && mirrorsMap[strings.ToLower(m.Name)]

			if !countryMatch && !urlMatch && !nameMatch {
				continue
			}
		}
		filtered = append(filtered, m)
	}

	// Limita se exigido (> 0)
	if limit > 0 && len(filtered) > limit {
		return filtered[:limit]
	}

	return filtered
}
