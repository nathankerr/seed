.PHONY: all
all: seed
	-rm -rf bud
	bin/seed -dot -json -model kvs.seed

seed: src/*
	GOPATH=/Users/alaster/Projects/seed go install -a seed

print: *.go tmp version.tex
	./gen-print.sh
	context print
	acroread print.pdf

tmp:
	mkdir tmp

.PHONY: view-figures
view-figures: figures
	open "../figures/kvs.dot.pdf" "../figures/kvs-network.dot.pdf" "../figures/kvs-replicated.dot.pdf" "../figures/kvs-network-replicated.dot.pdf"

.PHONY: figures
figures: seed
	-rm -rf bud
	./seed -dot -network=false -replicate=false kvs.seed
	cp bud/kvs.dot "../figures/kvs.dot"
	rm -rf bud
	./seed -dot -network=true -replicate=false kvs.seed
	cp bud/kvsserver.dot "../figures/kvs-network.dot"
	rm -rf bud
	./seed -dot -network=false -replicate=true kvs.seed
	cp bud/kvs.dot "../figures/kvs-replicated.dot"
	rm -rf bud
	./seed -dot -network=true -replicate=true kvs.seed
	cp bud/kvsserver.dot "../figures/kvs-network-replicated.dot"
	dot -T pdf -O ../figures/*.dot

.PHONY: version.tex
version.tex:
	echo > version.tex
	git log -n1 --abbrev-commit --format=format:"%h %ai" >> version.tex
	if [ "`git diff --shortstat`" != '' ]; then echo " [WITH MODIFICATIONS]" >> version.tex; fi

.PHONY: clean
clean:
	-rm seed
	-rm -rf bud
	-rm -rf tmp
	context --purge
	-rm print.tex print.pdf 
