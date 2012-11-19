package main

import (
	"fmt"
)

func add_replicated_tables(name string, orig *service, services map[string]*service) map[string]*service {
	info()

	// find an existing service to modify or create a new one
	seed, ok := services[name]
	if !ok {
		seed = &service{collections: make(map[string]*collection)}
	}

	// add helper tables, rules
	for tname, table := range orig.collections {
		seed.collections[tname] = table
		if table.ctype != collectionTable {
			continue
		}

		// create the table for replicants of this table
		replicants_name := fmt.Sprintf("%s_replicants", tname)
		seed.collections[replicants_name] = &collection{
			ctype: collectionTable,
			key: []string{"address"},
			source: table.source,
		}

		// add collections needed to handle each operation type
		for _, operation := range []string{"insert", "delete", "update"} {
			// scratch used to intercept the table operation
			scratch_name := fmt.Sprintf("%s_%s", tname, operation)
			seed.collections[scratch_name] = &collection{
				ctype: collectionScratch,
				key: table.key,
				data: table.data,
				source: table.source,
			}

			// channel used for inter-replicant communication
			channel_name := fmt.Sprintf("%s_%s_channel", tname, operation)
			channel := &collection{
				ctype: collectionChannel,
				key: []string{"@address"},
				data: table.data,
				source: table.source,
			}
			for _, column := range table.key {
				channel.key = append(channel.key, column)
			}
			seed.collections[channel_name] = channel

			// rule to forward scratch to table
			scratch_to_table :=  &rule{
				supplies: tname,
				operation: "<+",
				source: table.source,
				}
			for _, column := range table.key {
				scratch_to_table.projection = append(scratch_to_table.projection,
					qualifiedColumn{
						collection: scratch_name,
						column: column,
					})
			}
			for _, column := range table.data {
				scratch_to_table.projection = append(scratch_to_table.projection,
					qualifiedColumn{
						collection: scratch_name,
						column: column,
					})
			}
			seed.rules = append(seed.rules, scratch_to_table)

			// rule to forward scratch to channel
			scratch_to_channel :=  &rule{
				supplies: channel_name,
				operation: "<~",
				projection: []qualifiedColumn{qualifiedColumn{
					collection: replicants_name,
					column: "address",
					}},
				source: table.source,
				}
			for _, column := range table.key {
				scratch_to_channel.projection = append(scratch_to_channel.projection,
					qualifiedColumn{
						collection: scratch_name,
						column: column,
						})
			}
			for _, column := range table.data {
				scratch_to_channel.projection = append(scratch_to_channel.projection,
					qualifiedColumn{
						collection: scratch_name,
						column: column,
						})
			}
			seed.rules = append(seed.rules, scratch_to_channel)

			// rule to forward channel to table
			channel_to_table :=  &rule{
				supplies: tname,
				source: table.source,
				}
			switch operation {
			case "insert":
				channel_to_table.operation = "<+"
			case "delete":
				channel_to_table.operation = "<-"
			case "update":
				channel_to_table.operation = "<+-"
			default:
				// shouldn't get here
				panic(operation)
			}
			for _, column := range table.key {
				channel_to_table.projection = append(channel_to_table.projection,
					qualifiedColumn{
						collection: channel_name,
						column: column,
						})
			}
			for _, column := range table.data {
				channel_to_table.projection = append(channel_to_table.projection,
					qualifiedColumn{
						collection: channel_name,
						column: column,
						})
			}
			seed.rules = append(seed.rules, channel_to_table)
		}
	}

	// rewrite and append rules from orig
	for _, rule := range orig.rules {
		info(rule)
		// rewrite rules feeding tables
		if orig.collections[rule.supplies].ctype == collectionTable {
			switch rule.operation {
			case "<+":
				rule.supplies += "_insert"
			case "<-":
				rule.supplies += "_delete"
			case "<+-":
				rule.supplies += "_update"
			default:
				// shouldn't get here
				panic(rule.operation)
			}
			rule.operation = "<="
		}

		seed.rules = append(seed.rules, rule)
	}

	services[name] = seed
	return services
}