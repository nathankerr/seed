PWD=$(shell pwd)
GO=GOPATH=$(PWD) go
COMMANDS=seed
PACKAGES=$(COMMANDS) executor network replication service
PATH+=:$(PWD)/bin

.PHONY: all
all: install
	$(GO) test executor

none:
	# -rm -rf build
	# seed -t "service go" -transformations "network" -execute -timeout=5s -sleep=10ms kvs.seed
	# seed -t "bloom dot json service" -transformations "network" -execute kvs.seed
	# seed -t "bloom dot json service" -transformations "network" kvs.seed
	# seed -t "bloom json" -transformations "" cart.seed
	# seed -t "bloom json" -transformations "network" cart.seed

.PHONY: run
run: install
	seed -t "bloom dot json service go" -transformations "network" -execute -monitor="127.0.0.1:8000" -sleep=2s kvs.seed

.PHONY: install
install:
	$(GO) install $(PACKAGES)

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
figures: install
	-rm -rf figures
	mkdir figures
	-rm -rf build
	seed -t dot -transformations "" kvs.seed
	cp build/kvs.dot "figures/kvs.dot"
	rm -rf build
	seed -t dot -transformations "network" kvs.seed
	cp build/kvsserver.dot "figures/kvs-network.dot"
	rm -rf build
	seed -t dot -transformations "replicate" kvs.seed
	cp build/kvs.dot "figures/kvs-replicated.dot"
	rm -rf build
	seed -t dot -transformations "network replicate" kvs.seed
	cp build/kvsserver.dot "figures/kvs-network-replicated.dot"
	dot -T pdf -O figures/*.dot
