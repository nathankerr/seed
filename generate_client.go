package main

import (
	"strings"
)

var generate_client = seedToBudTransformations{
	"101": generate_client_101,
}

func generate_client_101(buds budCollection, cluster *cluster, seed *seed, sname string) budCollection {
	transformationinfo()

	sname = strings.Title(sname) + "Client"

	bud, ok := buds[sname]
	if !ok {
		bud = newBud()
	}

	for name, _ := range cluster.collections {
		collection := seed.collections[name]
		switch collection.typ {
		case seedInput:
			input := seedTableToBudTable(name, budInterface, collection)
			input.input = true
			bud.collections[name] = input

			cname := name + "_channel"
			channel := seedTableToBudTable(cname, budChannel, collection)
			bud.collections[cname] = channel

			// TODO: these rules don't add the server address
			transfer := newRule(collection.source)
			transfer.value = cname + " <~ " + name
			bud.rules = append(bud.rules, transfer)
		case seedOutput:
			table := seedTableToBudTable(name, budInterface, collection)
			table.input = false
			bud.collections[name] = table
		}

		buds[sname] = bud
	}

	return buds
}
