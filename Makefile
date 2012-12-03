first: figures

.PHONY: all
all: seed
	-rm -rf bud
	./seed -dot -network=true -replicate=true kvs.seed
	dot -O -T pdf bud/*.dot
	open bud/*.pdf

seed: *.go kvs.seed
	go build -o seed

print: *.go tmp version.tex
	./gen-print.sh
	context print
	acroread print.pdf

tmp:
	mkdir tmp

.PHONY: figures
figures: seed
	-rm -rf bud
	./seed -dot -network=false -replicate=false kvs.seed
	cp bud/kvs.dot "../figures/1 kvs.dot"
	rm -rf bud
	./seed -dot -network=true -replicate=false kvs.seed
	cp bud/kvsserver.dot "../figures/2 kvs-network.dot"
	rm -rf bud
	./seed -dot -network=false -replicate=true kvs.seed
	cp bud/kvs.dot "../figures/3 kvs-replicated.dot"
	rm -rf bud
	./seed -dot -network=true -replicate=true kvs.seed
	cp bud/kvsserver.dot "../figures/4 kvs-network-replicated.dot"
	dot -T pdf -O ../figures/*.dot
	open "../figures/1 kvs.dot.pdf" "../figures/2 kvs-network.dot.pdf" "../figures/3 kvs-replicated.dot.pdf" "../figures/4 kvs-network-replicated.dot.pdf"

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
