package service

func (r *Rule) Collections() []string {
	collectionsmap := make(map[string]bool) // map only used for uniqueness

	// supplies
	collectionsmap[r.Supplies] = true

	for _, requires := range r.Requires() {
		collectionsmap[requires] = true
	}

	// convert map to []string
	collections := []string{}
	for collection, _ := range collectionsmap {
		collections = append(collections, collection)
	}

	return collections
}

func (r *Rule) Requires() []string {
	requiresmap := make(map[string]bool) // map only used for uniqueness

	// projection
	for _, qc := range r.Projection {
		requiresmap[qc.Collection] = true
	}

	// predicate
	for _, c := range r.Predicate {
		requiresmap[c.Left.Collection] = true
		requiresmap[c.Right.Collection] = true
	}

	// convert map to []string
	requires := []string{}
	for collection, _ := range requiresmap {
		requires = append(requires, collection)
	}

	return requires
}
