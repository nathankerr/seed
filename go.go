package seed

import (
	"fmt"
	"reflect"
	"strings"
)

func ToGo(seed *Seed, name string) ([]byte, error) {
	str := fmt.Sprintf("package main\n")
	str = fmt.Sprintf(`%s
import (
	"github.com/nathankerr/seed/executor"
	service "github.com/nathankerr/seed"
	"time"
	"flag"
	"log"
)`, str)
	str = fmt.Sprintf("%s\nfunc main() {", str)

	// command line options
	str = fmt.Sprintf(`%s
	var timeout = flag.String("timeout", "", "how long to run; if 0, run forever")
	var sleep = flag.String("sleep", "", "how long to sleep each timestep")
	var address = flag.String("address", "127.0.0.1:3000", "address the bud communicator uses")
	var monitor = flag.String("monitor", "", "address to access the debugger (http), empty means the debugger doesn't run")

	flag.Parse()
`, str)

	// open service
	str = fmt.Sprintf("%s\n\tseed := &service.Seed{", str)

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
	str = fmt.Sprintf(`%s
	var err error
	var sleepDuration time.Duration
	if *sleep != "" {
		sleepDuration, err = time.ParseDuration(*sleep)
		if err != nil {
			log.Fatalln(err)
		}
	}

	var timeoutDuration time.Duration
	if *timeout != "" {
		timeoutDuration, err = time.ParseDuration(*timeout)
		if err != nil {
			log.Fatalln(err)
		}
	}

	executor.Execute(seed, timeoutDuration, sleepDuration, *address, *monitor)
`, str)

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
	str = fmt.Sprintf("%s%sProjection: []service.Expression{\n", str, indent)
	for _, expression := range r.Projection {
		str = fmt.Sprintf("%s%s\t%v,\n", str, indent, expression.toGo(indent+"\t\t"))
	}
	str = fmt.Sprintf("%s%s},\n", str, indent)

	// Predicate
	str = fmt.Sprintf("%s%sPredicate: []service.Constraint{", str, indent)
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
	str := fmt.Sprintf("service.Source{\n")

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

func (expression Expression) toGo(indent string) string {
	str := "service.Expression{ Value:"
	switch value := expression.Value.(type) {
	case QualifiedColumn:
		str = fmt.Sprintf("%s %v", str, value.toGo(indent+"\t\t"))
	case FunctionCall:
		str = fmt.Sprintf("%s %v", str, value.toGo(indent+"\t\t"))
	default:
		panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression.Value).String()))
	}
	str += "}"

	return str
}

func (functionCall FunctionCall) toGo(indent string) string {
	str := fmt.Sprintf("service.FunctionCall{\n%s\tName: \"%s\",", indent, functionCall.Name)
	str = fmt.Sprintf("%s\n%s\tFunction: %s,", str, indent, functionCall.Name)

	arguments := []string{}
	for _, argument := range functionCall.Arguments {
		arguments = append(arguments,
			fmt.Sprintf("%s", argument.toGo(indent+"\t\t")))
	}
	str = fmt.Sprintf("%s\n%s\tArguments: []service.QualifiedColumn{%s},", str, indent, strings.Join(arguments, ",\n"))

	str = fmt.Sprintf("%s\n%s}", str, indent,)
	return str
}