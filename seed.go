package seed

import (
	"fmt"
)

func FromSeed(name string, input []byte) (*Seed, error) {
	info()

	p := yyParser{
		Seed: &Seed{
			Name:        name,
			Collections: make(map[string]*Collection),
		},
	}
	p.Init()
	p.ResetBuffer(string(input))

	err := p.Parse(ruleSeed)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%#v\n", p.Seed)

	return p.Seed, nil
}

func ToSeed(seed *Seed, name string) ([]byte, error) {
	info()
	var model string

	for cname, collection := range seed.Collections {
		model = fmt.Sprintf("%s%s\t#%s\n", model, collection.String(cname), collection.Source)
	}

	model += "\n"

	for rule_num, rule := range seed.Rules {
		model = fmt.Sprintf("%s%s\t#%s, rule %d\n", model, rule, rule.Source, rule_num)
	}

	return []byte(model), nil
}
