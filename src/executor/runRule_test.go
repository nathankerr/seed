package main

import (
	"service"
	"testing"
)

func TestProjection(t *testing.T) {
	service := service.Parse("projection test",
		"input in [unimportant, key] => [value]" +
		"table keep [key] => [value]" +
		"keep <+ [in.key, in.value]")

	collections := make(map[string]*collection)

	collections["in"] = newCollection("in", service.Collections["in"])
	collections["in"].addRows([][]interface{}{
			[]interface{}{1, 2, 3},
			[]interface{}{4, 5, 6},
		},
	)

	result := runRule(collections, service, service.Rules[0])
	// t.Errorf("%#v", result)

	expected := &collection{
		name: "keep",
		key: []string{"key"},
		columns: map[string]int{"key":0, "value":1},
		rows: map[string][]interface {}{
			"[2]":[]interface {}{2, 3},
			"[5]":[]interface {}{5, 6},
		},
	}

	if !collectionsAreEquivalent(expected, result) {
		t.Errorf("\nexpected:\t%v\ngot:\t\t%v", expected, result)
	}
}

// determines if two collections are equivalent
func collectionsAreEquivalent(a, b *collection) bool {
	// check the names
	// string
	if a.name != b.name {
		return false
	}

	// check the keys
	// []string
	if len(a.key) != len(b.key) {
		return false
	}
	for i, value := range a.key {
		if b.key[i] != value {
			return false
		}
	}

	// check the columns
	// map[string]int
	if len(a.columns) != len(b.columns) {
		return false
	}
	for columnName, index := range a.columns {
		bindex, ok := b.columns[columnName]
		if !ok {
			return false
		}

		if index != bindex {
			return false
		}
	}

	// check the rows
	// map[string][]interface{}
	if len(b.rows) != len(a.rows) {
		return false
	}
	for key, row := range a.rows {
		bRow, ok := b.rows[key]
		if !ok {
			return false
		}

		// []interface{}
		if len(bRow) != len(row) {
			return false
		}
		for i, column := range row {
			if bRow[i] != column {
				return false
			}
		}
	}

	return true
}