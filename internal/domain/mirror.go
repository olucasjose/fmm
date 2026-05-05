package domain

type MirrorType string

const (
	TypeMint MirrorType = "mint"
	TypeBase MirrorType = "base"
)

// Mirror representa um espelho de repositório.
type Mirror struct {
	URL       string
	Country   string // Código ISO do país (ex: BR, US)
	Region    string // Ex: Americas
	Subregion string // Ex: South America
	Name      string
	Type      MirrorType
}
