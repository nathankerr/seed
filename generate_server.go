package main

import (
	"fmt"
	"strings"
)

var generate_server = seedToBudTransformations{
	"101": generate_server_101,
}

func generate_server_101(buds budCollection, cluster *cluster, seed *seed, sname string) budCollection {
	transformationinfo()

	sname = strings.Title(sname) + "Server"

	bud, ok := buds[sname]
	if !ok {
		bud = newBud()
	}

	for name, _ := range cluster.collections {
		collection := seed.collections[name]
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

		buds[sname] = bud
	}

	return buds
}
