package main

type rule struct {
	value    string
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
