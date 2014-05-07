// Package bloom exports a seed to the bloom host environment
package bloom

import (
	"fmt"
	"github.com/nathankerr/seed"
	"reflect"
	"strings"
)

// ToBloom converts a seed into a bloom program
func ToBloom(s *seed.Seed, name string) ([]byte, error) {
	str := fmt.Sprintf("module %s\n", strings.Title(name))

	// collections
	str = fmt.Sprintf("%s  state do\n", str)
	for cname, collection := range s.Collections {
		str = fmt.Sprintf("%s    %s\n",
			str,
			collectionToBloom(collection, cname),
		)
	}
	str = fmt.Sprintf("%s  end\n", str)

	// rules
	str = fmt.Sprintf("%s\n  bloom do\n", str)
	for ruleNum, rule := range s.Rules {
		str = fmt.Sprintf("%s    %s # rule %d\n",
			str,
			ruleToBloom(rule),
			ruleNum,
		)
	}
	str = fmt.Sprintf("%s  end\n", str)

	return []byte(fmt.Sprintf("%send\n", str)), nil
}

func ruleToBloom(r *seed.Rule) string {
	var selecter string
	collections := r.Requires()

	index := make(map[string]string)
	names := []string{}
	for i, c := range collections {
		name := fmt.Sprintf("c%d", i)
		index[c] = name
		names = append(names, name)
	}

	intension := []string{}
	for _, expression := range r.Intension {
		switch value := expression.(type) {
		case seed.QualifiedColumn:
			intension = append(intension, fmt.Sprintf("%s.%s",
				index[value.Collection], value.Column))
		case seed.MapFunction:
			arguments := []string{}
			for _, qc := range value.Arguments {
				arguments = append(arguments, fmt.Sprintf("%s.%s",
					index[qc.Collection], qc.Column))
			}

			intension = append(intension, fmt.Sprintf("%s(%s)",
				value.Name, strings.Join(arguments, ", ")))
		case seed.ReduceFunction:
			for _, qc := range value.Arguments {
				intension = append(intension, fmt.Sprintf("%s.%s",
					index[qc.Collection], qc.Column))
			}
		default:
			panic(fmt.Sprintf("unhandled type: %v",
				reflect.TypeOf(expression).String()))
		}
	}

	if len(collections) == 1 {
		selecter = fmt.Sprintf("%s do |%s|\n      [%s]\n    end",
			collections[0],
			strings.Join(names, ", "),
			strings.Join(intension, ", "))
	} else {
		predicates := []string{}
		for _, p := range r.Predicate {
			predicates = append(predicates, p.String())
		}

		selecter = fmt.Sprintf("(%s).combos(%s) do |%s|\n      [%s]\n    end",
			strings.Join(collections, " * "),
			strings.Join(predicates, ", "),
			strings.Join(names, ", "),
			strings.Join(intension, ", "))
	}

	return fmt.Sprintf("%s %s %s",
		r.Supplies,
		r.Operation,
		selecter)
}

func collectionToBloom(c *seed.Collection, name string) string {
	var declaration string

	switch c.Type {
	case seed.CollectionInput:
		declaration = "interface input,"
	case seed.CollectionOutput:
		declaration = "interface output,"
	case seed.CollectionChannel:
		declaration = "channel"
	case seed.CollectionTable:
		declaration = "table"
	case seed.CollectionScratch:
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
