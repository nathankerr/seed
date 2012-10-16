package main

import (
	"strings"
)

var generate_server = seedToBudTransformations{
	"101": generate_server_101,
	"111": generate_server_111,
}

func generate_server_101(buds budCollection, cluster *cluster, seed *seed, sname string) budCollection {
	transformationinfo()

	sname = strings.Title(sname) + "Server"

	bud, ok := buds[sname]
	if !ok {
		bud = newBud()
	}

	// as the inputs and outputs are already
	// projected, there is no need to remove the
	// address columns added to the interfaces
	// when converting them to channels
	for name, _ := range cluster.collections {
		collection := seed.collections[name]
		switch collection.typ {
		case seedInput:
			channel := seedTableToBudTable(name, budChannel, collection)
			bud.collections[name] = channel
		case seedOutput:
			output := seedTableToBudTable(name, budScratch, collection)
			bud.collections[name] = output
		case seedTable:
			table := seedTableToBudTable(name, budPersistant, collection)
			bud.collections[name] = table
		}

		buds[sname] = bud
	}

	for _, rule := range cluster.rules {
		bud.rules = append(bud.rules, seed.rules[rule])
	}

	return buds
}

func generate_server_111(buds budCollection, cluster *cluster, seed *seed, sname string) budCollection {
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
			key := []string{"client"}
			for _, ckey := range collection.key {
				key = append(key, ckey)
			}
			collection.key = key

			input := seedTableToBudTable(name, budChannel, collection)
			bud.collections[name] = input
		case seedOutput:
			output := seedTableToBudTable(name, budChannel, collection)
			output.key[0] = "@client"
			bud.collections[name] = output
		case seedTable:
			table := seedTableToBudTable(name, budPersistant, collection)
			bud.collections[name] = table
		default:
			panic("should not get here")
		}

		buds[sname] = bud
	}

	for _, rulenum := range cluster.rules {
		rule := seed.rules[rulenum]

		// find the input channel and add its client to the beginning of output
		for collection, _ := range rule.requires {
			if seed.collections[collection].typ == seedInput {
				output := []qualifiedColumn{qualifiedColumn{collection: collection, column: "client"}}
				for _, o := range rule.output {
					output = append(output, o)
				}
				rule.output = output
				break
			}
		}

		if rule.typ == ruleSet {
			rule.typ = ruleAsyncInsert
		}

		bud.rules = append(bud.rules, rule)
	}

	return buds
}
