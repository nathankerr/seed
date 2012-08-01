package main

type budTransform func(buds budCollection) budCollection

func applyBudTransforms(buds budCollection, transformations []budTransform) budCollection {
	for _, transform := range transformations {
		buds = transform(buds)
	}

	return buds
}
