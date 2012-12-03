package main

import (
	"fmt"
	"strings"
)

func (s *service) toDot(name string) string {
	dot := fmt.Sprintf("digraph %s {", name)
	dot = fmt.Sprintf("%s\n\tmargin=\"0\"", dot)
	dot = fmt.Sprintf("%s\n\tsize=\"4.5,7.1\"", dot)
	dot = fmt.Sprintf("%s\n\tpage=\"324,12\"", dot)
	dot = fmt.Sprintf("%s\n\tnode [fontname=\"Alegreya\" fontsize=\"9\"]", dot)
	dot = fmt.Sprintf("%s\n", dot)

	for cname, collection := range s.collections {
		columns := collection.key
		for _, column := range collection.data {
			columns = append(columns, column)
		}

		dot = fmt.Sprintf("%s\n\t%s [shape=record,label=\"%s\\n(%s)|{%s}\"] // %s", dot, cname, cname, collection.ctype, strings.Join(columns, " | "), collection.source)
	}

	for rule_num, rule := range s.rules {
		rule_name := fmt.Sprintf("rule%d", rule_num)
		dot = fmt.Sprintf("%s\n\n\t%s [shape=diamond,label=\"rule %d\"] // %s", dot, rule_name, rule_num, rule.source)

		for _, collection := range rule.requires() {
			dot = fmt.Sprintf("%s\n\t%s -> %s", dot, collection, rule_name)
		}

		dot = fmt.Sprintf("%s\n\t%s -> %s", dot, rule_name, rule.supplies)
	}

	return fmt.Sprintf("%s\n}", dot)
}
