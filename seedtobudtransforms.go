package main

import (
	"fmt"
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
		for _, tkey := range t.key {
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
	for sname, seed := range seeds {
		bud := newBud()
		sname = strings.Title(sname) + "Server"

		for name, collection := range seed.collections {
			switch collection.typ {
			case seedInput:
				// replace the inputs with channels and scratches
				// this removes the need to rewrite the rules
				input := seedTableToBudTable(name, budScratch, collection)
				bud.collections[name] = input

				cname := name + "_channel"
				channel := seedTableToBudTable(cname, budChannel, collection)
				bud.collections[cname] = channel

				rewrite := newRule(collection.source)
				rewrite.value = fmt.Sprintf("%s <= %s.payloads", name, cname)
				bud.rules = append(bud.rules, rewrite)
			case seedOutput:
				// replace the outputs with channels and scratches
				output := seedTableToBudTable(name, budScratch, collection)
				bud.collections[name] = output

				cname := name + "_channel"
				channel := seedTableToBudTable(cname, budChannel, collection)
				bud.collections[name] = channel

				rewrite := newRule(collection.source)
				rewrite.value = fmt.Sprintf("%s <~ %s.payloads", cname, name)
				bud.rules = append(bud.rules, rewrite)
			case seedTable:
				table := seedTableToBudTable(name, budPersistant, collection)
				bud.collections[name] = table
			}
		}

		for _, r := range seed.rules {
			bud.rules = append(bud.rules, r)
		}

		buds[sname] = bud
	}

	return buds
}

// clients only need the interfaces themselves
func generateClient(seeds seedCollection, buds budCollection) budCollection {
	for seed_name, seed := range seeds {
		bud := newBud()
		seed_name = strings.Title(seed_name) + "Client"

		for name, collection := range seed.collections {
			switch collection.typ {
			case seedInput:
				input := seedTableToBudTable(name, budInterface, collection)
				input.input = true
				bud.collections[name] = input

				cname := name + "_channel"
				channel := seedTableToBudTable(cname, budChannel, collection)
				bud.collections[cname] = channel

				transfer := newRule(collection.source)
				transfer.value = cname + " <~ " + name
				bud.rules = append(bud.rules, transfer)
			case seedOutput:
				table := seedTableToBudTable(name, budInterface, collection)
				table.input = false
				bud.collections[name] = table
			}
		}

		buds[seed_name] = bud
	}

	return buds
}
