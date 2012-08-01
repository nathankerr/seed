package main

type bud struct {
	collections tableCollection
	rules       []*rule
}

func newBud() *bud {
	return &bud{
		collections: newTableCollection(),
	}
}

type budCollection map[string]*bud

func newBudCollection() budCollection {
	return make(map[string]*bud)
}

func (buds *budCollection) toRuby(dir string) error {
	return nil
}
