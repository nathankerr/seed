package seed

import (
	"fmt"
	"reflect"
	"strings"
)

func ToGo(service *Seed, name string) ([]byte, error) {
	str := fmt.Sprintf("package main\n")
	str = fmt.Sprintf(`%s
import (
	"github.com/nathankerr/seed/executor"
	"github.com/nathankerr/seed/executor/bud"
	"github.com/nathankerr/seed/executor/wsjson"
	"github.com/nathankerr/seed/executor/monitor"
	"github.com/nathankerr/seed"
	"time"
	"flag"
	"log"
)`, str)
	str = fmt.Sprintf("%s\nfunc main() {", str)

	// command line options
	str = fmt.Sprintf(`%s
	var sleep = flag.String("sleep", "", "how long to sleep each timestep")
	var communicator = flag.String("communicator", "bud", "which communicator to use (bud wsjson")
	var address = flag.String("address", ":3000", "address the communicator uses")
	var monitorAddress = flag.String("monitor", "", "address to access the debugger (http), empty means the debugger doesn't run")


	flag.Parse()
`, str)

	// open seed
	str = fmt.Sprintf("%s\n\tservice := &seed.Seed{", str)

	// Name
	str = fmt.Sprintf("%s\n\t\tName: %#v,", str, service.Name)

	// source
	str = fmt.Sprintf("%s\n\t\tSource: %v,", str, service.Source.toGo("\t\t\t"))

	// collections
	str = fmt.Sprintf("%s\n\t\tCollections: map[string]*seed.Collection{", str)
	for name, collection := range service.Collections {
		str = fmt.Sprintf("%s\n\t\t\t\"%s\": %v,", str, name, collection.toGo("\t\t\t\t"))
	}
	str = fmt.Sprintf("%s\n\t\t},", str)

	// rules
	str = fmt.Sprintf("%s\n\t\tRules: []*seed.Rule{", str)
	for _, rule := range service.Rules {
		str = fmt.Sprintf("%s\n\t\t\t%v,", str, rule.toGo("\t\t\t\t"))
	}
	str = fmt.Sprintf("%s\n\t\t},", str)

	// close seed
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

	useMonitor := false
	if *monitorAddress != "" {
		useMonitor = true
	}

	println("Starting " + service.Source.Name + " on " + *address)
	channels := executor.Execute(service, sleepDuration, *address, useMonitor)

	if useMonitor {
		println("Starting monitor" + " on " + *monitorAddress)
		go monitor.StartMonitor(*monitorAddress, channels, service)
	}

	switch *communicator {
	case "bud":
		bud.BudCommunicator(service, channels, *address)
	case "wsjson":
		wsjson.Communicator(service, channels, *address)
	default:
		log.Fatalln("Unknown communicator:", *communicator)
	}
`, str)

	// close main
	str = fmt.Sprintf("%s\n}\n", str)

	return []byte(str), nil
}

func (c *Collection) toGo(indent string) string {
	str := fmt.Sprintf("&seed.Collection{\n")

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
	str = fmt.Sprintf("%s%sType: seed.%s,\n", str, indent, typestr)

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
	str := fmt.Sprintf("&seed.Rule{\n")

	// Supplies
	str = fmt.Sprintf("%s%sSupplies:  %#v,\n", str, indent, r.Supplies)

	// Operation
	str = fmt.Sprintf("%s%sOperation: %#v,\n", str, indent, r.Operation)

	// Projection
	str = fmt.Sprintf("%s%sProjection: []seed.Expression{\n", str, indent)
	for _, expression := range r.Projection {
		str = fmt.Sprintf("%s%s\t%v,\n", str, indent, expression.toGo(indent+"\t\t"))
	}
	str = fmt.Sprintf("%s%s},\n", str, indent)

	// Predicate
	if len(r.Predicate) != 0 {
		str = fmt.Sprintf("%s%sPredicate: []seed.Constraint{", str, indent)
		for _, c := range r.Predicate {
			str = fmt.Sprintf("%s\n%s\t%v,\n", str, indent, c.toGo(indent+"\t\t"))
		}
		str = fmt.Sprintf("%s%s},\n", str, indent)
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
	str := fmt.Sprintf("seed.Source{\n")

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
	str := fmt.Sprintf("seed.QualifiedColumn{\n")

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
	str := fmt.Sprintf("seed.Constraint{")

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
	str := "seed.Expression{ Value:"
	switch value := expression.Value.(type) {
	case QualifiedColumn:
		str = fmt.Sprintf("%s %v", str, value.toGo(indent+"\t\t"))
	case MapFunction:
		str = fmt.Sprintf("%s %v", str, value.toGo(indent+"\t\t"))
	case ReduceFunction:
		str = fmt.Sprintf("%s %v", str, value.toGo(indent+"\t\t"))
	default:
		panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression.Value).String()))
	}
	str += "}"

	return str
}

func (functionCall MapFunction) toGo(indent string) string {
	str := fmt.Sprintf("seed.MapFunction{\n%s\tName: \"%s\",", indent, functionCall.Name)
	str = fmt.Sprintf("%s\n%s\tFunction: %s,", str, indent, functionCall.Name)

	arguments := []string{}
	for _, argument := range functionCall.Arguments {
		arguments = append(arguments,
			fmt.Sprintf("%s", argument.toGo(indent+"\t\t")))
	}
	str = fmt.Sprintf("%s\n%s\tArguments: []seed.QualifiedColumn{\n\t%s%s},", str, indent, indent, strings.Join(arguments, ",\n"+indent+"\t"))

	str = fmt.Sprintf("%s\n%s}", str, indent)
	return str
}

func (functionCall ReduceFunction) toGo(indent string) string {
	str := fmt.Sprintf("seed.ReduceFunction{\n%s\tName: \"%s\",", indent, functionCall.Name)
	str = fmt.Sprintf("%s\n%s\tFunction: %s,", str, indent, functionCall.Name)

	arguments := []string{}
	for _, argument := range functionCall.Arguments {
		arguments = append(arguments,
			fmt.Sprintf("%s", argument.toGo(indent+"\t\t")))
	}
	str = fmt.Sprintf("%s\n%s\tArguments: []seed.QualifiedColumn{\n\t%s%s},", str, indent, indent, strings.Join(arguments, ",\n"+indent+"\t"))

	str = fmt.Sprintf("%s\n%s}", str, indent)
	return str
}
