package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

		fmt.Fprintf(out, "\nclass %s\n", name)
		fmt.Fprintf(out, "  include Bud\n")

		fmt.Fprintf(out, "\n  state do\n")
		for _, collection := range bud.collections {
			fmt.Fprintf(out, "    %s #%s\n", collection, collection.source)
		}
		fmt.Fprintf(out, "  end\n")

		fmt.Fprintf(out, "\n  bloom do\n")
		for _, rule := range bud.rules {
			fmt.Fprintf(out, "    %s #%s\n", rule.Ruby(), rule.source)
		}
		fmt.Fprintf(out, "  end\n")

		fmt.Fprintf(out, "end\n")
		out.Close()
	}

	return nil
}

func (r *rule) Ruby() string {
	var selecter string

	collections := []string{}
	for c, _ := range r.requires {
		collections = append(collections, c)
	}

	index := make(map[string]string)
	names := []string{}
	for i, c := range collections {
		name := fmt.Sprintf("c%d", i)
		index[c] = name
		names = append(names, name)
	}

	output := []string{}
	for _, o := range r.output {
		output = append(output, fmt.Sprintf("%s.%s", index[o.collection], o.column))
	}

	if len(r.requires) == 1 {
		selecter = fmt.Sprintf("%s do |%s| [%s] end",
			r.output[0].collection,
			strings.Join(names, ", "),
			strings.Join(output, ", "))
	} else {
		predicates := []string{}
		for _, p := range r.predicates {
			predicates = append(predicates, p.String())
		}

		selecter = fmt.Sprintf("(%s).combos(%s) do |%s| [%s] end",
			strings.Join(collections, " * "),
			strings.Join(predicates, ", "),
			strings.Join(names, ", "),
			strings.Join(output, ", "))
	}

	// at this point, str contains the translation of the join and projection ([]: =>) specifier. If there is a block, this needs to be put into a scratch and then the scratch needs the block to be applied to it
	if len(r.block) > 0 {
		scratch_name := r.source.name[:len(r.source.name)-
			len(filepath.Ext(r.source.name))]
		scratch_name = fmt.Sprintf("%s%d_scratch",
			scratch_name, r.source.line)
		scratch := fmt.Sprintf("temp :%s <= %s #%s",
			scratch_name, selecter, r.source)
		indented_block := strings.Replace(r.block, "\n", "\n\t\t", -1)
		if r.block[0] == 'd' {
			// do block
			return fmt.Sprintf("%s\n\t\t%s %s %s %s",
				scratch, r.supplies, r.typ, scratch_name, indented_block)
		} else {
			// reduce block
			return fmt.Sprintf("%s\n\t\t%s %s %s.reduce({}) do %s",
				scratch, r.supplies, r.typ, scratch_name, indented_block[7:])
		}
	}

	return fmt.Sprintf("%s %s %s",
		r.supplies,
		r.typ,
		selecter)
}
