package main

import (
	"encoding/json"
)

type collection struct {
	name    string                   // collection name
	key     []string                 // key column names
	columns map[string]int           // column names -> array indexes
	rows    map[string][]interface{} // storage for the data
}

func newCollection(name string, key []string, data []string, rows [][]interface{}) *collection {
	c := &collection{
		name:    name,
		key:     key,
		columns: make(map[string]int),
		rows:    make(map[string][]interface{}),
	}

	// fill in the columns mapping
	column_index := 0
	for _, column_name := range key {
		c.columns[column_name] = column_index
		column_index++
	}
	for _, column_name := range data {
		c.columns[column_name] = column_index
		column_index++
	}

	// fill in the rows
	len_columns := len(c.columns)
	for i, row := range rows {
		if len(row) != len_columns {
			panic("row " + string(i) + "does not have the correct number of columns")
		}
		c.rows[c.key_for(row)] = row
	}

	return c
}

// creates a string form of the key so the built-in map
// type can be used to ensure the uniqueness property
// of collections
// method: marshal the key columns to json
func (c *collection) key_for(row []interface{}) string {
	bytes_of_key, err := json.Marshal(row[:len(c.key)])
	if err != nil {
		panic(err)
	}

	return string(bytes_of_key)
}

// merge to_merge into c
// both collections need to have the same number of colums
// and have them in the same order
func (c *collection) merge(to_merge *collection) {
	if len(c.columns) != len(to_merge.columns) {
		panic("merge not possible as collections have different numbers of columns")
	}

	for _, row := range to_merge.rows {
		c.rows[c.key_for(row)] = row
	}
}

// delete to_delete from c
func (c *collection) delete(to_delete *collection) {
	if len(c.key) > len(to_delete.columns) {
		panic("delete not possible as the collection containing the rows to delete does not have enough columns to match the key of the main collection")
	}

	for _, row := range to_delete.rows {
		delete(c.rows, c.key_for(row))
	}
}
