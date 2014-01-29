// Represents Seeds as OpenNets
package opennet

import (
	"fmt"
	"github.com/nathankerr/seed"
)

type OpenNet struct {
	Places      map[string]Place // union of I, O, P
	Transitions map[string]Transition
}

type Place struct {
	Type PlaceType
	*seed.Collection
}

type PlaceType int

const (
	INTERNAL PlaceType = iota
	INPUT
	OUTPUT
)

type Transition struct {
	Consume []string
	Produce []string
	*seed.Rule
}

func SeedAsOpenNet(s *seed.Seed) *OpenNet {
	net := &OpenNet{
		Places:      map[string]Place{},
		Transitions: map[string]Transition{},
	}

	for collectionName, collection := range s.Collections {
		place := Place{
			Collection: collection,
		}
		switch collection.Type {
		case seed.CollectionInput:
			place.Type = INPUT
		case seed.CollectionOutput:
			place.Type = OUTPUT
		case seed.CollectionTable, seed.CollectionScratch:
			place.Type = INTERNAL
		case seed.CollectionChannel:
			// creates two places, one for input and one for output
			net.Places[collectionName+"_channel_output"] = Place{
				Type:       OUTPUT,
				Collection: collection,
			}

			collectionName += "_channel_input"
			place.Type = INPUT
		}

		net.Places[collectionName] = place
	}

	for ruleNumber, rule := range s.Rules {
		transition := Transition{
			Rule: rule,
		}

		// fill transition.Consumes
		for _, collectionName := range rule.Requires() {
			collection := s.Collections[collectionName]
			switch collection.Type {
			case seed.CollectionInput, seed.CollectionOutput, seed.CollectionTable, seed.CollectionScratch:
				transition.Consume = append(transition.Consume, collectionName)
			case seed.CollectionChannel:
				transition.Consume = append(transition.Consume, collectionName+"_channel_input")
			default:
				panic(collection.Type)
			}
		}

		// fill transition.Produce
		collection := s.Collections[rule.Supplies]
		switch collection.Type {
		case seed.CollectionInput, seed.CollectionOutput, seed.CollectionTable, seed.CollectionScratch:
			transition.Produce = append(transition.Produce, rule.Supplies)
		case seed.CollectionChannel:
			transition.Produce = append(transition.Produce, rule.Supplies+"_channel_output")
		default:
			panic(collection.Type)
		}

		net.Transitions[fmt.Sprint("rule", ruleNumber)] = transition
	}

	return net
}
