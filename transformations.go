package main

import (
	"fmt"
)

// toggle on and off by commenting the first return statement
func transformationinfo(args ...interface{}) {
	// return
	info(args...)
}

type seedToBudTransformation func(buds budCollection, cluster *cluster,
	seed *seed, sname string) budCollection
type seedToBudTransformations map[string]seedToBudTransformation

func applySeedToBudTransformations(seeds seedCollection,
	transformationList ...seedToBudTransformations) budCollection {
	transformationinfo()

	buds := make(budCollection)

	for _, transformations := range transformationList {
		for sname, seed := range seeds {
			clusters := getClusters(sname, seed)

			for name, cluster := range clusters {
				transformation, ok := transformations[cluster.typ()]
				if !ok {
					fmt.Println("Seed to Bud Transformation for", name, cluster.typ(),
						"not supported!\n", cluster)
					continue
				}

				buds = transformation(buds, cluster, seed, sname)
			}
		}
	}

	return buds
}

type seedTransformation func(seeds seedCollection, cluster *cluster,
	seed *seed, sname string) (sc seedCollection, delete_seed bool)
type seedTransformations map[string]seedTransformation

func applySeedTransformations(seeds seedCollection,
	transformationList ...seedTransformations) seedCollection {
	transformationinfo()

	for _, transformations := range transformationList {
		// iterating over the changing set of seeds also iterated
		// (inconsistently) over the seeds which were added
		seedsCopy := make(seedCollection, len(seeds))
		for sname, seed := range seeds {
			seedsCopy[sname] = seed
		}

		for sname, seed := range seedsCopy {
			clusters := getClusters(sname, seed)
			delete_seed := false

			for name, cluster := range clusters {
				transformation, ok := transformations[cluster.typ()]
				if !ok {
					fmt.Println("Transformation for", name, cluster.typ(),
						"not supported!")
					continue
				}

				var del bool
				seeds, del = transformation(seeds, cluster, seed, sname)
				if del {
					delete_seed = true
				}
			}

			if delete_seed {
				delete(seeds, sname)
			}
		}
	}

	return seeds
}

type budTransformation func(buds budCollection) budCollection

func applyBudTransforms(buds budCollection,
	transformationList ...budTransformation) budCollection {
	transformationinfo()

	for _, transformation := range transformationList {
		buds = transformation(buds)
	}

	return buds
}
