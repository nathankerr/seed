package replicate

import (
	"fmt"
	"github.com/nathankerr/seed"
)

// Transform adds table replication using a simple mechanism
func Transform(orig *seed.Seed) (*seed.Seed, error) {
	replicated := &seed.Seed{
		Name:        orig.Name,
		Collections: make(map[string]*seed.Collection),
	}

	handleInsert := false
	handleDelete := false
	handleUpdate := false
	insertOperation := map[string]string{}

	// rewrite and append rules from orig
	for _, rule := range orig.Rules {
		// rewrite rules feeding tables
		if orig.Collections[rule.Supplies].Type == seed.CollectionTable {
			switch rule.Operation {
			case "<+", "<=":
				rule.Supplies += "_insert"
				handleInsert = true
				if insertOperation[rule.Supplies] != "<=" {
					insertOperation[rule.Supplies] = rule.Operation
				}
			case "<-":
				rule.Supplies += "_delete"
				handleDelete = true
			case "<+-":
				rule.Supplies += "_update"
				handleUpdate = true
			default:
				// shouldn't get here
				panic(rule.Operation)
			}
			rule.Operation = "<="
		}

		replicated.Rules = append(replicated.Rules, rule)
	}

	// add helper tables, rules
	for tname, table := range orig.Collections {
		replicated.Collections[tname] = table
		if table.Type != seed.CollectionTable {
			continue
		}

		// create the table for replicants of this table
		replicantsName := fmt.Sprintf("%s_replicants", tname)
		replicated.Collections[replicantsName] = &seed.Collection{
			Type: seed.CollectionTable,
			Key:  []string{"address"},
		}

		// add collections needed to handle each operation type
		toHandle := []string{}
		if handleInsert {
			toHandle = append(toHandle, "insert")
		}
		if handleDelete {
			toHandle = append(toHandle, "update")
		}
		if handleUpdate {
			toHandle = append(toHandle, "delete")
		}

		for _, operation := range toHandle {
			// scratch used to intercept the table operation
			scratchName := fmt.Sprintf("%s_%s", tname, operation)
			replicated.Collections[scratchName] = &seed.Collection{
				Type: seed.CollectionScratch,
				Key:  table.Key,
				Data: table.Data,
			}

			// channel used for inter-replicant communication
			channelName := fmt.Sprintf("%s_%s_channel", tname, operation)
			channel := &seed.Collection{
				Type: seed.CollectionChannel,
				Key:  []string{"@address"},
				Data: table.Data,
			}
			for _, column := range table.Key {
				channel.Key = append(channel.Key, column)
			}
			replicated.Collections[channelName] = channel

			// rule to forward scratch to table
			scratchToTable := &seed.Rule{
				Supplies: tname,
			}
			switch operation {
			case "insert":
				scratchToTable.Operation = insertOperation[scratchToTable.Supplies]
			case "delete":
				scratchToTable.Operation = "<-"
			case "update":
				scratchToTable.Operation = "<+-"
			default:
				// shouldn't get here
				panic(operation)
			}
			for _, column := range table.Key {
				scratchToTable.Projection = append(
					scratchToTable.Projection,
					seed.QualifiedColumn{
						Collection: scratchName,
						Column:     column,
					},
				)
			}
			for _, column := range table.Data {
				scratchToTable.Projection = append(
					scratchToTable.Projection,
					seed.QualifiedColumn{
						Collection: scratchName,
						Column:     column,
					},
				)
			}
			replicated.Rules = append(replicated.Rules, scratchToTable)

			// rule to forward scratch to channel
			scratchToChannel := &seed.Rule{
				Supplies:  channelName,
				Operation: "<~",
				Projection: []seed.Expression{
					seed.QualifiedColumn{
						Collection: replicantsName,
						Column:     "address",
					},
				},
			}
			for _, column := range table.Key {
				scratchToChannel.Projection = append(
					scratchToChannel.Projection,
					seed.QualifiedColumn{
						Collection: scratchName,
						Column:     column,
					},
				)
			}
			for _, column := range table.Data {
				scratchToChannel.Projection = append(
					scratchToChannel.Projection,
					seed.QualifiedColumn{
						Collection: scratchName,
						Column:     column,
					},
				)
			}
			replicated.Rules = append(replicated.Rules, scratchToChannel)

			// rule to forward channel to table
			channelToTable := &seed.Rule{
				Supplies: tname,
			}
			switch operation {
			case "insert":
				channelToTable.Operation = "<+"
			case "delete":
				channelToTable.Operation = "<-"
			case "update":
				channelToTable.Operation = "<+-"
			default:
				// shouldn't get here
				panic(operation)
			}
			for _, column := range table.Key {
				channelToTable.Projection = append(
					channelToTable.Projection,
					seed.QualifiedColumn{
						Collection: channelName,
						Column:     column,
					},
				)
			}
			for _, column := range table.Data {
				channelToTable.Projection = append(
					channelToTable.Projection,
					seed.QualifiedColumn{
						Collection: channelName,
						Column:     column,
					},
				)
			}
			replicated.Rules = append(replicated.Rules, channelToTable)
		}
	}

	return replicated, nil
}
