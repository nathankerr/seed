package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type bud struct {
	collections budTableCollection
	rules       []*rule
}

func newBud() *bud {
	return &bud{
		collections: newBudTableCollection(),
	}
}

func (b *bud) String() string {
	str := "collections:"
	for name, collection := range b.collections {
		str += fmt.Sprintf("\n\t%s (%s):\n\t\t%s", name, collection.source, collection)
	}

	str += "\nrules:"
	for k, v := range b.rules {
		str = fmt.Sprint(str, "\n\t", k, "\t", v, "\t(", v.source, ")")
	}

	return str
}

type budCollection map[string]*bud

func newBudCollection() budCollection {
	return make(map[string]*bud)
}

func (buds *budCollection) toRuby(dir string) error {
	dir = filepath.Clean(dir)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	for name, bud := range *buds {
		filename := filepath.Join(dir, strings.ToLower(name)+".rb")
		out, err := os.Create(filename)
		if err != nil {
			return err
		}

		fmt.Fprintln(out, "require 'rubygems'")
		fmt.Fprintln(out, "require 'bud'")

		fmt.Fprintf(out, "\nmodule %s\n", name)
		fmt.Fprintf(out, "  include Bud\n")

		fmt.Fprintf(out, "\n  state do\n")
		for _, collection := range bud.collections {
			fmt.Fprintf(out, "    %s #%s\n", collection, collection.source)
		}
		fmt.Fprintf(out, "  end\n")

		fmt.Fprintf(out, "\n  bloom do\n")
		for _, rule := range bud.rules {
			fmt.Fprintf(out, "    %s #%s\n", rule, rule.source)
		}
		fmt.Fprintf(out, "  end\n")

		fmt.Fprintf(out, "end\n")
		out.Close()
	}

	return nil
}
