package seed

import (
	"errors"
	"fmt"
	"reflect"
)

func collection_error_messagef(collection *Collection, format string, args ...interface{}) error {
	format = fmt.Sprintf("Error for collection:\n\t%s\n%s\n", collection, format)
	return error_messagef(format, args...)
}

func rule_error_messagef(rule *Rule, format string, args ...interface{}) error {
	format = fmt.Sprintf("Error for rule:\n\t%s\n%s\n", rule, format)
	return error_messagef(format, args...)
}

func error_messagef(format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return errors.New(message)
}

func (s *Seed) Validate() error {
	for _, collection := range s.Collections {
		// Type should be known
		switch collection.Type {
		case CollectionInput, CollectionOutput, CollectionTable, CollectionScratch, CollectionChannel:
			// known collection types
		default:
			return collection_error_messagef(collection, "Unknown collection type %d", collection.Type)
		}
	}

	for _, rule := range s.Rules {
		// supplies must be a valid collection
		_, ok := s.Collections[rule.Supplies]
		if !ok {
			return rule_error_messagef(rule, "%s is not a known collection.", rule.Supplies)
		}

		// operation must be known
		switch rule.Operation {
		case "<+", "<-", "<+-", "<~", "<=":
			// known operations
		default:
			return rule_error_messagef(rule, "Unknown operation: %s", rule.Operation)
		}

		// the intension should be valid
		for _, expression := range rule.Intension {
			switch value := expression.(type) {
			case QualifiedColumn:
				err := s.validateQualifiedColumn(value, rule)
				if err != nil {
					return err
				}
			case MapFunction:
				for _, qc := range value.Arguments {
					err := s.validateQualifiedColumn(qc, rule)
					if err != nil {
						return err
					}
				}
			case ReduceFunction:
				for _, qc := range value.Arguments {
					err := s.validateQualifiedColumn(qc, rule)
					if err != nil {
						return err
					}
				}
			default:
				panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression).String()))
			}
		}

		// the predicate should refer only to existing columns
		for _, constraint := range rule.Predicate {
			// check existence of left
			err := s.validateQualifiedColumn(constraint.Left, rule)
			if err != nil {
				return err
			}

			// check existence of right
			err = s.validateQualifiedColumn(constraint.Right, rule)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Seed) validateQualifiedColumn(qc QualifiedColumn, rule *Rule) error {
	// collection should exist
	collection, ok := s.Collections[qc.Collection]
	if !ok {
		return rule_error_messagef(rule, "%s is not a known collection.", qc.Collection)
	}

	// column name should exist in the collection
	for _, column := range collection.Key {
		if column == qc.Column {
			return nil
		} else if column[0] == '@' && (string(column[1:]) == qc.Column) {
			return nil
		}
	}
	for _, column := range collection.Data {
		if column == qc.Column {
			return nil
		}
	}

	return rule_error_messagef(rule, "%s does not refer to an existing column", qc)
}

func (s *Seed) InSubset() error {
	// check collection types
	for _, collection := range s.Collections {
		switch collection.Type {
		case CollectionInput, CollectionOutput, CollectionTable:
			// in subset
		case CollectionScratch, CollectionChannel:
			return collection_error_messagef(collection, "Collection type %s not allowed in subset", collection.Type.String())
		default:
			return collection_error_messagef(collection, "Unknown collection type %d", collection.Type)
		}
	}

	// check operations
	for _, rule := range s.Rules {
		switch rule.Operation {
		case "<+", "<-", "<+-":
			// in subset
		case "<~", "<=":
			return rule_error_messagef(rule, "%s operation not allowed in subset", rule.Operation)
		default:
			return rule_error_messagef(rule, "Unknown operation: %s", rule.Operation)
		}
	}

	return nil
}
