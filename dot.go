package seed

import (
	"fmt"
	"strings"
)

func ToDot(seed *Seed, name string) ([]byte, error) {
	info()

	dot := fmt.Sprintf("digraph %s {", name)
	dot = fmt.Sprintf("%s\n\tmargin=\"0\"", dot)
	dot = fmt.Sprintf("%s\n", dot)

	for cname, collection := range seed.Collections {
		columns := collection.Key
		for _, column := range collection.Data {
			columns = append(columns, column)
		}

		dot = fmt.Sprintf("%s\n\t%s [shape=record,label=\"%s\\n(%s)|{%s}\"]", dot, cname, cname, collection.Type, strings.Join(columns, " | "))
	}

	for rule_num, rule := range seed.Rules {
		rule_name := fmt.Sprintf("rule%d", rule_num)
		dot = fmt.Sprintf("%s\n\n\t%s [shape=diamond,label=\"rule %d\"]", dot, rule_name, rule_num)

		for _, collection := range rule.Requires() {
			dot = fmt.Sprintf("%s\n\t%s -> %s", dot, collection, rule_name)
		}

		dot = fmt.Sprintf("%s\n\t%s -> %s", dot, rule_name, rule.Supplies)
	}

	return []byte(fmt.Sprintf("%s\n}", dot)), nil
}
