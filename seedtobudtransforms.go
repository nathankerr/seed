package main

type seedToBudTransform func(seeds seedCollection, buds budCollection) budCollection

func applySeedToBudTransforms(seeds seedCollection, transformations []seedToBudTransform) budCollection {
	buds := make(budCollection)

	for _, transform := range transformations {
		buds = transform(seeds, buds)
	}

	return buds
}
