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
		new_seed = &service{Collections: make(map[string]*collection)}
	}

	// name the output address columns
	output_addrs := []string{}
	for name, collection := range seed.Collections {
		if collection.Type == collectionOutput {
			output_addrs = append(output_addrs, name+"_addr")
		}
	}

	// Add correlation information to the collections
	for name, collection := range seed.Collections {
		switch collection.Type {
		case collectionInput:
			// add output_addrs to the beginning of key
			key := []string{"@address"}
			for _, output_addr := range output_addrs {
				key = append(key, output_addr)
			}
			for _, ckey := range collection.Key {
				key = append(key, ckey)
			}
			collection.Key = key
		case collectionOutput:
			key := []string{"@" + name + "_addr"}
			for _, ckey := range collection.Key {
				key = append(key, ckey)
			}
			collection.Key = key
		case collectionTable:
			// no-op
		default:
			// should not get here
			panic(collection.Type)
		}
		new_seed.Collections[name] = collection

		transformed[sname] = new_seed
	}

	// rewrite the rules to take the correlation data into account
	for _, rule := range seed.Rules {
		inputs := []string{}
		for _, name := range rule.collections() {
			if seed.Collections[name].Type == collectionInput {
				inputs = append(inputs, name)
			}
		}

		// The correlation data needs to be matched in the predicates
		for i := 1; i < len(inputs); i++ {
			for _, output_addr := range output_addrs {
				rule.Predicate = append(rule.Predicate, constraint{
					Left: qualifiedColumn{
						Collection: inputs[0],
						Column:     output_addr},
					Right: qualifiedColumn{
						Collection: inputs[i],
						Column:     output_addr},
				})
			}
		}

		switch seed.Collections[rule.Supplies].Type {
		case collectionOutput:
			// convert to async insert as required by channels
			rule.Operation = "<~"
			// add correlation data to projection
			projection := []qualifiedColumn{}
			if len(inputs) > 0 {
				projection = append(projection, qualifiedColumn{
					Collection: inputs[0],
					Column:     rule.Supplies + "_addr"})
			}
			for _, o := range rule.Projection {
				projection = append(projection, o)
			}
			rule.Projection = projection
		case collectionTable:
			// no-op
		default:
			// should not get here
			panic(seed.Collections[rule.Supplies].Type)
		}

		new_seed.Rules = append(new_seed.Rules, rule)
	}

	// convert the inputs and outputs into channels
	for _, collection := range new_seed.Collections {
		switch collection.Type {
		case collectionInput, collectionOutput:
			collection.Type = collectionChannel
		case collectionChannel, collectionTable:
			// no-op
		default:
			// shouldn't get here
			panic(collection.Type)
		}
	}

	return transformed
}
