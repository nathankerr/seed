package main

import (
	"fmt"
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

	for _, rule := range(cluster.rules) {
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

	for _, rulenum := range(cluster.rules) {
		rule := seed.rules[rulenum]

		value, ok := rule.value.(*join)
		if !ok {
			panic("unsupported rule type for: " + rule.String())
		}

		// find the input channel and add its client to the beginning of value.output
		for collection, _ := range value.collections {
			if seed.collections[collection].typ == seedInput {
				output := []qualifiedColumn{qualifiedColumn{collection: collection, column: "client"}}
				for _, o := range value.output {
					output = append(output, o)
				}
				value.output = output
				break
			}
		}
		
		bud.rules = append(bud.rules, rule)
	}

	return buds
}
