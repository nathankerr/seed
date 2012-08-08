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

type budTableType int

const (
	budPersistant budTableType = iota
	budChannel
	budInterface
	budScratch
)

type budTable struct {
	typ     budTableType
	name    string
	key     []string
	columns []string
	source  source
	input   bool // only used if typ is budInterface
}

func newBudTable() *budTable {
	return new(budTable)
}

func (t *budTable) String() string {
	declaration := ""

	switch t.typ {
	case budPersistant:
		declaration += "table"
	case budChannel:
		declaration += "channel"
	case budInterface:
		declaration += "interface "
		if t.input {
			declaration += "input"
		} else {
			declaration += "output"
		}
	case budScratch:
		declaration += "scratch"
	default:
		panic("budTable:String: unknown table type: " + string(t.typ))
	}

	declaration += fmt.Sprintf(" :%s, [", t.name)

	for _, v := range t.key {
		declaration += fmt.Sprintf(":%s, ", v)
	}

	if len(t.columns) > 0 {
		declaration = declaration[:len(declaration)-2] + "] => ["

		for _, v := range t.columns {
			declaration += fmt.Sprintf(":%s, ", v)
		}

		declaration = declaration[:len(declaration)-2] + "]"
	} else {
		declaration = declaration[:len(declaration)-2] + "]"
	}

	return declaration
}

type budTableCollection map[string]*budTable

func newBudTableCollection() budTableCollection {
	return make(map[string]*budTable)
}
