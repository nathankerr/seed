package main

import (
	"fmt"
)

func (s *service) toModel(name string) string {
	info()
	var model string

	for cname, collection := range s.Collections {
		model = fmt.Sprintf("%s%s\t#%s\n", model, collection.String(cname), collection.Source)
	}

	model += "\n"

	for rule_num, rule := range s.Rules {
		model = fmt.Sprintf("%s%s\t#%s, rule %d\n", model, rule, rule.Source, rule_num)
	}

	return model
}
