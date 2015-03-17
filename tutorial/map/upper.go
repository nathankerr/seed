package main

import (
	"strings"

	"github.com/nathankerr/seed"
)

func upper(input seed.Tuple) seed.Element {
	return strings.ToUpper(input[0].(string))
}
