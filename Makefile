.PHONY: all
all: bin/seed
	-rm -rf build
	bin/seed -t "bloom dot json service" kvs.seed

types.dot.pdf: types.dot
	dot -O -T pdf types.dot
	open types.dot.pdf

.PHONY: bin/seed
bin/seed: src/*
	GOPATH=/Users/alaster/Projects/seed go install  seed

.PHONY: view-figures
view-figures: figures
	open "figures/kvs.dot.pdf" "figures/kvs-network.dot.pdf" "figures/kvs-replicated.dot.pdf" "figures/kvs-network-replicated.dot.pdf"

.PHONY: figures
figures: bin/seed
	-rm -rf figures
	mkdir figures
	-rm -rf build
	bin/seed -t dot -transformations "" kvs.seed
	cp build/kvs.dot "figures/kvs.dot"
	rm -rf build
	bin/seed -t dot -transformations "network" kvs.seed
	cp build/kvsserver.dot "figures/kvs-network.dot"
	rm -rf build
	bin/seed -t dot -transformations "replicate" kvs.seed
	cp build/kvs.dot "figures/kvs-replicated.dot"
	rm -rf build
	bin/seed -t dot -transformations "network replicate" kvs.seed
	cp build/kvsserver.dot "figures/kvs-network-replicated.dot"
	dot -T pdf -O figures/*.dot

.PHONY: format
format:
	go fmt seed service network service

.PHONY: clean
clean:
	-rm -rf build bin pkg figures
