package domain

import (
	"testing"
)

func TestFilterMirrors(t *testing.T) {
	mirrors := []Mirror{
		{URL: "http://m1", Country: "BR", Name: "Mirror 1"},
		{URL: "http://m2", Country: "US", Name: "Mirror 2"},
		{URL: "http://m3", Country: "BR", Name: "Mirror 3"},
	}

	// Testando Limit e Country
	filtered := FilterMirrors(mirrors, []string{"BR"}, []string{}, 1)
	if len(filtered) != 1 {
		t.Fatalf("Esperava 1 mirror, obteve %d", len(filtered))
	}
	if filtered[0].URL != "http://m1" {
		t.Errorf("Esperava http://m1, obteve %v", filtered[0].URL)
	}

	// Testando Ignore Case
	filteredCase := FilterMirrors(mirrors, []string{"us"}, []string{}, 0)
	if len(filteredCase) != 1 {
		t.Fatalf("Esperava 1 mirror após filtro ignore-case, obteve %d", len(filteredCase))
	}
}
