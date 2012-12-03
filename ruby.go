package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (s *service) toRuby(name string) string {
	name = strings.Title(name)
	str := fmt.Sprintf("module %s\n", name)

	str = fmt.Sprintf("%s  state do\n", str)
	for cname, collection := range s.collections {
		str = fmt.Sprintf("%s    %s #%s\n", str, collection.Ruby(cname),
			collection.source)
	}
	str = fmt.Sprintf("%s  end\n", str)

	str = fmt.Sprintf("%s\n  bloom do\n", str)
	for _, rule := range s.rules {
		str = fmt.Sprintf("%s    %s #%s\n", str, rule.Ruby(), rule.source)
	}
	str = fmt.Sprintf("%s  end\n", str)

	str = fmt.Sprintf("%send\n", str)

	return str
}

func (r *rule) Ruby() string {
	var selecter string

	collections := r.requires()

	index := make(map[string]string)
	names := []string{}
	for i, c := range collections {
		name := fmt.Sprintf("c%d", i)
		index[c] = name
		names = append(names, name)
	}

	output := []string{}
	for _, o := range r.projection {
		output = append(output,
			fmt.Sprintf("%s.%s", index[o.collection], o.column))
	}

	if len(collections) == 1 {
		selecter = fmt.Sprintf("%s do |%s| [%s] end",
			r.projection[0].collection,
			strings.Join(names, ", "),
			strings.Join(output, ", "))
	} else {
		predicates := []string{}
		for _, p := range r.predicate {
			predicates = append(predicates, p.String())
		}

		selecter = fmt.Sprintf("(%s).combos(%s) do |%s| [%s] end",
			strings.Join(collections, " * "),
			strings.Join(predicates, ", "),
			strings.Join(names, ", "),
			strings.Join(output, ", "))
	}

	// at this point, str contains the translation of the join and projection
	// ([]: =>) specifier. If there is a block, this needs to be put into a
	// scratch and then the scratch needs the block to be applied to it
	if len(r.block) > 0 {
		scratch_name := r.source.name[:len(r.source.name)-
			len(filepath.Ext(r.source.name))]
		scratch_name = fmt.Sprintf("%s%d_scratch",
			scratch_name, r.source.line)
		scratch := fmt.Sprintf("temp :%s <= %s #%s",
			scratch_name, selecter, r.source)
		indented_block := strings.Replace(r.block, "\n", "\n\t\t", -1)
		if r.block[0] == 'm' {
			// map block
			return fmt.Sprintf("%s\n\t\t%s %s %s %s",
				scratch, r.supplies, r.operation, scratch_name, indented_block)
		} else {
			// reduce block
			return fmt.Sprintf("%s\n\t\t%s %s %s.reduce({}) do %s", scratch,
				r.supplies, r.operation, scratch_name, indented_block[7:])
		}
	}

	return fmt.Sprintf("%s %s %s",
		r.supplies,
		r.operation,
		selecter)
}

func (c *collection) Ruby(name string) string {
	declaration := ""

	switch c.ctype {
	case collectionInput:
		declaration += "interface input,"
	case collectionOutput:
		declaration += "interface output,"
	case collectionChannel:
		declaration += "channel"
	case collectionTable:
		declaration += "table"
	case collectionScratch:
		declaration += "scratch"
	default:
		// shouldn't get here
		panic(c.ctype)
	}

	declaration += fmt.Sprintf(" :%s, [", name)

	for _, v := range c.key {
		declaration += fmt.Sprintf(":%s, ", v)
	}

	if len(c.data) > 0 {
		declaration = declaration[:len(declaration)-2] + "] => ["

		for _, v := range c.data {
			declaration += fmt.Sprintf(":%s, ", v)
		}

		declaration = declaration[:len(declaration)-2] + "]"
	} else {
		declaration = declaration[:len(declaration)-2] + "]"
	}

	return declaration
}
