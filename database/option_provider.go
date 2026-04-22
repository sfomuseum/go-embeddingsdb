package database

type ProviderOption struct {
	Option
	provider string
}

func NewProviderOption(p string) Option {

	o := &ProviderOption{
		provider: p,
	}

	return o
}

func (o *ProviderOption) Type() OptionType {
	return ProviderOptionType
}

func (o *ProviderOption) Provider() string {
	return o.provider
}
