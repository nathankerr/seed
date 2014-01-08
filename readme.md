Seed is a declarative distributed systems language.

Its main goals are to be analyzable and transformable such that a service can be specified in a subset of the language and then transformations can be applied to implement various things such as dealing with network interfaces and data replication.

This language is experimental! It is not finished yet!

# Installation:

1. Install go: http://golang.org/doc/install
2. Install Seed
```
go get github.com/nathankerr/seed/seed
```

# Examples

The examples directory contains example transformations and services.

## kvs

The kvs directory contains a Key Value Store service.

## time

The time directory contains a time service which calculates the current time for a given timezone.

## add_network_interface.go

Adds a network interface to a service without one by changing input and output collections to channels and adding and handling the required network addresses through the service.

## add_replicated_tables.go

Changes table collections such that their contents are replicated between several of the same services running in a replica group.

# A note on types

Seed does not deal with types (this includes literals) because it couples the Service definition to a specific host environment and serialization mechanism. For example, to add a conditional like "table.column < 42" would require knowing what type table.column is, that it is comparable with whatever 42 is, and how to do the comparison. It might be fairly simple in this case, but soon becomes complicated with floating point, decimal, signed and unsigned numbers, etc. It gets worse with user defined types which then also need serialization code, etc.

The solution here is to leave typing to the host environment and out of Seed. If constants (such as the aforementioned 42) are needed, a table can be created to hold that single value and the actual value can be added at start up. The value can then be referenced as life_the_universe_and_everything.answer.

# Ideas on handling boolean predicates

operations to handle:
- =
- <
- >
- <=
- >=
- in (true if element in set)
- not (boolean negation)

operations have equal precedence, grouping with ()

## implementation idea

type CompareFn func(a Element, b Element) (int, error)
returns < 0 if a < b
returns 0 if a == b
returns > 0 if a > b
returns error if comparison not possible

register functions in a comparison map

comparisons map[string]map[string] CompareFn
each string is a qualified column, order does not matter (i.e., will check a,b and b,a for a function)

during startup, check if needed comparisons are registered. If not, list the one which are missing