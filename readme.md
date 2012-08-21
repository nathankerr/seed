# Seed Prototype

Makes bud programs that can be scaled by distribution from seed programs

# Seed language

~~~
input <id> <schema>
output <id> <schema>
table <id> <schema>

<id> <+ <id>
<id> <- <id>
<id> <= <id>
<id> <+- <id>
~~~

To add:

- implicit map:
<id> <op> <id> {<...>}

- flat_map
<id> <op> <id>.flatmap do <...> end

- reduce, inject
<id> <op> <id>.reduce({}) do <...> end
<id> <op> <id>.inject({}) do <...> end

- include?
<id> <op> <id>.include?

- budcollection methods
<id> <op> <id>.schema
<id> <op> <id>.cols
<id> <op> <id>.key_cols
<id> <op> <id>.val_cols
<id> <op> <id>.keys
<id> <op> <id>.values
<id> <op> <id>.payloads // only defined for channels, therefore not needed in seed
<id> <op> <id>.inspected
<id> <op> <id>.exists?
<id> <op> <id>.notin(<id>, <...>)

- built in aggregates
min
max
choose
count
sum
avg
accum

- combinations
<id> <op> (<hash pairs>).pairs
<id> <op> (<hash pairs>).combos
<id> <op> (<hash pairs>).matches
<id> <op> (<hash pairs>).lefts
<id> <op> (<hash pairs>).rights
<id> <op> (<hash pairs>).outer
<id> <op> (<hash pairs>).flatten

- temp blocks (scratch that only exists in the bloom block)

# Refactor todo
- init rule.source in newRule