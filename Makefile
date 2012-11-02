.PHONY: all
all: seed
	-rm -rf bud
	./seed kvs.seed

seed: *.go kvs.seed
	go build -o seed

print: *.go tmp version.tex
	./gen-print.sh
	context print
	acroread print.pdf

tmp:
	mkdir tmp

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
