package seed

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

func ToLaTeX(seed *Seed, name string) ([]byte, error) {
	funcMap := template.FuncMap{
		"tex":       escapeTeX,
		"upper":     strings.ToUpper,
		"join":      joinText,
		"transient": transientCollections,
		"immediate": immediateRules,
		"deferred":  deferredRules,
		"operation": setOperation,
	}

	template, err := template.New("LaTeX").Funcs(funcMap).Delims("[[", "]]").Parse(`\documentclass[a4paper]{article}
\usepackage[T1]{fontenc}
\usepackage[utf8]{inputenc}
\usepackage{mathtools}
\usepackage{a4wide}

\title{[[.Name]]}
\date{}
\author{}

\begin{document}
\maketitle

\section{General}

A collection has a type, an ordered list of key columns, and an ordered list of data columns. The set of collection types is $\{ \text{input}, \text{output}, \text{table}, \text{scratch}, \text{channel} \}$. Column names must be unique within the collection. The information about a collection is expressed as a tuple:

$$\text{collection}_\text{collection\_name} = (type, key, data)$$

A collection is a set of tuples whose length is the sum of the lengths of that collection's key and data column lists. Elements of a tuple can be accessed using the corresponding column name. A collection is referred to by its name in all upper case:

$$\text{COLLECTION\_NAME}$$

Also, by convention:

$$\text{collection\_name} \in \text{COLLECTION\_NAME}$$

An element from $\text{collection\_name}$ can be referred to as:

$$\text{collection\_name}.\text{column\_name}$$

where $\text{column\_name}$ is in the collection's list of key or data columns.

\section{Collections}

[[range $name, $collection := .Seed.Collections]]
[[$name := tex $name]]
\begin{equation}
\text{collection}_\text{[[$name]]} = ( \text{[[.Type.String]]}, % Type
([[join .Key]]), % Key
([[join .Data]])% Data
)
\end{equation}
[[end]]

\section{Timestep}
\subsection{Phase 1: Clear Transient Collections}

[[$transient := transient .Seed.Collections]]
[[range $name, $collection := $transient]]
\begin{equation}
\text{[[upper $name|tex]]} = \emptyset
\end{equation}
[[end]]

\subsection{Phase 2: Execute Immediate Rules}

[[$immediate := immediate .Seed.Rules]]
[[range $immediate]]
\begin{equation}
[[.String | tex]]
\end{equation}
[[end]]

\subsection{Phase 3: Execute Deferred Rules}

[[$deferred := deferred .Seed.Rules]]
[[range $deferred]]
\begin{equation}
[[upper .Supplies | tex]] = [[upper .Supplies | tex]] \, [[operation .Operation]] \, \{ \text{[[.Projection]]} \mid \text{[[.Predicate]]} \}
\end{equation}
[[end]]

\section{Rules}
\begin{enumerate}
	[[range .Seed.Rules]]
	\item [[.String | tex]]
	[[end]]
\end{enumerate}

\end{document}
	`)
	if err != nil {
		return nil, err
	}

	type input struct {
		Seed *Seed
		Name string
	}

	buffer := new(bytes.Buffer)
	err = template.Execute(buffer, &input{Seed: seed, Name: name})
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func escapeTeX(text string) string {
	text = strings.Replace(text, "_", "\\_", -1)

	return text
}

func joinText(strs []string) string {
	for i, str := range strs {
		strs[i] = fmt.Sprintf("\\text{%s}", escapeTeX(str))
	}
	return strings.Join(strs, ", ")
}

func transientCollections(collections map[string]*Collection) map[string]*Collection {
	transient := make(map[string]*Collection)

	for name, collection := range collections {
		switch collection.Type {
		case CollectionInput, CollectionOutput, CollectionScratch, CollectionChannel:
			transient[name] = collection
		case CollectionTable:
			// not transient
		default:
			panic(fmt.Sprintf("unhandled collection type: %v", collection.Type))
		}
	}

	return transient
}

func immediateRules(rules []*Rule) []*Rule {
	immediate := []*Rule{}

	for _, rule := range rules {
		switch rule.Operation {
		case "<=":
			immediate = append(immediate, rule)
		case "<+", "<+-", "<-":
			// deferred rules
		default:
			panic(fmt.Sprintf("unhandled rule operation: %v", rule.Operation))
		}
	}

	return immediate
}

func deferredRules(rules []*Rule) []*Rule {
	deferred := []*Rule{}

	for _, rule := range rules {
		switch rule.Operation {
		case "<=":
			// immediate rules
		case "<+", "<+-", "<-":
			deferred = append(deferred, rule)
		default:
			panic(fmt.Sprintf("unhandled rule operation: %v", rule.Operation))
		}
	}

	return deferred
}

func setOperation(operation string) string {
	setop := ""
	switch operation {
	case "<+", "<=", "<~":
		setop = "\\cup"
	case "<-":
		setop = "\\setminus"
	case "<+-":
		setop = "FIXME <+-"
	default:
		panic(fmt.Sprintf("unhandled rule operation: %v", operation))
	}

	return setop
}
