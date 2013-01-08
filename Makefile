GO=GOPATH=/Users/alaster/Projects/seed go
COMMANDS=seed demo
PACKAGES=$(COMMANDS) executor network replication service

.PHONY: all
all: install
	# -rm -rf build
	bin/seed -t "service go" -transformations "network" kvs.seed
	# bin/seed -t "bloom dot json service" -transformations "network" -execute kvs.seed
	# bin/seed -t "bloom dot json service" -transformations "network" kvs.seed
	# bin/seed -t "bloom json" -transformations "" cart.seed
	# bin/seed -t "bloom json" -transformations "network" cart.seed

.PHONY: demo
demo: install
	bin/demo

.PHONY: run
run: install
	bin/seed -t "bloom dot json service" -transformations "network" -execute kvs.seed

.PHONY: install
install: test
	$(GO) install $(COMMANDS)

.PHONY: test
test: clean vet
	$(GO) test -i $(PACKAGES)
	$(GO) test $(PACKAGES)

.PHONY: format
format:
	$(GO) fmt $(PACKAGES)

.PHONY: vet
vet:
	$(GO) vet $(PACKAGES)

.PHONY: clean
clean:
	-rm -rf build bin pkg figures


types.dot.pdf: types.dot
	dot -O -T pdf types.dot
	open types.dot.pdf

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
