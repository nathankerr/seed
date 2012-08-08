package main

type ruleType int

const (
	ruleInsert ruleType = iota
	ruleSet
	ruleDelete
	ruleUpdate
)

type rule struct {
	value    string
	typ ruleType
	supplies []string
	requires []string
	source   source
}

func newRule() *rule {
	return new(rule)
}

func (r *rule) String() string {
	return r.value
}
