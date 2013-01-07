package executor

import (
	"encoding/json"
	"fmt"
	"reflect"
	"service"
	"strings"
)

type collection struct {
	name    string                   // collection name
	key     []string                 // key column names
	columns map[string]int           // column names -> array indexes
	rows    map[string][]interface{} // storage for the data
}

// create a new collection
// example:
//
// newCollectionFromRaw(
// 	"kvs",
// 	[]string{"key"},
// 	[]string{"value"},
// 	[][]interface{}{
// 		[]interface{}{1, 2},
// 		[]interface{}{3, 4},
// 	},
// )
func newCollectionFromRaw(name string, key []string, data []string, rows [][]interface{}) *collection {
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
	c.addRows(rows)

	return c
}

// add rows to c
func (c *collection) addRows(rows [][]interface{}) {
	for _, row := range rows {
		c.addRow(row)
	}
}

// add a row to the collection
func (c *collection) addRow(row []interface{}) {
	len_columns := len(c.columns)
	if len(row) != len_columns {
		panic(fmt.Sprintf("row %v does not have the correct number of columns", row))
	}
	c.rows[c.key_for(row)] = row
}

// create a collection from a service collection
func newCollection(name string, from *service.Collection) *collection {
	c := &collection{
		name:    name,
		key:     from.Key,
		columns: make(map[string]int),
		rows:    make(map[string][]interface{}),
	}

	// fill in the columns mapping
	column_index := 0
	for _, column_name := range from.Key {
		c.columns[column_name] = column_index
		column_index++
	}
	for _, column_name := range from.Data {
		c.columns[column_name] = column_index
		column_index++
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
		panic(fmt.Sprintf("merge not possible from\n\t%v\nto\n\t%v\nas the collections have different numbers of columns.", to_merge, c))
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

// pretty print data
func (c *collection) String() string {
	rows := []string{}
	for _, row := range c.rows {
		columns := []string{}
		for _, column := range row {
			switch columnTyped := column.(type) {
			case []uint8:
				columns = append(columns, string(columnTyped))
			case int8:
				columns = append(columns, fmt.Sprintf("%#v", columnTyped))
			default:
				panic("unhandled type:" + reflect.TypeOf(column).String())
			}
		}
		rows = append(rows, fmt.Sprintf("[%s]", strings.Join(columns, ", ")))
	}

	return fmt.Sprintf("[%s]", strings.Join(rows, ", "))
}
