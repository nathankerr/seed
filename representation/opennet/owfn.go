package opennet

import (
	"bytes"
	"fmt"
	"github.com/nathankerr/seed"
	"strings"
)

func SeedToOWFN(seed *seed.Seed, name string) ([]byte, error) {
	return OpenNetToOWFN(SeedAsOpenNet(seed), name)
}

func OpenNetToOWFN(net *OpenNet, name string) ([]byte, error) {
	buffer := new(bytes.Buffer)

	internal := []string{}
	input := []string{}
	output := []string{}
	for placeName, place := range net.Places {
		switch place.Type {
		case INTERNAL:
			internal = append(internal, placeName)
		case INPUT:
			input = append(input, placeName)
		case OUTPUT:
			output = append(output, placeName)
		}
	}
	fmt.Fprintf(buffer, "PLACE\n")
	fmt.Fprintf(buffer, "\tINTERNAL\n\t\t%s;\n", strings.Join(internal, ", "))
	fmt.Fprintf(buffer, "\tINPUT\n\t\t%s;\n", strings.Join(input, ", "))
	fmt.Fprintf(buffer, "\tOUTPUT\n\t\t%s;\n", strings.Join(output, ", "))

	fmt.Fprint(buffer, "\n\nINITIALMARKING\n\t;")
	fmt.Fprint(buffer, "\n\nFINALCONDITION\n\t;")

	fmt.Fprint(buffer, "\n\n")
	for transitionName, transition := range net.Transitions {
		fmt.Fprintf(buffer, "TRANSITION %s\n\tCONSUME %s;\n\tPRODUCE %s;\n",
			transitionName,
			strings.Join(transition.Consume, ", "),
			strings.Join(transition.Produce, ", "),
		)
	}

	return buffer.Bytes(), nil
}
