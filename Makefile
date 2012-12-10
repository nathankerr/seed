.PHONY: all
all: bin/seed
	-rm -rf bud
	bin/seed -dot -json -model kvs.seed

bin/seed: src/*
	GOPATH=/Users/alaster/Projects/seed go install -a seed

.PHONY: view-figures
view-figures: figures
	open "figures/kvs.dot.pdf" "figures/kvs-network.dot.pdf" "figures/kvs-replicated.dot.pdf" "figures/kvs-network-replicated.dot.pdf"

.PHONY: figures
figures: seed
	-rm -rf figures
	mkdir figures
	-rm -rf bud
	bin/seed -dot -network=false -replicate=false kvs.seed
	cp bud/kvs.dot "figures/kvs.dot"
	rm -rf bud
	bin/seed -dot -network=true -replicate=false kvs.seed
	cp bud/kvsserver.dot "figures/kvs-network.dot"
	rm -rf bud
	bin/seed -dot -network=false -replicate=true kvs.seed
	cp bud/kvs.dot "figures/kvs-replicated.dot"
	rm -rf bud
	bin/seed -dot -network=true -replicate=true kvs.seed
	cp bud/kvsserver.dot "figures/kvs-network-replicated.dot"
	dot -T pdf -O figures/*.dot

.PHONY: clean
clean:
	-rm -rf bud bin pkg figures
