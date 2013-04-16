package seed

import(
	"fmt"
	"reflect"
)

func (this *Seed) EquivalentTo(that *Seed) error {
	// Name
	if this.Name != that.Name {
		return fmt.Errorf("different names: expected %#v, got %#v", this.Name, that.Name)
	}

	// Collections
	if len(this.Collections) != len(that.Collections) {
		return fmt.Errorf("different number of collections: expected %#v, got %#v", len(this.Collections), len(that.Collections))
	}

	for name, thisCollection := range this.Collections {
		thatCollection, exists := that.Collections[name]
		if !exists {
			return fmt.Errorf("expected collection %#v not found", name)
		}

		err := thisCollection.EquivalentTo(thatCollection)
		if err != nil {
			// return err
		}
	}

	// Rules
	if len(this.Rules) != len(that.Rules) {
		return fmt.Errorf("different number of rules: expected %d, got %d", len(this.Rules), len(that.Rules))
	}
	for ruleNum, thisRule := range this.Rules {
		thatRule := that.Rules[ruleNum]

		err := thisRule.EquivalentTo(thatRule)
		if err != nil {
			return err
		}
	}

	// Source
	err := this.Source.EquivalentTo(&that.Source)
	if err != nil {
		return err
	}

	return nil
}

func (this *Collection) EquivalentTo(that *Collection) error {
	if !reflect.DeepEqual(this, that) {
		return fmt.Errorf("Collections are not equivalent: expected %#v, got %#v", this, that)
	}
	return nil
}

func (this *Rule) EquivalentTo(that *Rule) error {
	if !reflect.DeepEqual(this, that) {
		return fmt.Errorf("Rules are not equivalent: expected\n%#v\n, got\n%#v", this, that)
	}
	return nil
}

func (this *Source) EquivalentTo(that *Source) error {
	if !reflect.DeepEqual(this, that) {
		return fmt.Errorf("Sources are not equivalent: expected %#v, got %#v", this, that)
	}

	return nil
}
