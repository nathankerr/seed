package main

import (
	"fmt"
)

type seedTransform func(seedCollection) seedCollection

func applySeedTransforms(seeds seedCollection, transformations []seedTransform) seedCollection {
	for _, transform := range transformations {
		seeds = transform(seeds)
	}

	return seeds
}

func splitSeeds(seeds seedCollection) seedCollection {
	placement := make(map[string]string)
	s := newSeedCollection()

	for sname, seed := range seeds {
		for _, rule := range seed.rules {
			// find or create the seed, placing its name in name
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
				s[name] = newSeed()
			}

			// add the rule
			s[name].rules = append(s[name].rules, rule)

			// add the relevant collections
			for _, cname := range rule.supplies {
				placement[cname] = name
				s[name].collections[cname] = seed.collections[cname]
			}
			for _, cname := range rule.requires {
				placement[cname] = name
				s[name].collections[cname] = seed.collections[cname]
			}
		}
	}

	return s
}
