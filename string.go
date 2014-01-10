package seed

import (
	"fmt"
	"reflect"
	"strings"
)

func (qc QualifiedColumn) String() string {
	return fmt.Sprintf("%s.%s", qc.Collection, qc.Column)
}

func (c *Constraint) String() string {
	return fmt.Sprintf("%s => %s", c.Left.String(), c.Right.String())
}

func (s *Seed) String() string {
	str := "\nCollections:"
	for cname, collection := range s.Collections {
		str = fmt.Sprintf("%s\n\t%s", str, collection.String(cname))
	}

	str = fmt.Sprintf("%s\nRules:", str)
	for rnum, rule := range s.Rules {
		str = fmt.Sprintf("%s\n%d\t%s", str, rnum, rule)
	}

	return str
}

func (c *Collection) String(cname string) string {
	var ctype string
	switch c.Type {
	case CollectionInput:
		ctype = "input"
	case CollectionOutput:
		ctype = "output"
	case CollectionTable:
		ctype = "table"
	case CollectionChannel:
		ctype = "channel"
	case CollectionScratch:
		ctype = "scratch"
	default:
		// shouldn't get here
		panic(c.Type)
	}

	var key string
	if len(c.Key) > 0 {
		key = fmt.Sprintf("[%s]",
			strings.Join(c.Key, ", "))
	}

	var values string
	if len(c.Data) > 0 {
		values = fmt.Sprintf("=> [%s]",
			strings.Join(c.Data, ", "))
	}

	return fmt.Sprintf("%s %s %s %s",
		ctype,
		cname,
		key,
		values,
	)
}

func (ctype CollectionType) String() string {
	switch ctype {
	case CollectionInput:
		return "input"
	case CollectionOutput:
		return "output"
	case CollectionTable:
		return "table"
	case CollectionChannel:
		return "channel"
	case CollectionScratch:
		return "scratch"
	default:
		// shouldn't get here
		panic(fmt.Sprintf("%#v", ctype))
	}
}

func (r *Rule) String() string {
	columns := []string{}
	for _, expression := range r.Projection {
		switch value := expression.(type) {
		case QualifiedColumn:
			columns = append(columns,
				fmt.Sprintf("%s", value))
		case MapFunction:
			arguments := []string{}
			for _, qc := range value.Arguments {
				arguments = append(arguments,
					fmt.Sprintf("%s", qc))
			}
			columns = append(columns,
				fmt.Sprintf("(%s %s)", value.Name, strings.Join(arguments, " ")))
		case ReduceFunction:
			arguments := []string{}
			for _, qc := range value.Arguments {
				arguments = append(arguments,
					fmt.Sprintf("%s", qc))
			}
			columns = append(columns,
				fmt.Sprintf("{%s %s}", value.Name, strings.Join(arguments, " ")))
		default:
			panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression).String()))
		}
	}
	projection := fmt.Sprintf("[%s]",
		strings.Join(columns, ", "),
	)

	predicates := []string{}
	for _, c := range r.Predicate {
		predicates = append(predicates,
			fmt.Sprintf("%v => %s", c.Left, c.Right),
		)
	}
	predicate := ""
	if len(r.Predicate) > 0 {
		predicate = fmt.Sprintf(": %s",
			strings.Join(predicates, ", "),
		)
	}

	return fmt.Sprintf("%s %s %s%s",
		r.Supplies,
		r.Operation,
		projection,
		predicate,
	)
}

func (tuple *Tuple) String() string {
	columns := []string{}
	for _, column := range *tuple {
		switch typed := column.(type) {
		case []byte:
			columns = append(columns, string(typed))
		default:
			columns = append(columns, fmt.Sprintf("%#v", column))
		}
	}

	return fmt.Sprintf("[%s]", strings.Join(columns, ", "))
}
