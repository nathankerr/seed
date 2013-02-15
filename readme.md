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

Changes table collections such that their contents are replicated between several of the same services running in a replicant group.

