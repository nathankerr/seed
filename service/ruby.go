package service

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ToBloom(seed *Service, name string) ([]byte, error) {
	return []byte(seed.toRuby(name)), nil
}

func (s *Service) toRuby(name string) string {
	info()

	name = strings.Title(name)
	str := fmt.Sprintf("module %s\n", name)
	collections := s.Collections

	rules := fmt.Sprintf("\n  bloom do\n")
	for rule_num, rule := range s.Rules {
		rule_str, additional_collections := rule.Ruby(s)
		rules = fmt.Sprintf("%s    %s #%s rule %d\n", rules, rule_str, rule.Source, rule_num)
		for cname, collection := range additional_collections {
			collections[cname] = collection
		}
	}
	rules = fmt.Sprintf("%s  end\n", rules)

	str = fmt.Sprintf("%s  state do\n", str)
	for cname, collection := range collections {
		str = fmt.Sprintf("%s    %s #%s\n", str, collection.Ruby(cname),
			collection.Source)
	}
	str = fmt.Sprintf("%s  end\n", str)

	str += rules

	str = fmt.Sprintf("%send\n", str)

	return str
}

func (r *Rule) Ruby(service *Service) (string, map[string]*Collection) {
	var selecter string

	additional_collections := make(map[string]*Collection)
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
		selecter = fmt.Sprintf("%s do |%s|\n      [%s]\n    end",
			r.Projection[0].Collection,
			strings.Join(names, ", "),
			strings.Join(output, ", "))
	} else {
		predicates := []string{}
		for _, p := range r.Predicate {
			predicates = append(predicates, p.String())
		}

		selecter = fmt.Sprintf("(%s).combos(%s) do |%s|\n      [%s]\n    end",
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
		scratch_collection := &Collection{
			Type: CollectionScratch,
			// Key: service.Collections[r.Supplies].Key,
			// Data: service.Collections[r.Supplies].Data,
			Source: r.Source,
		}
		// for _, column := range service.Collections[r.Supplies].Key {
		// 	scratch_collection.Key = append(scratch_collection.Key,
		// 		strings.Replace(column, "@", "", -1))
		// }
		for _, column := range r.Projection {
			scratch_collection.Key = append(scratch_collection.Key,
				column.Column)
		}
		additional_collections[scratch_name] = scratch_collection
		scratch := fmt.Sprintf("%s <= %s #%s",
			scratch_name, selecter, r.Source)
		indented_block := strings.Replace(r.Block, "\n", "\n    ", -1)
		indented_block = strings.Replace(indented_block, "\t", "  ", -1)
		if r.Block[0] == 'm' {
			// map block
			return fmt.Sprintf("%s\n    %s %s %s do %s",
					scratch, r.Supplies, r.Operation, scratch_name, indented_block[4:]),
				additional_collections
		} else {
			// reduce block
			return fmt.Sprintf("%s\n    %s %s %s.reduce({}) do %s", scratch,
					r.Supplies, r.Operation, scratch_name, indented_block[7:]),
				additional_collections
		}
	}

	return fmt.Sprintf("%s %s %s",
			r.Supplies,
			r.Operation,
			selecter),
		additional_collections
}

func (c *Collection) Ruby(name string) string {
	var declaration string

	switch c.Type {
	case CollectionInput:
		declaration = "interface input,"
	case CollectionOutput:
		declaration = "interface output,"
	case CollectionChannel:
		declaration = "channel"
	case CollectionTable:
		declaration = "table"
	case CollectionScratch:
		declaration = "scratch"
	default:
		// shouldn't get here
		panic(c.Type)
	}

	declaration = fmt.Sprintf("%s :%s, [", declaration, name)

	for _, v := range c.Key {
		declaration += fmt.Sprintf(":%s, ", v)
	}
	if len(c.Key) != 0 {
		declaration = declaration[:len(declaration)-2]
	}

	if len(c.Data) > 0 {
		declaration += "] => ["

		for _, v := range c.Data {
			declaration += fmt.Sprintf(":%s, ", v)
		}

		declaration = declaration[:len(declaration)-2] + "]"
	} else {
		declaration += "]"
	}

	return declaration
}
