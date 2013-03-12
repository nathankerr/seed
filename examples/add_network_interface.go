package examples

import (
	"errors"
	"fmt"
	service "github.com/nathankerr/seed"
	"strings"
)

func Add_network_interface(orig *service.Seed) (*service.Seed, error) {
	err := orig.InSubset()
	if err != nil {
		return nil, errors.New("Adding a network interface requires that the specified orig be in the subset. " + err.Error())
	}

	groups := getGroups(orig.Name, orig)

	networked := &service.Seed{
		Collections: make(map[string]*service.Collection),
		Source: orig.Source,
		Name: strings.Title(orig.Name) + "Server",
	}

	for _, group := range groups {
		switch group.typ() {
		case "000", "010", "0n0", "100", "n00": // not possible
			panic(group.typ())
		case "011", "01n", "0n1", "0nn": // not input driven (not handled)
			panic(group.typ())
		case "001", "00n", "101", "10n", "n01", "n0n", // passthrough
			"110", "111", "11n", // single output
			"1n0", "1n1", "1nn", // multiple output
			"n10", "n11", "n1n", "nn0", "nn1", "nnn": // multiple input
			networked = add_network_interface_helper(orig, group, networked)
		default:
			// shouldn't get here
			panic(group.typ())
		}
	}

	return networked, nil
}

type group struct {
	rules       []int
	collections map[string]service.CollectionType
}

func getGroups(sname string, seed *service.Seed) map[string]*group {
	groups := make(map[string]*group)
	collectionToGroupMap := make(map[string]string)

	for num, rule := range seed.Rules {
		// find or create the group, ref with groupName
		groupName := ""
		for _, collection := range rule.Collections() {
			// tables are not included in the group
			if seed.Collections[collection].Type == service.CollectionTable {
				continue
			}

			name, ok := collectionToGroupMap[collection]
			if ok {
				groupName = name
				break
			}
		}
		if groupName == "" {
			groupName = fmt.Sprintf("%s%d", sname, rule.Source.Line)
			collections := make(map[string]service.CollectionType)
			groups[groupName] = &group{collections: collections}
		}

		// add the rule
		groups[groupName].rules = append(groups[groupName].rules, num)

		// add the relevant collections
		collectionToGroupMap[rule.Supplies] = groupName
		groups[groupName].collections[rule.Supplies] =
			seed.Collections[rule.Supplies].Type

		for _, cname := range rule.Collections() {
			collectionToGroupMap[cname] = groupName
			groups[groupName].collections[cname] = seed.Collections[cname].Type
		}
	}

	return groups
}

func (g *group) typ() string {
	var inputs, outputs, tables int

	for _, ctyp := range g.collections {
		switch ctyp {
		case service.CollectionInput:
			inputs++
		case service.CollectionOutput:
			outputs++
		case service.CollectionTable:
			tables++
		default:
			// shouldn't get here
			panic(ctyp)
		}
	}

	return fmt.Sprint(count(inputs), count(outputs), count(tables))
}

func count(i int) string {
	var str string
	switch {
	case i == 0:
		str = "0"
	case i == 1:
		str = "1"
	case i > 1:
		str = "n"
	default:
		// shouldn't get here
		panic(i)
	}
	return str
}

// adds a network interface by adding and handling explicit correlation data
func add_network_interface_helper(orig *service.Seed, group *group, networked *service.Seed) *service.Seed {

	// name the output address columns
	output_addrs := []string{}
	for name, _ := range group.collections {
		collection := orig.Collections[name]
		if collection.Type == service.CollectionOutput {
			output_addrs = append(output_addrs, name+"_addr")
		}
	}

	// Add correlation information to the collections
	for name, _ := range group.collections {
		collection := orig.Collections[name]
		switch collection.Type {
		case service.CollectionInput:
			// add output_addrs to the beginning of key
			key := []string{"@address"}
			for _, output_addr := range output_addrs {
				key = append(key, output_addr)
			}
			for _, ckey := range collection.Key {
				key = append(key, ckey)
			}
			collection.Key = key
		case service.CollectionOutput:
			key := []string{"@" + name + "_addr"}
			for _, ckey := range collection.Key {
				key = append(key, ckey)
			}
			collection.Key = key
		case service.CollectionTable:
			// no-op
		default:
			// should not get here
			panic(collection.Type)
		}
		networked.Collections[name] = collection
	}

	// rewrite the rules to take the correlation data into account
	for _, rulenum := range group.rules {
		rule := orig.Rules[rulenum]

		inputs := []string{}
		for _, name := range rule.Collections() {
			if orig.Collections[name].Type == service.CollectionInput {
				inputs = append(inputs, name)
			}
		}

		// The correlation data needs to be matched in the predicates
		for i := 1; i < len(inputs); i++ {
			for _, output_addr := range output_addrs {
				rule.Predicate = append(rule.Predicate, service.Constraint{
					Left: service.QualifiedColumn{
						Collection: inputs[0],
						Column:     output_addr},
					Right: service.QualifiedColumn{
						Collection: inputs[i],
						Column:     output_addr},
				})
			}
		}

		switch orig.Collections[rule.Supplies].Type {
		case service.CollectionOutput:
			// convert to async insert as required by channels
			rule.Operation = "<~"
			// add correlation data to projection
			projection := []service.Expression{}
			if len(inputs) > 0 {
				projection = append(projection, service.Expression{Value: service.QualifiedColumn{
					Collection: inputs[0],
					Column:     rule.Supplies + "_addr"}})
			}
			for _, o := range rule.Projection {
				projection = append(projection, o)
			}
			rule.Projection = projection
		case service.CollectionTable:
			// no-op
		default:
			// should not get here
			panic(orig.Collections[rule.Supplies].Type)
		}

		networked.Rules = append(networked.Rules, rule)
	}

	for cname, _ := range group.collections {
		collection := orig.Collections[cname]
		switch collection.Type {
		case service.CollectionInput, service.CollectionOutput:
			collection.Type = service.CollectionChannel
		case service.CollectionTable:
			// no-op
		default:
			// shouldn't get here
			panic(collection.Type)
		}
	}

	return networked
}
