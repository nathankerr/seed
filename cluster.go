package main

import(
	"fmt"
)

type cluster struct {
	rules []int
	collections map[string]seedCollectionType
}

func newCluster() *cluster {
	return &cluster{collections: make(map[string]seedCollectionType)}
}

func getClusters(sname string, seed *seed) map[string]*cluster {
	placement := make(map[string]string)
	clusters := make(map[string]*cluster)

	for num, rule := range seed.rules {
		// find or create the ruleSet
		name := ""
		for _, cname := range rule.supplies {
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
		for _, cname := range rule.requires {
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
			clusters[name] = newCluster()
		}

		// add the rule
		clusters[name].rules = append(clusters[name].rules, num)

		// add the relevant collections
		for _, cname := range rule.supplies {
			placement[cname] = name
			clusters[name].collections[cname] = seed.collections[cname].typ
		}
		for _, cname := range rule.requires {
			placement[cname] = name
			clusters[name].collections[cname] = seed.collections[cname].typ
		}
	}

	return clusters
}