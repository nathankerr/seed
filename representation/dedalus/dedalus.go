package dedalus

import (
	"bytes"
	"fmt"
	"github.com/nathankerr/seed"
	"reflect"
	"strings"
)

func SeedToDedalusFile(s *seed.Seed, name string) ([]byte, error) {
	buffer := new(bytes.Buffer)

	for collectionName, collection := range s.Collections {
		schema := []string{}
		for _, columnName := range append(collection.Key, collection.Data...) {
			columnName = strings.Replace(columnName, "@", "#", 1)
			schema = append(schema, columnName)
		}

		switch collection.Type {
		case seed.CollectionInput, seed.CollectionOutput, seed.CollectionScratch, seed.CollectionChannel:
			fmt.Fprintf(buffer, "%[1]s(%[2]s) := %[1]s_pos(%[2]s), ~%[1]s_neg(%[2]s)\n", collectionName, strings.Join(schema, ", "))
		case seed.CollectionTable:
			fmt.Fprintf(buffer, "%[1]s_pos(%[2]s)@next := %[1]s_pos(%[2]s), ~%[1]s_neg(%[2]s)\n", collectionName, strings.Join(schema, ", "))
		default:
			panic(collection.Type)
		}
	}

	for _, rule := range s.Rules {

		// idea:
		// init supplies schema with its own column names
		// create a map of collection names to predicate arrays for each required collection
		// predicate arrays are initialized with each column = "_" to tell dedalus to ignore this column
		// setup a mapping of equivalent column names like was done in fieldgraphs
		// go through the projection, replacing QualifiedColumns with the associated column name of the supplies collection
		// do the same for equivalent column names
		// for map and reduce functions first map the column names, then place the function call in the supplies schema

		suppliesCollection, ok := s.Collections[rule.Supplies]
		if !ok {
			panic(rule.Supplies)
		}

		supplies := []string{}
		for _, columnName := range append(suppliesCollection.Key, suppliesCollection.Data...) {
			columnName = strings.Replace(columnName, "@", "#", 1)
			supplies = append(supplies, columnName)
		}

		predicates := map[string][]string{}
		columnNumbers := map[string]map[string]int{} // collectionName -> columnName -> columnNumber
		for _, collectionName := range rule.Requires() {
			collection, ok := s.Collections[collectionName]
			if !ok {
				panic(collectionName)
			}

			if _, ok := columnNumbers[collectionName]; !ok {
				columnNumbers[collectionName] = map[string]int{}
			}

			for columnNumber, columnName := range append(collection.Key, collection.Data...) {
				predicates[collectionName] = append(predicates[collectionName], "_")
				columnNumbers[collectionName][columnName] = columnNumber
			}
		}

		equivalents := map[string]seed.QualifiedColumn{}
		for _, constraint := range rule.Predicate {
			equivalents[constraint.Left.String()] = constraint.Right
			equivalents[constraint.Right.String()] = constraint.Left
		}

		for i, expression := range rule.Projection {
			switch expression := expression.(type) {
			case seed.QualifiedColumn:
				predicates[expression.Collection][columnNumbers[expression.Collection][expression.Column]] = supplies[i]

				equivalent, ok := equivalents[expression.String()]
				if ok {
					predicates[equivalent.Collection][columnNumbers[equivalent.Collection][equivalent.Column]] = supplies[i]
				}
			case seed.MapFunction:
				println("MapFunction: TODO")
			case seed.ReduceFunction:
				println("ReduceFunction: TODO")
				// arguments := []string{}
				// for _, arguments := range expression.Arguments {
				// 	arguments = append(arguments, )
				// }
			default:
				panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression).String()))
			}
		}

		predicate := []string{}
		for collectionName, columns := range predicates {
			predicate = append(predicate, fmt.Sprintf("%s(%s)", collectionName, strings.Join(columns, ", ")))
		}

		fmt.Fprintf(buffer, "%[1]s(%[2]s) := %s\n", rule.Supplies, strings.Join(supplies, ", "), strings.Join(predicate, ", "))
	}

	return buffer.Bytes(), nil
}
