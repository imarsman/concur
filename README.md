# goparallel

A parallel workalike in Go.

The [parallel](https://en.wikipedia.org/wiki/GNU_parallel) command is a ridiculously complex and has command line
arguments that would require the creation of a separate parsing metchanism. Arguments such as "::::" do not work with Go
argument parsing.

What is likely to be implemented is substitution for incoming lists, numbered substitution and shuffling. The `--link`
option may or may not be implemented, although I have to understand it better before making a decision. Perl regular
expressions will not be implemented, though perl can be invoked in the command part along with things like sed and awk.

It is clear that the original parallel uses its own conventions. For example

```sh
$ ls -1 /var/log/*log | parallel echo "1 {1} 2 {2}"
1 /var/log/fsck_hfs.log 2
1 /var/log/fsck_apfs.log 2
1 /var/log/fsck_apfs_error.log 2
1 /var/log/install.log 2
1 /var/log/wifi.log 2
1 /var/log/acroUpdaterTools.log 2
1 /var/log/shutdown_monitor.log 2
1 /var/log/system.log 2
```

```sh
$ ls -1 /var/log/*log | goparallel echo "1 {1} 2 {2}" -a "a b c"
1 /var/log/acroUpdaterTools.log 2 a
1 /var/log/wifi.log 2 b
1 /var/log/fsck_apfs.log 2 b
1 /var/log/install.log 2 b
1 /var/log/fsck_apfs_error.log 2 c
1 /var/log/shutdown_monitor.log 2 c
1 /var/log/fsck_hfs.log 2 a
1 /var/log/system.log 2 a
```

## Usage

### Escaping command shell commands

The command specified can include calls that will be run by goparallel against an input. However, the command will bee
run prior to invocation unless escaped. Examples of characters and sequences that need to be escaped include "`" and "$(".

```sh
$ ls -1 /var/log/*log | goparallel "echo count \`wc -l {1}\`"
count 32 /var/log/fsck_apfs_error.log
count 432 /var/log/acroUpdaterTools.log
count 524 /var/log/system.log
count 395 /var/log/wifi.log
count 357 /var/log/fsck_hfs.log
count 39 /var/log/shutdown_monitor.log
count 817 /var/log/fsck_apfs.log
count 140367 /var/log/install.log
```

```sh
$ ls -1 /var/log/*log | goparallel "echo count \$(wc -l {1})"
count 32 /var/log/fsck_apfs_error.log
count 432 /var/log/acroUpdaterTools.log
count 524 /var/log/system.log
count 395 /var/log/wifi.log
count 357 /var/log/fsck_hfs.log
count 39 /var/log/shutdown_monitor.log
count 817 /var/log/fsck_apfs.log
count 140367 /var/log/install.log
```

**Note** that the same result can be obtained without escaping by using single quotes around the command.

### Arguments

Lists in arguments **need to be quoted**. Lists are split up separately.

The command to be run does not need to be quoted.

e.g. -a "{1..4}", -f "/var/log/*log", -a "1 2 3 4"

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

```sh
$ goparallel echo "{#} {1} {2}" -f "/var/log/*log" -a "$(echo {1..10..2})"
2 /var/log/fsck_apfs.log 3
1 /var/log/acroUpdaterTools.log 1
4 /var/log/fsck_hfs.log 7
8 /var/log/wifi.log 5
3 /var/log/fsck_apfs_error.log 5
7 /var/log/system.log 3
6 /var/log/shutdown_monitor.log 1
5 /var/log/install.log 9
```

```sh
ian@ian-macbookair ~/git/goparallel
$ seq 15 | goparallel echo "Slot {%} {1} {2}" -f "/var/log/*log" -s 2
Slot 1 2 /var/log/fsck_apfs.log
Slot 2 1 /var/log/acroUpdaterTools.log
Slot 1 4 /var/log/fsck_hfs.log
Slot 2 3 /var/log/fsck_apfs_error.log
Slot 1 6 /var/log/shutdown_monitor.log
Slot 2 5 /var/log/install.log
Slot 2 7 /var/log/system.log
Slot 1 8 /var/log/wifi.log
Slot 2 9 /var/log/acroUpdaterTools.log
Slot 1 10 /var/log/fsck_apfs.log
Slot 2 11 /var/log/fsck_apfs_error.log
Slot 1 12 /var/log/fsck_hfs.log
Slot 2 13 /var/log/install.log
Slot 1 14 /var/log/shutdown_monitor.log
```

## Benchmarks

Initial benchmarks are encouraging

```sh
$ time parallel echo "Argument: {}" ::: 1 2 3 4 5 {6..10}
Argument: 1
Argument: 4
Argument: 2
Argument: 6
Argument: 5
Argument: 3
Argument: 8
Argument: 7
Argument: 9
Argument: 10

parallel echo "Argument: {}" ::: 1 2 3 4 5 {6..10}  0.33s user 0.19s system 241% cpu 0.216 total
```

```sh
$ time goparallel echo "Argument: {}" -a '1 2 3 4 5 {6..10}'
Argument: 8
Argument: 1
Argument: 4
Argument: 5
Argument: 2
Argument: 6
Argument: 3
Argument: 7
Argument: 9
Argument: 10

goparallel echo "Argument: {}" -a '1 2 3 4 5 {6..10}'  0.02s user 0.03s system 176% cpu 0.027 total
```