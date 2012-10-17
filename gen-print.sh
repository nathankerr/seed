#!/bin/sh
OUT=print.tex

echo "\\environment layout

\\usemodule[vim]
\\definevimtyping[Go][
        syntax=go,
        tab=4,
        directory=tmp,
        alternative=blackandwhite,
        style=\\Code,
        strip=yes,
        lines=split,
        numbering=yes,
]

\\defineexternalfilter[markdown][
	filtercommand={cat \\externalfilterinputfile | pandoc -t context -o \\externalfilteroutputfile},
	directory=tmp,
	cache=yes,
]

\\setuphead
   [chapter]
   [numberstyle=\\ChapterNumber,
    style=\\ChapterText,
    grid={low,-3pt},
    header=high,
    page=yes,
   ]

\\setupheadertexts[chapter]
\\setuptolerance[horizontal,strict]

\\definelayer[version][x=40mm,y=20.45mm]
\\setlayer[version][]{\\input{version.tex}}
\\setupbackgrounds[page][background=version]
\\\\version[temporary]

\\starttext

\\\\title{Contents}
\\startcolumns[n=2,rule=on]
	\\placelist[chapter]
\\stopcolumns
" > $OUT

echo '\\chapter{readme.md}
\\processmarkdownfile{readme.md}' >> $OUT

for file in `ls *.go`
do
	echo "\\\\chapter{$file}" >> $OUT
	echo "\\\\typeGofile{$file}\n" >> $OUT
done

echo "\\stoptext" >> $OUT
