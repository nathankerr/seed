package service

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (s *Service) ToRuby(name string) string {
	info()

	name = strings.Title(name)
	str := fmt.Sprintf("module %s\n", name)

	str = fmt.Sprintf("%s  state do\n", str)
	for cname, collection := range s.Collections {
		str = fmt.Sprintf("%s    %s #%s\n", str, collection.Ruby(cname),
			collection.Source)
	}
	str = fmt.Sprintf("%s  end\n", str)

	str = fmt.Sprintf("%s\n  bloom do\n", str)
	for rule_num, rule := range s.Rules {
		str = fmt.Sprintf("%s    %s #%s rule %d\n", str, rule.Ruby(), rule.Source, rule_num)
	}
	str = fmt.Sprintf("%s  end\n", str)

	str = fmt.Sprintf("%send\n", str)

	return str
}

func (r *Rule) Ruby() string {
	var selecter string

	collections := r.Requires()

	index := make(map[string]string)
	names := []string{}
	for i, c := range collections {
		name := fmt.Sprintf("c%d", i)
		index[c] = name
		names = append(names, name)
	}

	output := []string{}
	for _, o := range r.Projection {
		output = append(output,
			fmt.Sprintf("%s.%s", index[o.Collection], o.Column))
	}

	if len(collections) == 1 {
		selecter = fmt.Sprintf("%s do |%s| [%s] end",
			r.Projection[0].Collection,
			strings.Join(names, ", "),
			strings.Join(output, ", "))
	} else {
		predicates := []string{}
		for _, p := range r.Predicate {
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
	if len(r.Block) > 0 {
		scratch_name := r.Source.Name[:len(r.Source.Name)-
			len(filepath.Ext(r.Source.Name))]
		scratch_name = fmt.Sprintf("%s%d_scratch",
			scratch_name, r.Source.Line)
		scratch := fmt.Sprintf("temp :%s <= %s #%s",
			scratch_name, selecter, r.Source)
		indented_block := strings.Replace(r.Block, "\n", "\n\t\t", -1)
		if r.Block[0] == 'm' {
			// map block
			return fmt.Sprintf("%s\n\t\t%s %s %s %s",
				scratch, r.Supplies, r.Operation, scratch_name, indented_block)
		} else {
			// reduce block
			return fmt.Sprintf("%s\n\t\t%s %s %s.reduce({}) do %s", scratch,
				r.Supplies, r.Operation, scratch_name, indented_block[7:])
		}
	}

	return fmt.Sprintf("%s %s %s",
		r.Supplies,
		r.Operation,
		selecter)
}

func (c *Collection) Ruby(name string) string {
	declaration := ""

	switch c.Type {
	case CollectionInput:
		declaration += "interface input,"
	case CollectionOutput:
		declaration += "interface output,"
	case CollectionChannel:
		declaration += "channel"
	case CollectionTable:
		declaration += "table"
	case CollectionScratch:
		declaration += "scratch"
	default:
		// shouldn't get here
		panic(c.Type)
	}

	declaration += fmt.Sprintf(" :%s, [", name)

	for _, v := range c.Key {
		declaration += fmt.Sprintf(":%s, ", v)
	}

	if len(c.Data) > 0 {
		declaration = declaration[:len(declaration)-2] + "] => ["

		for _, v := range c.Data {
			declaration += fmt.Sprintf(":%s, ", v)
		}

		declaration = declaration[:len(declaration)-2] + "]"
	} else {
		declaration = declaration[:len(declaration)-2] + "]"
	}

	return declaration
}
