package main

type seedTransform func(seedCollection) seedCollection

func applySeedTransforms(seeds seedCollection, transformations []seedTransform) seedCollection {
	for _, transform := range transformations {
		seeds = transform(seeds)
	}

	return seeds
}

func splitSeeds(seeds seedCollection) seedCollection {
	s := newSeedCollection()
	for sname, seed := range seeds {
		clusters := getClusters(sname, seed)

		for name, cluster := range clusters {
			s[name] = newSeed()

			for cname, _ := range cluster.collections {
				s[name].collections[cname] = seed.collections[cname]
			}

			for _, rnum := range cluster.rules {
				s[name].rules = append(s[name].rules, seed.rules[rnum])
			}
		}

	}
	return s
}
