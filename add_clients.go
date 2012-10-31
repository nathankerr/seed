package main

import (
	"strings"
)

// not in seed: 000, 010, 0n0, 100, n00
// unknowable send_to_addr: 011, 01n, 0n1, 0nn
func add_clients(buds budCollection, cluster *cluster, seed *seed, sname string) budCollection {
	transformationinfo()

	sname = strings.Title(sname) + "Server"

	bud, ok := buds[sname]
	if !ok {
		bud = newBud()
	}

	// create a list of output_addrs
	output_addrs := []string{}
	for name, _ := range cluster.collections {
		collection := seed.collections[name]
		if collection.typ == seedOutput {
			output_addrs = append(output_addrs, name+"_addr")
		}
	}

	// process the collections
	for name, _ := range cluster.collections {
		collection := seed.collections[name]
		switch collection.typ {
		case seedInput:
			// add output_addrs to the beginning of key
			key := []string{}
			for _, output_addr := range output_addrs {
				key = append(key, output_addr)
			}
			for _, ckey := range collection.key {
				key = append(key, ckey)
			}
			collection.key = key

			input := seedTableToBudTable(name, budChannel, collection)
			bud.collections[name] = input
		case seedOutput:
			output := seedTableToBudTable(name, budChannel, collection)
			output.key[0] = "@" + name + "_addr"
			bud.collections[name] = output
		case seedTable:
			table := seedTableToBudTable(name, budPersistant, collection)
			bud.collections[name] = table
		default:
			panic("should not get here")
		}

		buds[sname] = bud
	}

	// process the rules
	for _, rulenum := range cluster.rules {
		rule := seed.rules[rulenum]

		inputs := []string{}
		for name, _ := range rule.requires {
			if seed.collections[name].typ == seedInput {
				inputs = append(inputs, name)
			}
		}

		// add predicates when needed
		for i := 1; i < len(inputs); i++ {
			for _, output_addr := range output_addrs {
				rule.predicates = append(rule.predicates, predicate{
					left:  qualifiedColumn{collection: inputs[0], column: output_addr},
					right: qualifiedColumn{collection: inputs[i], column: output_addr},
				})
			}
		}

		switch seed.collections[rule.supplies].typ {
		case seedOutput:
			rule.typ = ruleAsyncInsert
			output := []qualifiedColumn{}
			if len(inputs) > 0 {
				output = append(output, qualifiedColumn{collection: inputs[0], column: rule.supplies + "_addr"})
			}
			for _, o := range rule.output {
				output = append(output, o)
			}
			rule.output = output
		case seedTable:
			// no-op
		default:
			panic("shouldn't get here")
		}

		bud.rules = append(bud.rules, rule)
	}

	return buds
}
