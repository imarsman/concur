# goparallel

A parallel workalike in Go.

The [parallel](https://en.wikipedia.org/wiki/GNU_parallel) command is a ridiculously complex and has command line
arguments that would require the creation of a separate parsing metchanism. Arguments such as "::::" do not work with Go
argument parsing.

What is likely to be implemented is substitution for incoming lists, numbered substitution and shuffling. The `--link`
option may or may not be implemented, although I have to understand it better before making a decision. Perl regular
expressions will not be implemented, though perl can be invoked in the command part along with things like sed and awk.

## Usage

### Arguments

Separate lists of arguments need to be quoted.

Simple sequences are supported

```sh
$ goparallel echo "Argument: {}" -a "{1..4}"
Argument: 1
Argument: 4
Argument: 2
Argument: 3
```

Argument lists can be specified separated by spaces

```sh
$ goparallel echo "Argument: {}" -a "1 2 3 4"
Argument: 1
Argument: 4
Argument: 2
Argument: 3
```

Argument lists can include literals and ranges

```sh
$ goparallel echo "Argument: {}" -a '1 2 3 4 5 {6..10}'
Argument: 7
Argument: 2
Argument: 6
Argument: 4
Argument: 5
Argument: 1
Argument: 3
Argument: 8
Argument: 10
Argument: 9
```

Shell calls can be made to create lists

```sh
goparallel echo "Argument: {1} {2}" -a "{0..9}" -a "$(echo {100..199})"
Argument: 1 100
Argument: 4 100
Argument: 5 100
Argument: 0 100
Argument: 6 100
Argument: 2 100
Argument: 7 100
Argument: 9 100
Argument: 3 100
Argument: 8 100
```