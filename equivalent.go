package seed

import (
	"fmt"
	"reflect"
)

// EquivalentTo returns an error if the collections are not equivalent.
// The error string contains the first difference found.
func (s *Seed) EquivalentTo(that *Seed) error {
	// Name
	if s.Name != that.Name {
		return fmt.Errorf("different names: expected %#v, got %#v", s.Name, that.Name)
	}

	// Collections
	if len(s.Collections) != len(that.Collections) {
		return fmt.Errorf("different number of collections: expected %#v, got %#v", len(s.Collections), len(that.Collections))
	}

	for name, sCollection := range s.Collections {
		thatCollection, exists := that.Collections[name]
		if !exists {
			return fmt.Errorf("expected collection %#v not found", name)
		}

		err := sCollection.EquivalentTo(thatCollection)
		if err != nil {
			// return err
		}
	}

	// Rules
	if len(s.Rules) != len(that.Rules) {
		return fmt.Errorf("different number of rules: expected %d, got %d", len(s.Rules), len(that.Rules))
	}
	for ruleNum, sRule := range s.Rules {
		thatRule := that.Rules[ruleNum]

		err := sRule.EquivalentTo(thatRule)
		if err != nil {
			return err
		}
	}

	return nil
}

// EquivalentTo returns an error if both collections are not equivalent.
// The error string contains the first difference found.
func (c *Collection) EquivalentTo(that *Collection) error {
	if !reflect.DeepEqual(c, that) {
		return fmt.Errorf("Collections are not equivalent: expected %#v, got %#v", c, that)
	}
	return nil
}

// EquivalentTo returns an error if both collections are not equivalent.
// The error string contains the first difference found.
func (r *Rule) EquivalentTo(that *Rule) error {
	if !reflect.DeepEqual(r, that) {
		return fmt.Errorf("Rules are not equivalent: expected\n%#v\n, got\n%#v", r, that)
	}
	return nil
}
