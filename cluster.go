package main

import (
	"fmt"
)

type cluster struct {
	rules       []int
	collections map[string]seedCollectionType
}

func (c *cluster) String() string {
	str := "collections:"
	for collection, typ := range c.collections {
		str = fmt.Sprintf("%s\n\t%s %s", str, typ, collection)
	}

	str = fmt.Sprintf("%s\n rules: %d", str, c.rules)
	return str
}

func getClusters(sname string, seed *seed) map[string]*cluster {
	placement := make(map[string]string)
	clusters := make(map[string]*cluster)

	for num, rule := range seed.rules {
		// find or create the ruleSet
		name := ""

		if seed.collections[rule.supplies].typ != seedTable {
			sname, ok := placement[rule.supplies]
			if ok {
				name = sname
				break
			}
		}

		for cname, _ := range rule.requires {
			// tables are not a basis for splitting
			if seed.collections[cname].typ == seedTable {
				continue
			}

			sname, ok := placement[cname]
			if ok {
				name = sname
				break
			}
		}
		if name == "" {
			name = fmt.Sprintf("%s%d", sname, rule.source.line)
			collections := make(map[string]seedCollectionType)
			clusters[name] = &cluster{collections: collections}
		}

		// add the rule
		clusters[name].rules = append(clusters[name].rules, num)

		// add the relevant collections
		placement[rule.supplies] = name
		clusters[name].collections[rule.supplies] =
			seed.collections[rule.supplies].typ

		for cname, _ := range rule.requires {
			placement[cname] = name
			clusters[name].collections[cname] = seed.collections[cname].typ
		}
	}

	return clusters
}

func (c *cluster) typ() string {
	var inputs, outputs, tables int

	for _, ctyp := range c.collections {
		switch ctyp {
		case seedInput:
			inputs++
		case seedOutput:
			outputs++
		case seedTable:
			tables++
		case seedScratch:
			// no-op
		}
	}

	return fmt.Sprint(count(inputs), count(outputs), count(tables))
}

func count(i int) string {
	switch {
	case i == 0:
		return "0"
	case i == 1:
		return "1"
	case i > 1:
		return "n"
	}
	return "?"
}
