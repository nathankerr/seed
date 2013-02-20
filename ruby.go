package seed

import (
	"fmt"
	"reflect"
	"strings"
)

func ToBloom(seed *Seed, name string) ([]byte, error) {
	return []byte(seed.toRuby(name)), nil
}

func (s *Seed) toRuby(name string) string {
	info()

	name = strings.Title(name)
	str := fmt.Sprintf("module %s\n", name)
	collections := s.Collections

	rules := fmt.Sprintf("\n  bloom do\n")
	for rule_num, rule := range s.Rules {
		rule_str, additional_collections := rule.Ruby(s)
		rules = fmt.Sprintf("%s    %s #%s rule %d\n", rules, rule_str, rule.Source, rule_num)
		for cname, collection := range additional_collections {
			collections[cname] = collection
		}
	}
	rules = fmt.Sprintf("%s  end\n", rules)

	str = fmt.Sprintf("%s  state do\n", str)
	for cname, collection := range collections {
		str = fmt.Sprintf("%s    %s #%s\n", str, collection.Ruby(cname),
			collection.Source)
	}
	str = fmt.Sprintf("%s  end\n", str)

	str += rules

	str = fmt.Sprintf("%send\n", str)

	return str
}

func (r *Rule) Ruby(service *Seed) (string, map[string]*Collection) {
	var selecter string

	additional_collections := make(map[string]*Collection)
	collections := r.Requires()

	index := make(map[string]string)
	names := []string{}
	for i, c := range collections {
		name := fmt.Sprintf("c%d", i)
		index[c] = name
		names = append(names, name)
	}

	output := []string{}
	for _, expression := range r.Projection {
		switch value := expression.Value.(type) {
		case QualifiedColumn:
			output = append(output,
				fmt.Sprintf("%s.%s", index[value.Collection], value.Column))
		case MapFunction:
			arguments := []string{}
			for _, qc := range value.Arguments {
				arguments = append(arguments,
					fmt.Sprintf("%s.%s", index[qc.Collection], qc.Column))
			}

			output = append(output,
				fmt.Sprintf("%s(%s)", value.Name, strings.Join(arguments, ", ")))
		case ReduceFunction:
			for _, qc := range value.Arguments {
				output = append(output,
					fmt.Sprintf("%s.%s", index[qc.Collection], qc.Column))
			}
		default:
			panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression.Value).String()))
		}
	}

	if len(collections) == 1 {
		selecter = fmt.Sprintf("%s do |%s|\n      [%s]\n    end",
			collections[0],
			strings.Join(names, ", "),
			strings.Join(output, ", "))
	} else {
		predicates := []string{}
		for _, p := range r.Predicate {
			predicates = append(predicates, p.String())
		}

		selecter = fmt.Sprintf("(%s).combos(%s) do |%s|\n      [%s]\n    end",
			strings.Join(collections, " * "),
			strings.Join(predicates, ", "),
			strings.Join(names, ", "),
			strings.Join(output, ", "))
	}

	return fmt.Sprintf("%s %s %s",
			r.Supplies,
			r.Operation,
			selecter),
		additional_collections
}

func (c *Collection) Ruby(name string) string {
	var declaration string

	switch c.Type {
	case CollectionInput:
		declaration = "interface input,"
	case CollectionOutput:
		declaration = "interface output,"
	case CollectionChannel:
		declaration = "channel"
	case CollectionTable:
		declaration = "table"
	case CollectionScratch:
		declaration = "scratch"
	default:
		// shouldn't get here
		panic(c.Type)
	}

	declaration = fmt.Sprintf("%s :%s, [", declaration, name)

	for _, v := range c.Key {
		declaration += fmt.Sprintf(":%s, ", v)
	}
	if len(c.Key) != 0 {
		declaration = declaration[:len(declaration)-2]
	}

	if len(c.Data) > 0 {
		declaration += "] => ["

		for _, v := range c.Data {
			declaration += fmt.Sprintf(":%s, ", v)
		}

		declaration = declaration[:len(declaration)-2] + "]"
	} else {
		declaration += "]"
	}

	return declaration
}
