package main

import (
	"fmt"
	"strings"
)

func (qc qualifiedColumn) String() string {
	return fmt.Sprintf("%s.%s", qc.collection, qc.column)
}

func (c *constraint) String() string {
	return fmt.Sprintf("%s => %s", c.left.String(), c.right.String())
}

func (s source) String() string {
	return fmt.Sprint(s.name, ":", s.line)
}

func (s *service) String() string {
	str := "\nCollections:"
	for cname, collection := range s.collections {
		str = fmt.Sprintf("%s\n\t%s", str, collection.String(cname))
	}

	str = fmt.Sprintf("%s\nRules:", str)
	for rnum, rule := range s.rules {
		str = fmt.Sprintf("%s\n%d\t%s", str, rnum, rule)
	}

	return str
}

func (c *collection) String(cname string) string {
	var ctype string
	switch c.ctype {
	case collectionInput:
		ctype = "input"
	case collectionOutput:
		ctype = "output"
	case collectionTable:
		ctype = "table"
	case collectionChannel:
		ctype = "channel"
	case collectionScratch:
		ctype = "scratch"
	default:
		// shouldn't get here
		panic(c.ctype)
	}

	var key string
	if len(c.key) > 0 {
		key = fmt.Sprintf("[%s]",
			strings.Join(c.key, ", "))
	}

	var values string
	if len(c.data) > 0 {
		values = fmt.Sprintf("=> [%s]",
			strings.Join(c.data, ", "))
	}

	return fmt.Sprintf("%s %s %s %s",
		ctype,
		cname,
		key,
		values,
	)
}

func (ctype collectionType) String() string {
	switch ctype {
	case collectionInput:
		return "input"
	case collectionOutput:
		return "output"
	case collectionTable:
		return "table"
	case collectionChannel:
		return "channel"
	case collectionScratch:
		return "scratch"
	default:
		// shouldn't get here
		panic(ctype)
	}
	return ""
}

func (r *rule) String() string {
	columns := []string{}
	for _, qc := range r.projection {
		columns = append(columns,
			fmt.Sprintf("%s", qc))
	}
	projection := fmt.Sprintf("[%s]",
		strings.Join(columns, ", "),
	)

	predicates := []string{}
	for _, c := range r.predicate {
		predicates = append(predicates,
			fmt.Sprintf("%v => %s", c.left, c.right),
		)
	}
	predicate := ""
	if len(r.predicate) > 0 {
		predicate = fmt.Sprintf(": %s",
			strings.Join(predicates, ", "),
		)
	}

	return fmt.Sprintf("%s %s %s%s",
		r.supplies,
		r.operation,
		projection,
		predicate,
	)
}
