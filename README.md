# goparallel

A parallel workalike in Go.

The [parallel](https://en.wikipedia.org/wiki/GNU_parallel) command is a ridiculously complex and has command line
arguments that would require the creation of a separate parsing metchanism. Arguments such as "::::" do not work with Go
argument parsing.

What is likely to be implemented is substitution for incoming lists, numbered substitution and shuffling. The `--link`
option may or may not be implemented, although I have to understand it better before making a decision. Perl regular
expressions will not be implemented, though perl can be invoked in the command part along with things like sed and awk.