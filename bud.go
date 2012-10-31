package main

import (
	"fmt"
)

type bud struct {
	collections budTableCollection
	rules       []*rule
}

func newBud() *bud {
	return &bud{
		collections: newBudTableCollection(),
	}
}

func (b *bud) String() string {
	str := "collections:"
	for name, collection := range b.collections {
		str += fmt.Sprintf("\n\t%s (%s):\n\t\t%s", name, collection.source, collection)
	}

	str += "\nrules:"
	for k, v := range b.rules {
		str = fmt.Sprint(str, "\n\t", k, "\t", v, "\t(", v.source, ")")
	}

	return str
}

type budCollection map[string]*bud

func newBudCollection() budCollection {
	return make(map[string]*bud)
}

func (buds budCollection) String() string {
	str := ""
	for name, bud := range buds {
		str = fmt.Sprintf("%s\n### bud: %s ###\n%s\n", str, name, bud)
	}
	return str
}

type budTableType int

const (
	budPersistant budTableType = iota
	budChannel
	budInterface
	budScratch
)

var budTableTypeNames = map[budTableType]string{
	budPersistant: "budPersistant",
	budChannel:    "budChannel",
	budInterface:  "budInterface",
	budScratch:    "budScratch",
}

func (typ budTableType) String() string {
	str, ok := budTableTypeNames[typ]
	if !ok {
		panic("unknown but table type")
	}
	return str
}

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

func seedTableToBudTable(name string, typ budTableType, t *table) *budTable {
	b := newBudTable()

	b.name = name
	b.typ = typ

	if b.typ == budChannel {
		key := []string{"@address"}
		for _, tkey := range t.key {
			key = append(key, tkey)
		}

		b.key = key
	} else {
		b.key = t.key
	}

	b.columns = t.columns
	b.source = t.source

	return b
}
