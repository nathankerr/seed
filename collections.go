package main

func (r *rule) collections() []string {
	collectionsmap := make(map[string]bool) // map only used for uniqueness

	// supplies
	collectionsmap[r.supplies] = true

	// projection
	for _, qc := range r.projection {
		collectionsmap[qc.collection] = true
	}

	// predicate
	for _, c := range r.predicate {
		collectionsmap[c.left.collection] = true
		collectionsmap[c.right.collection] = true
	}

	// convert map to []string
	collections := []string{}
	for collection, _ := range collectionsmap {
		collections = append(collections, collection)
	}

	return collections
}
