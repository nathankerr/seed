#!/bin/sh
OUT=print.tex

echo "\\environment layout

\usemodule[vim]
\\definevimtyping[Go][
  tab=4,
  directory=tmp,
  alternative=blackandwhite,
  style=\Code,
  strip=yes,
  lines=split,
  numbering=yes,
  before=\blank,
  after=\blank,
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
\\version[temporary]

\\starttext

\\title{Contents}
\\startcolumns[n=2,rule=on]
	\\placelist[chapter]
\\stopcolumns
" > $OUT

# echo '\chapter{readme.md}
# \processmarkdownfile{readme.md}' >> $OUT

for file in `ls *.go`
do
	echo "\\chapter{$file}" >> $OUT
	echo "\\typeGofile{$file}" >> $OUT
done

echo "\\stoptext" >> $OUT
