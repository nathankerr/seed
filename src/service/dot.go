package service

import (
	"fmt"
	"strings"
)

func (s *Service) ToDot(name string) string {
	info()

	dot := fmt.Sprintf("digraph %s {", name)
	dot = fmt.Sprintf("%s\n\tmargin=\"0\"", dot)
	dot = fmt.Sprintf("%s\n\tsize=\"4.5,7.1\"", dot)
	dot = fmt.Sprintf("%s\n\tpage=\"324,12\"", dot)
	dot = fmt.Sprintf("%s\n\tnode [fontname=\"Alegreya\" fontsize=\"9\"]", dot)
	dot = fmt.Sprintf("%s\n", dot)

	for cname, collection := range s.Collections {
		columns := collection.Key
		for _, column := range collection.Data {
			columns = append(columns, column)
		}

		dot = fmt.Sprintf("%s\n\t%s [shape=record,label=\"%s\\n(%s)|{%s}\"] // %s", dot, cname, cname, collection.Type, strings.Join(columns, " | "), collection.Source)
	}

	for rule_num, rule := range s.Rules {
		rule_name := fmt.Sprintf("rule%d", rule_num)
		dot = fmt.Sprintf("%s\n\n\t%s [shape=diamond,label=\"rule %d\"] // %s", dot, rule_name, rule_num, rule.Source)

		for _, collection := range rule.Requires() {
			dot = fmt.Sprintf("%s\n\t%s -> %s", dot, collection, rule_name)
		}

		dot = fmt.Sprintf("%s\n\t%s -> %s", dot, rule_name, rule.Supplies)
	}

	return fmt.Sprintf("%s\n}", dot)
}
