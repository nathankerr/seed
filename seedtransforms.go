package main

type seedTransform func(seedCollection) seedCollection

func applySeedTransforms(seeds seedCollection, transformations []seedTransform) seedCollection {
	for _, transform := range transformations {
		seeds = transform(seeds)
	}

	return seeds
}
