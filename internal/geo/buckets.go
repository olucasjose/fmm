// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package geo

import (
	"sort"

	"fmm/internal/domain"
)

// SortByGeoBuckets classifica mirrors em baldes geográficos hierárquicos,
// reproduzindo a lógica exata do mintsources.py.
//
// Ordem dos baldes: Worldwide → Local → Bordering → Network Neighbors →
// Subregional → Regional → Official.
//
// Mirrors que não se encaixam em nenhum balde são excluídos por default.
// Se hasExplicitFilter == true (usuário passou --countries ou --mirrors),
// os mirrors "other" são incluídos no final em vez de excluídos.
// Fallback de segurança: se < 2 mirrors nos baldes 1-7, retorna todos.
func SortByGeoBuckets(mirrors []domain.Mirror, userCountryCode string, defaultMirrorURL string, hasExplicitFilter bool) []domain.Mirror {
	LoadCountries()

	// Passo 1 — Resolver borders e networkNeighbors (exatamente como o mintsources).
	// Itera todos os países do JSON, checa se o cca3 de cada um está na lista
	// de borders/networkNeighbors do país do usuário, e coleta o cca2.
	borderingCountries := make(map[string]bool)
	networkNeighbors := make(map[string]bool)
	subregionCountries := make(map[string]bool)
	regionCountries := make(map[string]bool)

	localCountry := GetCountryData(userCountryCode)
	if localCountry != nil {
		for _, country := range GetAllCountries() {
			// Checa se o cca3 deste país está na lista borders do usuário
			for _, borderCCA3 := range localCountry.Borders {
				if country.CCA3 == borderCCA3 {
					borderingCountries[country.CCA2] = true
					break
				}
			}
			// Checa se o cca3 deste país está na lista networkNeighbors do usuário
			for _, nnCCA3 := range localCountry.NetworkNeighbors {
				if country.CCA3 == nnCCA3 {
					networkNeighbors[country.CCA2] = true
					break
				}
			}
			// Checa sub-região e região (exatamente como mintsources)
			if country.Region == localCountry.Region {
				if country.Subregion == localCountry.Subregion {
					subregionCountries[country.CCA2] = true
				} else {
					regionCountries[country.CCA2] = true
				}
			}
		}
	}

	// Passo 2 — Classificar cada mirror no balde correspondente.
	var worldwide, local, bordering, networkN, subregional, regional, official, other []domain.Mirror

	for _, m := range mirrors {
		switch {
		case m.Country == "WD":
			worldwide = append(worldwide, m)
		case m.Country == userCountryCode:
			local = append(local, m)
		case borderingCountries[m.Country]:
			bordering = append(bordering, m)
		case networkNeighbors[m.Country]:
			networkN = append(networkN, m)
		case subregionCountries[m.Country]:
			subregional = append(subregional, m)
		case regionCountries[m.Country]:
			regional = append(regional, m)
		case m.URL == defaultMirrorURL:
			official = append(official, m)
		default:
			other = append(other, m)
		}
	}

	// Dentro de cada balde, ordena alfabeticamente por country_code.
	sortByCountry := func(s []domain.Mirror) {
		sort.SliceStable(s, func(i, j int) bool {
			return s[i].Country < s[j].Country
		})
	}
	sortByCountry(worldwide)
	sortByCountry(bordering)
	sortByCountry(networkN)
	sortByCountry(subregional)
	sortByCountry(regional)

	// Concatena os baldes na ordem de prioridade (mintsources).
	visible := make([]domain.Mirror, 0, len(mirrors))
	visible = append(visible, worldwide...)
	visible = append(visible, local...)
	visible = append(visible, bordering...)
	visible = append(visible, networkN...)
	visible = append(visible, subregional...)
	visible = append(visible, regional...)
	visible = append(visible, official...)

	// Fallback de segurança: se < 2 mirrors visíveis, retorna todos.
	if len(visible) < 2 {
		return mirrors
	}

	// Se o usuário passou filtros explícitos (--countries/--mirrors),
	// inclui os "other" no final em vez de excluí-los.
	if hasExplicitFilter {
		sortByCountry(other)
		visible = append(visible, other...)
	}

	return visible
}
