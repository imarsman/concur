# goparallel

A parallel workalike in Go.

The [parallel](https://en.wikipedia.org/wiki/GNU_parallel) command is a ridiculously complex and has command line
arguments that would require the creation of a separate parsing metchanism. Arguments such as "::::" do not work with Go
argument parsing.
