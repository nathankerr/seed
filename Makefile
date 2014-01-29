# NOTE: this makefile is only used for building seed.leg.go

all: seed.leg.go
	rm -f ${GOPATH}/bin/seed
	rm -rf build
	go install ./...
	go fmt ./...
	go vet ./...
	seed -t "go seed fieldgraph graph opennet" -transformations "networkg" services/kvs/kvs.seed
	cd build; dot -T pdf -O kvsserver.opennet.dot; open kvsserver.opennet.dot.pdf

seed.leg.go: seed.leg
	leg $< > $@

test:
	go test