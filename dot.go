package main

import (
	"fmt"
)

func (s *service) toDot(name string) string {
	dot := fmt.Sprintf("digraph %s {", name)

	for rule_num, rule := range s.rules {
		rule_name := fmt.Sprintf("%s%d", name, rule_num)
		dot = fmt.Sprintf("%s\n\n%s [shape=diamond,label=rule]", dot, rule_name)

		for _, collection := range rule.requires() {
			dot = fmt.Sprintf("%s\n\t%s -> %s", dot, collection, rule_name)
		}

		dot = fmt.Sprintf("%s\n\t%s -> %s", dot, rule_name, rule.supplies)
	}

	return fmt.Sprintf("%s\n}", dot)
}
