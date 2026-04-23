package options

type OptionType uint8

const (
	UndefinedOptionType OptionType = 0 << iota
	DimensionsOptionType
	FilterOptionType
	ProviderOptionType
	SimilarProviderOptionType
	MaxDistanceOptionType
	MaxResultsOptionType
)

type Option interface {
	Type() OptionType
}
