package domain

// Species constants
const (
	SpeciesDog   = "dog"
	SpeciesCat   = "cat"
	SpeciesOther = "other"
)

// CoatType constants
const (
	CoatTypeHairless = "hairless"
	CoatTypeShort    = "short"
	CoatTypeDouble   = "double"
	CoatTypeCurly    = "curly"
	CoatTypeWire     = "wire"
	CoatTypeLong     = "long"
	CoatTypeUnknown  = "unknown"
)

// PetProfile representa el perfil canónico de una mascota para evaluación de elegibilidad.
// NOTA: No incluimos breed; las reglas se basan en atributos físicos (coat, weight, species).
// WeightKg usa int para simplicidad (gramos si necesitas precisión extra en futuro).
type PetProfile struct {
	Species  string
	WeightKg int
	CoatType string
}
