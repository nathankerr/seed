package service

import (
	"fmt"
)

func ToGo(seed *Service, name string) ([]byte, error) {
	str := fmt.Sprintf("package main\n")
	str = fmt.Sprintf("%s\nimport (\n\t\"github.com/nathankerr/seed/executor\"\n\t\"github.com/nathankerr/seed/service\"\n)\n", str)
	str = fmt.Sprintf("%s\nfunc main() {", str)

	// open service
	str = fmt.Sprintf("%s\n\tseed := &service.Service{", str)

	// source
	str = fmt.Sprintf("%s\n\t\tSource: %v,", str, seed.Source.toGo("\t\t\t"))

	// collections
	str = fmt.Sprintf("%s\n\t\tCollections: map[string]*service.Collection{", str)
	for name, collection := range seed.Collections {
		str = fmt.Sprintf("%s\n\t\t\t\"%s\": %v,", str, name, collection.toGo("\t\t\t\t"))
	}
	str = fmt.Sprintf("%s\n\t\t},", str)

	// rules
	str = fmt.Sprintf("%s\n\t\tRules: []*service.Rule{", str)
	for _, rule := range seed.Rules {
		str = fmt.Sprintf("%s\n\t\t\t%v,", str, rule.toGo("\t\t\t\t"))
	}
	str = fmt.Sprintf("%s\n\t\t},", str)

	// close service
	str = fmt.Sprintf("%s\n\t}", str)

	// execute
	str = fmt.Sprintf("%s\n\n\texecutor.Execute(seed)", str)

	// close main
	str = fmt.Sprintf("%s\n}\n", str)

	return []byte(str), nil
}

func (c *Collection) toGo(indent string) string {
	str := fmt.Sprintf("&service.Collection{\n")

	// type
	typestr := ""
	switch c.Type {
	case CollectionInput:
		typestr = "CollectionInput"
	case CollectionOutput:
		typestr = "CollectionOutput"
	case CollectionTable:
		typestr = "CollectionTable"
	case CollectionScratch:
		typestr = "CollectionScratch"
	case CollectionChannel:
		typestr = "CollectionChannel"
	default:
		panic(fmt.Sprintf("unhandled collection type: %d", c.Type))
	}
	str = fmt.Sprintf("%s%sType: service.%s,\n", str, indent, typestr)

	// key
	str = fmt.Sprintf("%s%sKey:  %#v,\n", str, indent, c.Key)

	// data
	str = fmt.Sprintf("%s%sData: %#v,\n", str, indent, c.Data)

	// source
	str = fmt.Sprintf("%s%sSource: %v,\n", str, indent, c.Source.toGo(indent+"\t"))

	if len(indent) > 0 {
		indent = indent[:len(indent)-1]
	}
	str = fmt.Sprintf("%s%s}", str, indent)
	return str
}

func (r *Rule) toGo(indent string) string {
	str := fmt.Sprintf("&service.Rule{\n")

	// Supplies
	str = fmt.Sprintf("%s%sSupplies:  %#v,\n", str, indent, r.Supplies)

	// Operation
	str = fmt.Sprintf("%s%sOperation: %#v,\n", str, indent, r.Operation)

	// Projection
	str = fmt.Sprintf("%s%sProjection: []QualifiedColumn{\n", str, indent)
	for _, qc := range r.Projection {
		str = fmt.Sprintf("%s%s\t%v,\n", str, indent, qc.toGo(indent+"\t\t"))
	}
	str = fmt.Sprintf("%s%s},\n", str, indent)

	// Predicate
	str = fmt.Sprintf("%s%sPredicate: []Constraint{", str, indent)
	for _, c := range r.Predicate {
		str = fmt.Sprintf("%s\n%s\t%v,\n", str, indent, c.toGo(indent+"\t\t"))
	}
	if len(r.Predicate) > 0 {
		str = fmt.Sprintf("%s%s},\n", str, indent)
	} else {
		str = fmt.Sprintf("%s},\n", str)
	}

	// Source
	str = fmt.Sprintf("%s%sSource: %v,\n", str, indent, r.Source.toGo(indent+"\t"))

	if len(indent) > 0 {
		indent = indent[:len(indent)-1]
	}
	str = fmt.Sprintf("%s%s}", str, indent)
	return str
}

func (s *Source) toGo(indent string) string {
	str := fmt.Sprintf("&service.Source{\n")

	// Name
	str = fmt.Sprintf("%s%sName:   %#v,\n", str, indent, s.Name)

	// Line
	str = fmt.Sprintf("%s%sLine:   %#v,\n", str, indent, s.Line)

	// Column
	str = fmt.Sprintf("%s%sColumn: %#v,\n", str, indent, s.Column)

	if len(indent) > 0 {
		indent = indent[:len(indent)-1]
	}
	str = fmt.Sprintf("%s%s}", str, indent)
	return str
}

func (qc QualifiedColumn) toGo(indent string) string {
	str := fmt.Sprintf("service.QualifiedColumn{\n")

	// Collection
	str = fmt.Sprintf("%s%sCollection: %#v,\n", str, indent, qc.Collection)

	// Column
	str = fmt.Sprintf("%s%sColumn:     %#v,\n", str, indent, qc.Column)

	if len(indent) > 0 {
		indent = indent[:len(indent)-1]
	}
	str = fmt.Sprintf("%s%s}", str, indent)
	return str
}

func (c Constraint) toGo(indent string) string {
	str := fmt.Sprintf("service.Constraint{")

	// Left
	str = fmt.Sprintf("%s\n%sLeft: %v,\n", str, indent, c.Left.toGo(indent+"\t"))

	// Right
	str = fmt.Sprintf("%s%sRight: %v,\n", str, indent, c.Right.toGo(indent+"\t"))

	if len(indent) > 0 {
		indent = indent[:len(indent)-1]
	}
	str = fmt.Sprintf("%s%s}", str, indent)
	return str
}
