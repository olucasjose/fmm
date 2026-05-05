package domain

import "strings"

// FilterMirrors processa as listas originais de mirrors e aplica os filtros para o Benchmark.
func FilterMirrors(mirrors []Mirror, selectedCountries []string, selectedMirrors []string, limit int) []Mirror {
	var filtered []Mirror

	countriesMap := make(map[string]bool)
	for _, c := range selectedCountries {
		countriesMap[strings.ToUpper(c)] = true
	}

	mirrorsMap := make(map[string]bool)
	for _, m := range selectedMirrors {
		mirrorsMap[strings.ToLower(m)] = true
	}

	for _, m := range mirrors {
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

	if limit > 0 && len(filtered) > limit {
		return filtered[:limit]
	}

	return filtered
}

// FilterForList aplica os filtros geográficos para a exibição no comando 'list'.
func FilterForList(mirrors []Mirror, countries []string, regions []string) []Mirror {
	if len(countries) == 0 && len(regions) == 0 {
		return mirrors
	}

	cMap := make(map[string]bool)
	for _, c := range countries {
		cMap[strings.ToUpper(c)] = true
	}

	rMap := make(map[string]bool)
	for _, r := range regions {
		rMap[strings.ToUpper(r)] = true
	}

	var filtered []Mirror
	for _, m := range mirrors {
		match := false

		// Checa País
		if len(cMap) > 0 && cMap[strings.ToUpper(m.Country)] {
			match = true
		}

		// Checa Região (Match amplo: bate tanto com a Region quanto com a Subregion)
		if len(rMap) > 0 && (rMap[strings.ToUpper(m.Region)] || rMap[strings.ToUpper(m.Subregion)]) {
			match = true
		}

		if match {
			filtered = append(filtered, m)
		}
	}
	return filtered
}
