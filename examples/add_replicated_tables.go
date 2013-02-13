package examples

import (
	"fmt"
	service "github.com/nathankerr/seed"
)

func Add_replicated_tables(name string, orig *service.Service, services map[string]*service.Service) map[string]*service.Service {
	// info()

	// find an existing service to modify or create a new one
	seed, ok := services[name]
	if !ok {
		seed = &service.Service{Collections: make(map[string]*service.Collection), Source: orig.Source}
	}

	// add helper tables, rules
	for tname, table := range orig.Collections {
		seed.Collections[tname] = table
		if table.Type != service.CollectionTable {
			continue
		}

		// create the table for replicants of this table
		replicants_name := fmt.Sprintf("%s_replicants", tname)
		seed.Collections[replicants_name] = &service.Collection{
			Type:   service.CollectionTable,
			Key:    []string{"address"},
			Source: table.Source,
		}

		// add collections needed to handle each operation type
		for _, operation := range []string{"insert", "delete", "update"} {
			// scratch used to intercept the table operation
			scratch_name := fmt.Sprintf("%s_%s", tname, operation)
			seed.Collections[scratch_name] = &service.Collection{
				Type:   service.CollectionScratch,
				Key:    table.Key,
				Data:   table.Data,
				Source: table.Source,
			}

			// channel used for inter-replicant communication
			channel_name := fmt.Sprintf("%s_%s_channel", tname, operation)
			channel := &service.Collection{
				Type:   service.CollectionChannel,
				Key:    []string{"@address"},
				Data:   table.Data,
				Source: table.Source,
			}
			for _, column := range table.Key {
				channel.Key = append(channel.Key, column)
			}
			seed.Collections[channel_name] = channel

			// rule to forward scratch to table
			scratch_to_table := &service.Rule{
				Supplies: tname,
				Source:   table.Source,
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
				scratch_to_table.Projection = append(scratch_to_table.Projection,
					service.QualifiedColumn{
						Collection: scratch_name,
						Column:     column,
					})
			}
			for _, column := range table.Data {
				scratch_to_table.Projection = append(scratch_to_table.Projection,
					service.QualifiedColumn{
						Collection: scratch_name,
						Column:     column,
					})
			}
			seed.Rules = append(seed.Rules, scratch_to_table)

			// rule to forward scratch to channel
			scratch_to_channel := &service.Rule{
				Supplies:  channel_name,
				Operation: "<~",
				Projection: []service.QualifiedColumn{service.QualifiedColumn{
					Collection: replicants_name,
					Column:     "address",
				}},
				Source: table.Source,
			}
			for _, column := range table.Key {
				scratch_to_channel.Projection = append(scratch_to_channel.Projection,
					service.QualifiedColumn{
						Collection: scratch_name,
						Column:     column,
					})
			}
			for _, column := range table.Data {
				scratch_to_channel.Projection = append(scratch_to_channel.Projection,
					service.QualifiedColumn{
						Collection: scratch_name,
						Column:     column,
					})
			}
			seed.Rules = append(seed.Rules, scratch_to_channel)

			// rule to forward channel to table
			channel_to_table := &service.Rule{
				Supplies: tname,
				Source:   table.Source,
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
				channel_to_table.Projection = append(channel_to_table.Projection,
					service.QualifiedColumn{
						Collection: channel_name,
						Column:     column,
					})
			}
			for _, column := range table.Data {
				channel_to_table.Projection = append(channel_to_table.Projection,
					service.QualifiedColumn{
						Collection: channel_name,
						Column:     column,
					})
			}
			seed.Rules = append(seed.Rules, channel_to_table)
		}
	}

	// rewrite and append rules from orig
	for _, rule := range orig.Rules {
		// info(rule)
		// rewrite rules feeding tables
		if orig.Collections[rule.Supplies].Type == service.CollectionTable {
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

		seed.Rules = append(seed.Rules, rule)
	}

	services[name] = seed
	return services
}
