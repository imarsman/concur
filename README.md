# goparallel

A parallel workalike in Go. Actually, parallel is more similar to xargs as implemented.

I came across the parallel command very recently and thought writing a utility along its lines would be a good way to
explore command execution and goroutines and other concurrent tools such as mutexes and semaphors.

I am working on this code to look into calling executables in parallel using Golang. I am not sure I am invested in
implementing the combinatorial capabilities of the original parallel tool, although I can see the value of this in
scientific research (randomized assignment). There are some things which I may or may not implement, such as saving
output to a directory structure.

As currently implemented, goparallel gets lists for input from three distinct sources and in order of source; 

1. from standard input as a set of lines
2. from lists of arguments using the -a flag (you can use one or more of these and each will be a separate list)
3. from the -f flag, which reads lines of files to a set of lines. 
 
Basically, the lists are sources of input for commands run. I have kept the {1}, {2} notation from parallel as well as
the notation used for splitting paths and files into their components. I have not implemented the complex combanitorial
ordering logic for incoming lists. Lists are processed one item per command and the length of the output in terms of
commands run is defined by the length of the longest incoming list. Any list that reaches its end before the end of
command runs is looped back to the first item of that list. As possible useful additions emerge I may add them. Things
like tying one list's members to a previous one would I think require a different approach.

Given that I am not a researcher carrying out randomized experiments the likely focus for this code will be allowing the
parallel execution of shell commands and the most likely mode of use I think would be the use of a single incoming list.

## Similarities with parallel

I have implemented the placeholder tokenization such as {} and {1} and {#} along with path and file tokens such as {.}
and {/}.

I have implemented flags allowing output from commands to be printed as blocks per command or as it it is produced by
the execution of commands.

I have added a flag to cause tasks lists to be shuffled prior to execution.

I have added the ability to specify the concurrency to be used.

If there are no placeholders in the command to be run they will be added and then replaced by their list values.

```sh
$ goparallel 'echo list values ' -a '1 2' -a '{4..5}'
list values 1 4
list values 2 5
```

## Tokens

- `{} or {1}` - list 1 item
- `{.} or {1.}` - list 1 item without extension or same with list number
- `{/} or {1/}` - list 1 item basename of input line or same with list number
- `{//} or {1//}` - list 1 item dirname of output line or same with list number
- `{./} or {1./}` - list 1 item bsename of input line without extension or same with list number
- `{#}` sequence number of the job
- `{%}` job slot number (based on concurrency)
- `{1..10}` - a range - specify in `-a` and make sure to quote
  - sequences can be used too such as `seq 1 10` and `'$({1..10})'` (shell invocation)
  - multiple sequences can be used and for each `-a` will be added to a task list

I also have to test out and decide what to do with path and file oriented placeholders like {/} and {2/} where the
pattern is not a path or file. Currently the path and file oriented updates occur. There could be problems with this.
One problem is 

## Things to implement and work on

I need to ensure that I have done as good a job as possible to allow commands to be escaped. Commands passed to the Go
code for goparallel first are interpreted by the shell environment (zsh has been used for testing). Some characters such
as { and } can trigger the shell's parser, necessitating thigs like 'echo {}' instead of just echo {}.

It would be most reliable to avoid using path and filename oriented tokens if the incoming data is not relevant for
that. Results otherwise are unpredictable.

## Usage

```
$ goparallel -h
Usage: goparallel [--arguments ARGUMENTS] [--files FILES] [--dry-run] 
                  [--slots SLOTS] [--shuffle] [--ordered] [--keep-order] [COMMAND]

Positional arguments:
  COMMAND

Options:
  --arguments ARGUMENTS, -a ARGUMENTS
                         lists of arguments
  --files FILES, -f FILES
                         files to read into lines
  --dry-run, -d          show command to run but don't run
  --slots SLOTS, -s SLOTS
                         number of parallel tasks
  --shuffle, -S          shuffle tasks prior to running
  --ordered, -o          run tasks in their incoming order
  --keep-order, -k       don't keep output for calls separate
  --help, -h             display this help and exit
```

Ping some hosts and waith for full output from each before printing.

```sh
goparallel 'ping -c 1 "{}"' -a '127.0.0.1 ibm.com cisco.com'
64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.057 ms

--- 127.0.0.1 ping statistics ---
1 packets transmitted, 1 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 0.057/0.057/0.057/0.000 ms
PING cisco.com (72.163.4.185): 56 data bytes
64 bytes from 72.163.4.185: icmp_seq=0 ttl=239 time=78.342 ms

--- cisco.com ping statistics ---
1 packets transmitted, 1 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 78.342/78.342/78.342/0.000 ms
PING ibm.com (184.86.42.71): 56 data bytes
64 bytes from 184.86.42.71: icmp_seq=0 ttl=56 time=24.171 ms

--- ibm.com ping statistics ---
1 packets transmitted, 1 packets received, 0.0% packet loss
```

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

```sh
$ ls -1 /var/log/*log | goparallel 'echo count $(wc -l {1})'
count 0 /var/log/fsck_apfs_error.log
count 294 /var/log/system.log
count 432 /var/log/acroUpdaterTools.log
count 432 /var/log/acroUpdaterTools.log
count 294 /var/log/system.log
count 0 /var/log/fsck_apfs_error.log
count 153250 /var/log/install.log
count 153250 /var/log/install.log
```

### Arguments

Lists in arguments **need to be quoted**. Lists are split up separately.

The command to be run does not need to be quoted unless there are characters like { and `.

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

Initial benchmarks are encouraging, though parallel is written in Perl and does all kinds of cool things.

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