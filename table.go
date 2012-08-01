package main

import (
	"fmt"
)

type table struct {
	key     []string
	columns []string
	source  source
}

func newTable() *table {
	return &table{}
}

func (t *table) String() string {
	return fmt.Sprint(t.key, "=>", t.columns)
}

type tableCollection map[string]*table

func newTableCollection() tableCollection {
	return make(map[string]*table)
}
