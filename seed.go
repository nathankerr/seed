package seed

import (
	"fmt"
)

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
