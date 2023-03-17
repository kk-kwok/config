package config

type TextProvider struct {
	ConfigText []byte
}

var _ Provider = &TextProvider{}

func (p *TextProvider) Name() string {
	return "text"
}

func (p *TextProvider) Config(helper *providerHelper) ([]byte, error) {
	content := p.ConfigText
	if len(content) == 0 {
		return nil, ErrEmptyConfig
	}
	return content, nil
}
