package main

import (
	"fmt"
)

func getGroups(sname string, seed *service) map[string]*group {
	groups := make(map[string]*group)
	collectionToGroupMap := make(map[string]string)

	for num, rule := range seed.rules {
		// find or create the group, ref with groupName
		groupName := ""
		for _, collection := range rule.collections() {
			// tables are not included in the group
			if seed.collections[collection].ctype == collectionTable {
				continue
			}

			name, ok := collectionToGroupMap[collection]
			if ok {
				groupName = name
				break
			}
		}
		if groupName == "" {
			groupName = fmt.Sprintf("%s%d", sname, rule.source.line)
			collections := make(map[string]collectionType)
			groups[groupName] = &group{collections: collections}
		}

		// add the rule
		groups[groupName].rules = append(groups[groupName].rules, num)

		// add the relevant collections
		collectionToGroupMap[rule.supplies] = groupName
		groups[groupName].collections[rule.supplies] =
			seed.collections[rule.supplies].ctype

		for _, cname := range rule.collections() {
			collectionToGroupMap[cname] = groupName
			groups[groupName].collections[cname] = seed.collections[cname].ctype
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
