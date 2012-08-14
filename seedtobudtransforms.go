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

// generateProtocol needs to have been ran first
func generateServer(seeds seedCollection, buds budCollection) budCollection {
	for name, seed := range seeds {
		bud := newBud()
		name = strings.Title(name) + "Server"

		// replace the inputs with channels and scratches
		// this removes the need to rewrite the rules
		for iname, table := range seed.inputs {
			input := seedTableToBudTable(iname, budScratch, table)
			bud.collections[iname] = input

			cname := iname + "_channel"
			channel := seedTableToBudTable(cname, budChannel, table)
			bud.collections[cname] = channel

			rewrite := newRule()
			rewrite.value = fmt.Sprintf("%s <= %s.payloads", iname, cname)
			rewrite.source = table.source
			bud.rules = append(bud.rules, rewrite)
		}

		// replace the outputs with channela and scratches
		for oname, table := range seed.outputs {
			output := seedTableToBudTable(oname, budScratch, table)
			bud.collections[oname] = output

			cname := oname + "_channel"
			channel := seedTableToBudTable(cname, budChannel, table)
			bud.collections[cname] = channel

			rewrite := newRule()
			rewrite.value = fmt.Sprintf("%s <~ %s.payloads", cname, oname)
			rewrite.source = table.source
			bud.rules = append(bud.rules, rewrite)
		}

		for tname, table := range seed.tables {
			btable := seedTableToBudTable(tname, budPersistant, table)
			bud.collections[tname] = btable
		}

		for _, r := range seed.rules {
			bud.rules = append(bud.rules, r)
		}

		buds[name] = bud
	}

	return buds
}

// clients only need the interfaces themselves
func generateClient(seeds seedCollection, buds budCollection) budCollection {
	for name, seed := range seeds {
		bud := newBud()
		name = strings.Title(name) + "Client"

		for sname, stable := range seed.inputs {
			input := seedTableToBudTable(sname, budInterface, stable)
			input.input = true
			bud.collections[sname] = input

			cname := sname + "_channel"
			channel := seedTableToBudTable(cname, budChannel, stable)
			bud.collections[name] = channel

			transfer := newRule()
			transfer.source = stable.source
			transfer.value = name + " <~ " + sname
			bud.rules = append(bud.rules, transfer)

		}

		for sname, stable := range seed.outputs {
			btable := seedTableToBudTable(sname, budInterface, stable)
			btable.input = false
			bud.collections[sname] = btable
		}

		buds[name] = bud
	}

	return buds
}
