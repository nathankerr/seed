package main

import (
	"fmt"
)

type dependency struct {
	on     string
	source source
}

type column struct {
	source       source
	dependencies []dependency
}

type analysis struct {
	columns map[string]column
}

func analyze(s *seed) *analysis {
	a := &analysis{make(map[string]column)}

	a.ingest_tables(s.inputs)
	a.ingest_tables(s.outputs)
	a.ingest_tables(s.tables)
	a.ingest_rules(s.rules)

	return a
}

func (a *analysis) ingest_tables(tables tableCollection) {
	for tname, table := range tables {
		for _, cname := range table.key {
			a.columns[tname+"."+cname] = column{source: table.source}
		}

		for _, cname := range table.columns {
			a.columns[tname+"."+cname] = column{source: table.source}
		}
	}
}

func (a *analysis) ingest_rules(rules []*rule) {
	for _, rule := range rules {
		d := dependency{source: rule.source}
		fmt.Println(rule)
	}
}
