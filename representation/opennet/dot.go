package opennet

import (
	"bytes"
	"fmt"
	"github.com/nathankerr/seed"
)

func SeedToDot(seed *seed.Seed, name string) ([]byte, error) {
	return OpenNetToDot(SeedAsOpenNet(seed), name)
}

func OpenNetToDot(net *OpenNet, name string) ([]byte, error) {
	buffer := new(bytes.Buffer)
	fmt.Fprintf(buffer, "digraph %s {", name)
	fmt.Fprintf(buffer, "\n\tmargin=\"0\"")
	fmt.Fprintf(buffer, "\n")

	for placeName, place := range net.Places {
		var style string
		switch place.Type {
		case INTERNAL:
			style = "solid"
		case INPUT, OUTPUT:
			style = "bold"
		default:
			panic(place.Type)
		}

		fmt.Fprintf(buffer, "\n\t%s [shape=\"circle\" style=\"%s\"]", placeName, style)
	}

	for transitionName, transition := range net.Transitions {
		fmt.Fprintf(buffer, "\n\t%s [shape=\"box\"]", transitionName)

		for _, consume := range transition.Consume {
			fmt.Fprintf(buffer, "\n\t%s -> %s", consume, transitionName)
		}

		for _, produce := range transition.Produce {
			fmt.Fprintf(buffer, "\n\t%s -> %s", transitionName, produce)
		}
	}

	fmt.Fprintf(buffer, "\n}")
	return buffer.Bytes(), nil
}
