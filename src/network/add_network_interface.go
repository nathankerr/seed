package main

import (
	"fmt"
	"strings"
)

func add_network_interface(sname string, seed *service, services map[string]*service) map[string]*service {
	groups := getGroups(sname, seed)

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
			services = add_network_interface_helper(services, group, seed, sname)
		default:
			// shouldn't get here
			panic(group.typ())
		}
	}

	return services
}

type group struct {
	rules       []int
	collections map[string]collectionType
}

func getGroups(sname string, seed *service) map[string]*group {
	groups := make(map[string]*group)
	collectionToGroupMap := make(map[string]string)

	for num, rule := range seed.Rules {
		// find or create the group, ref with groupName
		groupName := ""
		for _, collection := range rule.collections() {
			// tables are not included in the group
			if seed.Collections[collection].Type == collectionTable {
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
			collections := make(map[string]collectionType)
			groups[groupName] = &group{collections: collections}
		}

		// add the rule
		groups[groupName].rules = append(groups[groupName].rules, num)

		// add the relevant collections
		collectionToGroupMap[rule.Supplies] = groupName
		groups[groupName].collections[rule.Supplies] =
			seed.Collections[rule.Supplies].Type

		for _, cname := range rule.collections() {
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
		case collectionInput:
			inputs++
		case collectionOutput:
			outputs++
		case collectionTable:
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
func add_network_interface_helper(buds map[string]*service, group *group,
	seed *service, sname string) map[string]*service {
	info()

	sname = strings.Title(sname) + "Server"

	bud, ok := buds[sname]
	if !ok {
		bud = &service{Collections: make(map[string]*collection), Source: seed.Source}
	}

	// name the output address columns
	output_addrs := []string{}
	for name, _ := range group.collections {
		collection := seed.Collections[name]
		if collection.Type == collectionOutput {
			output_addrs = append(output_addrs, name+"_addr")
		}
	}

	// Add correlation information to the collections
	for name, _ := range group.collections {
		collection := seed.Collections[name]
		switch collection.Type {
		case collectionInput:
			// add output_addrs to the beginning of key
			key := []string{"@address"}
			for _, output_addr := range output_addrs {
				key = append(key, output_addr)
			}
			for _, ckey := range collection.Key {
				key = append(key, ckey)
			}
			collection.Key = key
		case collectionOutput:
			key := []string{"@" + name + "_addr"}
			for _, ckey := range collection.Key {
				key = append(key, ckey)
			}
			collection.Key = key
		case collectionTable:
			// no-op
		default:
			// should not get here
			panic(collection.Type)
		}
		bud.Collections[name] = collection

		buds[sname] = bud
	}

	// rewrite the rules to take the correlation data into account
	for _, rulenum := range group.rules {
		rule := seed.Rules[rulenum]

		inputs := []string{}
		for _, name := range rule.collections() {
			if seed.Collections[name].Type == collectionInput {
				inputs = append(inputs, name)
			}
		}

		// The correlation data needs to be matched in the predicates
		for i := 1; i < len(inputs); i++ {
			for _, output_addr := range output_addrs {
				rule.Predicate = append(rule.Predicate, constraint{
					Left: qualifiedColumn{
						Collection: inputs[0],
						Column:     output_addr},
					Right: qualifiedColumn{
						Collection: inputs[i],
						Column:     output_addr},
				})
			}
		}

		switch seed.Collections[rule.Supplies].Type {
		case collectionOutput:
			// convert to async insert as required by channels
			rule.Operation = "<~"
			// add correlation data to projection
			projection := []qualifiedColumn{}
			if len(inputs) > 0 {
				projection = append(projection, qualifiedColumn{
					Collection: inputs[0],
					Column:     rule.Supplies + "_addr"})
			}
			for _, o := range rule.Projection {
				projection = append(projection, o)
			}
			rule.Projection = projection
		case collectionTable:
			// no-op
		default:
			// should not get here
			panic(seed.Collections[rule.Supplies].Type)
		}

		bud.Rules = append(bud.Rules, rule)
	}

	for cname, _ := range group.collections {
		collection := seed.Collections[cname]
		switch collection.Type {
		case collectionInput, collectionOutput:
			collection.Type = collectionChannel
		case collectionTable:
			// no-op
		default:
			// shouldn't get here
			panic(collection.Type)
		}
	}

	return buds
}
