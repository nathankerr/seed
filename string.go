package main

import (
	"fmt"
)

func (qc *qualifiedColumn) String() string {
	return fmt.Sprintf("%s.%s", qc.collection, qc.column)
}

func (c *constraint) String() string {
	return fmt.Sprintf("%s => %s", c.left.String(), c.right.String())
}

func (s source) String() string {
	return fmt.Sprint(s.name, ":", s.line)
}
