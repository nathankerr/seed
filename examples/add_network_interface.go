package examples

import (
	"errors"
	"fmt"
	"github.com/nathankerr/seed"
	"strings"
)

func Add_network_interface(orig *seed.Seed) (*seed.Seed, error) {
	err := orig.InSubset()
	if err != nil {
		return nil, errors.New("Adding a network interface requires that the specified seed be in the subset. " + err.Error())
	}

	groups := getGroups(orig.Name, orig)

	networked := &seed.Seed{
		Collections: make(map[string]*seed.Collection),
		Name:        strings.Title(orig.Name) + "Server",
	}

	for _, group := range groups {
		switch group.typ() {
		case "000", "010", "0n0", "100", "n00": // not possible
			return nil, errors.New(group.typ() + " should not have been possible.")
		case "011", "01n", "0n1", "0nn": // not input driven (not handled)
			// while these types aren't handled, we will just pass them through
			networked = merge(orig, group, networked)
		case "001", "00n", "101", "10n", "n01", "n0n", // passthrough
			"110", "111", "11n", // single output
			"1n0", "1n1", "1nn", // multiple output
			"n10", "n11", "n1n", "nn0", "nn1", "nnn": // multiple input
			networked = add_interface(orig, group, networked)
		default:
			// shouldn't get here
			panic(group.typ())
		}
	}

	return networked, nil
}

type group struct {
	rules       []int
	collections map[string]seed.CollectionType
}

func getGroups(sname string, service *seed.Seed) map[string]*group {
	groups := make(map[string]*group)
	collectionToGroupMap := make(map[string]string)

	for num, rule := range service.Rules {
		// find or create the group, ref with groupName
		groupName := ""
		for _, collection := range rule.Collections() {
			// tables are not included in the group
			if service.Collections[collection].Type == seed.CollectionTable {
				continue
			}

			name, ok := collectionToGroupMap[collection]
			if ok {
				groupName = name
				break
			}
		}
		if groupName == "" {
			groupName = fmt.Sprintf("%s%d", sname, num)
			collections := make(map[string]seed.CollectionType)
			groups[groupName] = &group{collections: collections}
		}

		// add the rule
		groups[groupName].rules = append(groups[groupName].rules, num)

		// add the relevant collections
		collectionToGroupMap[rule.Supplies] = groupName
		groups[groupName].collections[rule.Supplies] =
			service.Collections[rule.Supplies].Type

		for _, cname := range rule.Collections() {
			collectionToGroupMap[cname] = groupName
			groups[groupName].collections[cname] = service.Collections[cname].Type
		}
	}

	return groups
}

func (g *group) typ() string {
	var inputs, outputs, tables int

	for _, ctyp := range g.collections {
		switch ctyp {
		case seed.CollectionInput:
			inputs++
		case seed.CollectionOutput:
			outputs++
		case seed.CollectionTable:
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
func add_interface(orig *seed.Seed, group *group, networked *seed.Seed) *seed.Seed {

	// name the output address columns
	output_addrs := []string{}
	for name, _ := range group.collections {
		collection := orig.Collections[name]
		if collection.Type == seed.CollectionOutput {
			output_addrs = append(output_addrs, name+"_addr")
		}
	}

	// Add correlation information to the collections
	for name, _ := range group.collections {
		collection := orig.Collections[name]
		switch collection.Type {
		case seed.CollectionInput:
			// add output_addrs to the beginning of key
			key := []string{"@address"}
			for _, output_addr := range output_addrs {
				key = append(key, output_addr)
			}
			for _, ckey := range collection.Key {
				key = append(key, ckey)
			}
			collection.Key = key
		case seed.CollectionOutput:
			key := []string{"@" + name + "_addr"}
			for _, ckey := range collection.Key {
				key = append(key, ckey)
			}
			collection.Key = key
		case seed.CollectionTable:
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
			if orig.Collections[name].Type == seed.CollectionInput {
				inputs = append(inputs, name)
			}
		}

		// The correlation data needs to be matched in the predicates
		for i := 1; i < len(inputs); i++ {
			for _, output_addr := range output_addrs {
				rule.Predicate = append(rule.Predicate, seed.Constraint{
					Left: seed.QualifiedColumn{
						Collection: inputs[0],
						Column:     output_addr},
					Right: seed.QualifiedColumn{
						Collection: inputs[i],
						Column:     output_addr},
				})
			}
		}

		switch orig.Collections[rule.Supplies].Type {
		case seed.CollectionOutput:
			// convert to async insert as required by channels
			rule.Operation = "<~"
			// add correlation data to projection
			projection := []seed.Expression{}
			if len(inputs) > 0 {
				projection = append(projection, seed.Expression{Value: seed.QualifiedColumn{
					Collection: inputs[0],
					Column:     rule.Supplies + "_addr"}})
			}
			for _, o := range rule.Projection {
				projection = append(projection, o)
			}
			rule.Projection = projection
		case seed.CollectionTable:
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
		case seed.CollectionInput, seed.CollectionOutput:
			collection.Type = seed.CollectionChannel
		case seed.CollectionTable:
			// no-op
		default:
			// shouldn't get here
			panic(collection.Type)
		}
	}

	return networked
}

func merge(orig *seed.Seed, group *group, networked *seed.Seed) *seed.Seed {
	for _, rulenum := range group.rules {
		networked.Rules = append(networked.Rules, orig.Rules[rulenum])
	}

	for collectionName, _ := range group.collections {
		networked.Collections[collectionName] = orig.Collections[collectionName]
	}

	return networked
}
