Originally a subset of bud with a little sugar. Now diverging to ease analysis.

# Seed language

~~~
input <id> <schema>
output <id> <schema>
table <id> <schema>

<id> <+ <expr>
<id> <- <expr>
<id> <+- <expr>
~~~

# EBNF

~~~
start := collection | rule
collection := keyword id schema
keyword := 'input' | 'output' | 'table'
id := [:letter:] ([:letter:] | '_')*
schema := array ('=>' array)?
array := '[' id (',' id)* ']'
rule := id operation expr
operation := '<+' | '<-' | '<+-'
expr := '[' qualifiedColumn (',' qualifiedColumn)* ']'
	(':' predicate (',' predicate)* )
predicate := qualifiedColumn '=>' qualifiedColumn
qualifiedColumn := id '.' id
~~~
