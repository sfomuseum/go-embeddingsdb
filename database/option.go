package database

type OptionType uint8

const (
	UndefinedOptionType OptionType = 0 << iota
	DimensionsOptionType
	FilterOptionType
	ProviderOptionType
)

type Option interface {
	Type() OptionType
}
