package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func (buds *budCollection) toRuby(dir string) error {
	dir = filepath.Clean(dir)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	for name, bud := range *buds {
		filename := filepath.Join(dir, strings.ToLower(name)+".rb")
		out, err := os.Create(filename)
		if err != nil {
			return err
		}

		fmt.Fprintln(out, "require 'rubygems'")
		fmt.Fprintln(out, "require 'bud'")

		fmt.Fprintf(out, "\nmodule %s\n", name)
		fmt.Fprintf(out, "  include Bud\n")

		fmt.Fprintf(out, "\n  state do\n")
		for _, collection := range bud.collections {
			fmt.Fprintf(out, "    %s #%s\n", collection, collection.source)
		}
		fmt.Fprintf(out, "  end\n")

		fmt.Fprintf(out, "\n  bloom do\n")
		for _, rule := range bud.rules {
			fmt.Fprintf(out, "    %s #%s\n", rule, rule.source)
		}
		fmt.Fprintf(out, "  end\n")

		fmt.Fprintf(out, "end\n")
		out.Close()
	}

	return nil
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
