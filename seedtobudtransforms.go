package main

import(
	"strings"
)

type seedToBudTransform func(seeds seedCollection, buds budCollection) budCollection

func applySeedToBudTransforms(seeds seedCollection, transformations []seedToBudTransform) budCollection {
	buds := make(budCollection)

	for _, transform := range transformations {
		buds = transform(seeds, buds)
	}

	return buds
}

func seedTableToBudTable(name string, typ budTableType, t *table) *budTable {
	b := newBudTable()

	b.name = name
	b.typ = typ

	if b.typ == budChannel {
		key := []string{"@address"}
		for _, tkey := range(t.key) {
			key = append(key, tkey)
		}

		b.key = key
	} else {
		b.key = t.key
	}

	b.columns = t.columns
	b.source = t.source

	return b
}

func generateServer(seeds seedCollection, buds budCollection) budCollection {
	for name, seed := range(seeds) {
		bud := newBud()
		name = strings.Title(name) + "Server"

		for sname, stable := range(seed.inputs) {
			btable := seedTableToBudTable(sname, budChannel, stable)
			bud.collections[sname] = btable
		}

		for sname, stable := range(seed.outputs) {
			btable := seedTableToBudTable(sname, budChannel, stable)
			bud.collections[sname] = btable
		}

		for sname, stable := range(seed.tables) {
			btable := seedTableToBudTable(sname, budPersistant, stable)
			bud.collections[sname] = btable
		}

		for _, rule := range(seed.rules) {
			rule.value += ".payloads"
			bud.rules = append(bud.rules, rule)
		}

		buds[name] = bud
	}

	return buds
}