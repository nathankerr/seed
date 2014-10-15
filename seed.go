package seed

import (
	"fmt"
)

// FromSeed loads a Seed in the Seed format from the input.
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

	return p.Seed, nil
}

// ToSeed encodes a Seed in the Seed format.
func ToSeed(seed *Seed, name string) ([]byte, error) {
	info()
	var model string

	for cname, collection := range seed.Collections {
		model = fmt.Sprintf("%s%s\t\n", model, collection.String(cname))
	}

	model += "\n"

	for ruleNum, rule := range seed.Rules {
		model = fmt.Sprintf("%s%s\t# rule %d\n", model, rule, ruleNum)
	}

	return []byte(model), nil
}
