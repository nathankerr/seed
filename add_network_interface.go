package main

import (
	"strings"
)

// adds a network interface by adding and handling explicit correlation data
func add_network_interface(sname string, seed *service, transformed map[string]*service) map[string]*service {
	info()

	sname = strings.Title(sname) + "Server"

	new_seed, ok := transformed[sname]
	if !ok {
		new_seed = &service{collections: make(map[string]*collection)}
	}

	// name the output address columns
	output_addrs := []string{}
	for name, collection := range seed.collections {
		if collection.ctype == collectionOutput {
			output_addrs = append(output_addrs, name+"_addr")
		}
	}

	// Add correlation information to the collections
	for name, collection := range seed.collections {
		switch collection.ctype {
		case collectionInput:
			// add output_addrs to the beginning of key
			key := []string{"@address"}
			for _, output_addr := range output_addrs {
				key = append(key, output_addr)
			}
			for _, ckey := range collection.key {
				key = append(key, ckey)
			}
			collection.key = key
		case collectionOutput:
			key := []string{"@" + name + "_addr"}
			for _, ckey := range collection.key {
				key = append(key, ckey)
			}
			collection.key = key
		case collectionTable:
			// no-op
		default:
			// should not get here
			panic(collection.ctype)
		}
		new_seed.collections[name] = collection

		transformed[sname] = new_seed
	}

	// rewrite the rules to take the correlation data into account
	for _, rule := range seed.rules {
		inputs := []string{}
		for _, name := range rule.collections() {
			if seed.collections[name].ctype == collectionInput {
				inputs = append(inputs, name)
			}
		}

		// The correlation data needs to be matched in the predicates
		for i := 1; i < len(inputs); i++ {
			for _, output_addr := range output_addrs {
				rule.predicate = append(rule.predicate, constraint{
					left: qualifiedColumn{
						collection: inputs[0],
						column:     output_addr},
					right: qualifiedColumn{
						collection: inputs[i],
						column:     output_addr},
				})
			}
		}

		switch seed.collections[rule.supplies].ctype {
		case collectionOutput:
			// convert to async insert as required by channels
			rule.operation = "<~"
			// add correlation data to projection
			projection := []qualifiedColumn{}
			if len(inputs) > 0 {
				projection = append(projection, qualifiedColumn{
					collection: inputs[0],
					column:     rule.supplies + "_addr"})
			}
			for _, o := range rule.projection {
				projection = append(projection, o)
			}
			rule.projection = projection
		case collectionTable:
			// no-op
		default:
			// should not get here
			panic(seed.collections[rule.supplies].ctype)
		}

		new_seed.rules = append(new_seed.rules, rule)
	}

	// convert the inputs and outputs into channels
	for _, collection := range new_seed.collections {
		switch collection.ctype {
		case collectionInput, collectionOutput:
			collection.ctype = collectionChannel
		case collectionChannel, collectionTable:
			// no-op
		default:
			// shouldn't get here
			panic(collection.ctype)
		}
	}

	return transformed
}
