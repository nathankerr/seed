package examples

import (
	"fmt"
	"github.com/nathankerr/seed"
)

func Add_replicated_tables(orig *seed.Seed) (*seed.Seed, error) {
	replicated := &seed.Seed{
		Name:        orig.Name,
		Collections: make(map[string]*seed.Collection),
	}

	// add helper tables, rules
	for tname, table := range orig.Collections {
		replicated.Collections[tname] = table
		if table.Type != seed.CollectionTable {
			continue
		}

		// create the table for replicants of this table
		replicants_name := fmt.Sprintf("%s_replicants", tname)
		replicated.Collections[replicants_name] = &seed.Collection{
			Type: seed.CollectionTable,
			Key:  []string{"address"},
		}

		// add collections needed to handle each operation type
		for _, operation := range []string{"insert", "delete", "update"} {
			// scratch used to intercept the table operation
			scratch_name := fmt.Sprintf("%s_%s", tname, operation)
			replicated.Collections[scratch_name] = &seed.Collection{
				Type: seed.CollectionScratch,
				Key:  table.Key,
				Data: table.Data,
			}

			// channel used for inter-replicant communication
			channel_name := fmt.Sprintf("%s_%s_channel", tname, operation)
			channel := &seed.Collection{
				Type: seed.CollectionChannel,
				Key:  []string{"@address"},
				Data: table.Data,
			}
			for _, column := range table.Key {
				channel.Key = append(channel.Key, column)
			}
			replicated.Collections[channel_name] = channel

			// rule to forward scratch to table
			scratch_to_table := &seed.Rule{
				Supplies: tname,
			}
			switch operation {
			case "insert":
				scratch_to_table.Operation = "<+"
			case "delete":
				scratch_to_table.Operation = "<-"
			case "update":
				scratch_to_table.Operation = "<+-"
			default:
				// shouldn't get here
				panic(operation)
			}
			for _, column := range table.Key {
				scratch_to_table.Projection = append(
					scratch_to_table.Projection,
					seed.Expression{Value: seed.QualifiedColumn{
						Collection: scratch_name,
						Column:     column,
					}},
				)
			}
			for _, column := range table.Data {
				scratch_to_table.Projection = append(
					scratch_to_table.Projection,
					seed.Expression{Value: seed.QualifiedColumn{
						Collection: scratch_name,
						Column:     column,
					}},
				)
			}
			replicated.Rules = append(replicated.Rules, scratch_to_table)

			// rule to forward scratch to channel
			scratch_to_channel := &seed.Rule{
				Supplies:  channel_name,
				Operation: "<~",
				Projection: []seed.Expression{
					seed.Expression{
						Value: seed.QualifiedColumn{
							Collection: replicants_name,
							Column:     "address",
						},
					},
				},
			}
			for _, column := range table.Key {
				scratch_to_channel.Projection = append(
					scratch_to_channel.Projection,
					seed.Expression{Value: seed.QualifiedColumn{
						Collection: scratch_name,
						Column:     column,
					}},
				)
			}
			for _, column := range table.Data {
				scratch_to_channel.Projection = append(
					scratch_to_channel.Projection,
					seed.Expression{Value: seed.QualifiedColumn{
						Collection: scratch_name,
						Column:     column,
					}},
				)
			}
			replicated.Rules = append(replicated.Rules, scratch_to_channel)

			// rule to forward channel to table
			channel_to_table := &seed.Rule{
				Supplies: tname,
			}
			switch operation {
			case "insert":
				channel_to_table.Operation = "<+"
			case "delete":
				channel_to_table.Operation = "<-"
			case "update":
				channel_to_table.Operation = "<+-"
			default:
				// shouldn't get here
				panic(operation)
			}
			for _, column := range table.Key {
				channel_to_table.Projection = append(
					channel_to_table.Projection,
					seed.Expression{Value: seed.QualifiedColumn{
						Collection: channel_name,
						Column:     column,
					}},
				)
			}
			for _, column := range table.Data {
				channel_to_table.Projection = append(
					channel_to_table.Projection,
					seed.Expression{Value: seed.QualifiedColumn{
						Collection: channel_name,
						Column:     column,
					}},
				)
			}
			replicated.Rules = append(replicated.Rules, channel_to_table)
		}
	}

	// rewrite and append rules from orig
	for _, rule := range orig.Rules {
		// rewrite rules feeding tables
		if orig.Collections[rule.Supplies].Type == seed.CollectionTable {
			switch rule.Operation {
			case "<+":
				rule.Supplies += "_insert"
			case "<-":
				rule.Supplies += "_delete"
			case "<+-":
				rule.Supplies += "_update"
			default:
				// shouldn't get here
				panic(rule.Operation)
			}
			rule.Operation = "<="
		}

		replicated.Rules = append(replicated.Rules, rule)
	}

	return replicated, nil
}
