package main

import (
	"strings"
)

// not in seed: 000, 010, 0n0, 100, n00
// unknowable send_to_addr: 011, 01n, 0n1, 0nn
func add_clients(buds map[string]*service, group *group, seed *service, sname string) map[string]*service {
	info()

	sname = strings.Title(sname) + "Server"

	bud, ok := buds[sname]
	if !ok {
		bud = &service{collections: make(map[string]*collection)}
	}

	// create a list of output_addrs
	output_addrs := []string{}
	for name, _ := range group.collections {
		collection := seed.collections[name]
		if collection.ctype == collectionOutput {
			output_addrs = append(output_addrs, name+"_addr")
		}
	}

	// process the collections
	for name, _ := range group.collections {
		collection := seed.collections[name]
		switch collection.ctype {
		case collectionInput:
			// add output_addrs to the beginning of key
			key := []string{}
			for _, output_addr := range output_addrs {
				key = append(key, output_addr)
			}
			for _, ckey := range collection.key {
				key = append(key, ckey)
			}
			collection.key = key

			collection = add_address(collection)
		case collectionOutput:
			collection = add_address(collection)
			collection.key[0] = "@" + name + "_addr"
		case collectionTable:
			// no-op
		default:
			panic("should not get here")
		}
		bud.collections[name] = collection

		buds[sname] = bud
	}

	// process the rules
	for _, rulenum := range group.rules {
		rule := seed.rules[rulenum]

		inputs := []string{}
		for _, name := range rule.collections() {
			if seed.collections[name].ctype == collectionInput {
				inputs = append(inputs, name)
			}
		}

		// add predicates when needed
		for i := 1; i < len(inputs); i++ {
			for _, output_addr := range output_addrs {
				rule.predicate = append(rule.predicate, constraint{
					left:  qualifiedColumn{collection: inputs[0], column: output_addr},
					right: qualifiedColumn{collection: inputs[i], column: output_addr},
				})
			}
		}

		switch seed.collections[rule.supplies].ctype {
		case collectionOutput:
			rule.operation = "<~"
			projection := []qualifiedColumn{}
			if len(inputs) > 0 {
				projection = append(projection,
					qualifiedColumn{collection: inputs[0], column: rule.supplies + "_addr"})
			}
			for _, o := range rule.projection {
				projection = append(projection, o)
			}
			rule.projection = projection
		case collectionTable:
			// no-op
		default:
			panic("shouldn't get here")
		}

		bud.rules = append(bud.rules, rule)
	}

	return buds
}

func add_address(collection *collection) *collection {
	switch collection.ctype {
	case collectionInput, collectionOutput:
		key := []string{"@address"}
		for _, ckey := range collection.key {
			key = append(key, ckey)
		}
		collection.key = key
	case collectionTable:
		// no-op
	default:
		panic("shouldn't get here")
	}

	return collection
}
