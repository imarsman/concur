# goparallel

A parallel workalike in `golang`, though it is not really a parallel workalike. It is more like a line processing tool with
the option of shell execution and the application of an `awk` script. One benefit is the ability to run commands against
input in parallel, similar to the `parallel` and `xargs` utilities.

`parallel` excels at producing lists of text values that can be used to do many amazing things when they are integrated
into shell commands. The implementation of `goparallel` is more deterministic, with one predictable set of inputs for
each line processed. There is no real re-arranging of input lists beyond randomization.

`goparallel` involves lists that can be used for input and those lists can be used to produce text values that are
integrated into shell commands. `goparallel` is not as focussed on producing varied sets of values to be used in
commands. All lists in `goparallel` are cycled through with the longest list defining how many operations to perform. If
there is a shorter list and its members are fully used the list will cycle back to the starting point.

List of input using the `-a` flag (which can be used repeatedly to result in separate input lists) can be arbitrary
literal lists or expansions of file globbing pattters. For example `-a '/var/log/*log'` will result in a list of paths.
One can also supply lists using shell calls such as 

```sh
$ goparallel -a "$(seq 5)"
1
2
3
4
5
```

There is a simple sequence token that can be used as well

```sh
$ goparallel -a '{1..5}'
2
5
3
4
1
```

`goparallel` includes the ability to send the output of either the set of incoming list items or the command run to an
awk intepreter (using goawk).

Note that the order of output is normally the result of parallel excecution and as such is random. This can be overriden.

## Tokens

Tokens can be used in the command input. If command input is used the result must be a valid shell call. If no command
is supplied the result will be a list of the incoming list values.

### Tokens that can be used in the command

- `{} or {1}` - list 1 item
- `{.} or {1.}` - list 1 item without extension or same with numbered task list item
- `{/} or {1/}` - list 1 item basename of input line or same with numbered task list item
- `{//} or {1//}` - list 1 item dirname of output line or same with numbered task list item
- `{./} or {1./}` - list 1 item bsename of input line without extension or same with numbered task list item
- `{#}` sequence number of the job
- `{%}` job slot number (based on concurrency)
- `{1..10}` - a range - specify in `-a` and make sure to quote
  - sequences can be used too such as `seq 1 10` and `'$({1..10})'` (shell invocation)
  - multiple sequences can be used and for each `-a` will be added to a task list

I also have to test out and decide what to do with path and file oriented placeholders like {/} and {2/} where the
pattern is not a path or file. Currently the path and file oriented updates occur. It is up to the writer of the call to
be careful not to use path and file oriented tokens on non paths or non files.

### Examples

```sh
$ goparallel 'echo {}' -a '{0..9}'
7
0
5
4
2
6
1
3
9
8
```

will result in 10 invocations of the `echo` of the incoming list. If there is no
command the output will just be the input. For example

```sh
$ goparallel -a '{0..9}'
0
3
8
9
2
7
5
6
4
1
```

This will show the sequence numbers and items for a list


```sh
$ goparallel 'echo {#} {}' -a '{0..9}' -o
1 0
2 1
3 2
4 3
5 4
6 5
7 6
8 7
9 8
10 9
```

Note the use of the `-o` (ordered) flag.

See below for how to use more than one argument list and numbered tokens to produce output

```sh
$ goparallel 'echo {#} {1} {2}' -a '{0..9}' -a '{10..19}' -o
1 0 10
2 1 11
3 2 12
4 3 13
5 4 14
6 5 15
7 6 16
8 7 17
9 8 18
10 9 19
```

### awk scripts

`awk` scripts can be run on the output of the initial stage (either the provision of all input fields or the running of
a command).

```sh
$ goparallel -A '{FS="\\s+"; OFS=","} {print "got "$1}' -a '{1..10}'
got 2
got 8
got 10
got 1
got 3
got 6
got 4
got 7
got 5
got 9
```

```sh
ian@ian-macbookair ~/git/goparallel/cmd/command ‹main●›
$ cat test.txt | goparallel 'echo' -A 'BEGIN {FS="\\s+"; OFS=","} /red/ {print $1,$2,$3}'
strawberry,red,3
apple,red,4
```

Note that empty lines are by default skipped. That can be overriden with a flag

Here is the `test.text` file

```
name       color  amount
apple      red    4
banana     yellow 6
strawberry red    3
raspberry  red    99
grape      purple 10
apple      green  8
plum       purple 2
kiwi       brown  4
potato     brown  9
pineapple  yellow 5
```

```sh
$ cat test.txt | goparallel 'echo' -A 'BEGIN {FS="\\s+"; OFS=","} /red/ {print $1,$2,$3}' -E


raspberry,red,99
strawberry,red,3


apple,red,4




```

Here is an ordered version of the previous no blank lines and no filtering

```sh
$ cat test.txt | goparallel 'echo' -A 'BEGIN {FS="\\s+"; OFS=","} {print $1,$2,$3}' -o
name,color,amount
apple,red,4
banana,yellow,6
strawberry,red,3
raspberry,red,99
grape,purple,10
apple,green,8
plum,purple,2
kiwi,brown,4
potato,brown,9
pineapple,yellow,5
```

I will find out if this is a useful utility. There are some interesting uses, including the ability to accept the otuput
of `tail`

```sh
$ tail -f /var/log/*log|goparallel -A 'BEGIN {FS="\\s+"; OFS=","} /completed/ {print $0}' -o
/dev/rdisk3s3: fsck_apfs completed at Mon Mar  7 14:16:56 2022
/dev/rdisk3s3: fsck_apfs completed at Wed Mar 16 22:00:41 2022
fsck_apfs completed at Wed Mar 16 22:00:41 2022
/dev/rdisk4s2: fsck_hfs completed at Thu Mar 17 21:39:26 2022
/dev/rdisk4s2: fsck_hfs completed at Thu Mar 17 21:39:26 2022
```

```sh
tail -f /var/log/*log|goparallel -A 'BEGIN {FS="\\s+"; OFS=","} {print $1,$2,$3}'
==>,/var/log/acroUpdaterTools.log,<==
Jan,12,,2022
installer:,Upgrading,at
installer:,The,upgrade
Jan,12,,2022
Jan,12,,2022
Jan,12,,2022
Jan,12,,2022
Jan,12,,2022
Jan,12,,2022
==>,/var/log/fsck_apfs.log,<==
/dev/rdisk3s3:,fsck_apfs,started
/dev/rdisk3s3:,**,QUICKCHECK
/dev/rdisk3s3:,fsck_apfs,completed
/dev/rdisk3s3:,fsck_apfs,started
/dev/rdisk3s3:,**,QUICKCHECK
/dev/rdisk3s3:,fsck_apfs,completed
...
...
```

Here is an example of using both a standard input list and an additional list with awk

```sh
$ cat test/test.txt | goparallel 'echo {1} {2}' -o -a 'a b c' -A '{FS="\\s+"; OFS=" "} {print $1, $2, $3, $4}' -o
name color amount a
apple red 4 b
banana yellow 6 c
strawberry red 3 a
raspberry red 99 b
grape purple 10 c
apple green 8 a
plum purple 2 b
kiwi brown 4 c
potato brown 9 a
pineapple yellow 5 b
```

I will very likely find errors and bugs.

## Usage

```
$ goparallel -h
Usage: goparallel [--arguments ARGUMENTS] [--awk AWK] [--dry-run] [--slots SLOTS] 
                  [--shuffle] [--ordered] [--keep-order] [--print-empty] [COMMAND]

Positional arguments:
  COMMAND

Options:
  --arguments ARGUMENTS, -a ARGUMENTS
                         lists of arguments
  --awk AWK, -A AWK      process using supplied awk script
  --dry-run, -d          show command to run but don't run
  --slots SLOTS, -s SLOTS
                         number of parallel tasks [default: 8]
  --shuffle, -S          shuffle tasks prior to running
  --ordered, -o          run tasks in their incoming order
  --keep-order, -k       don't keep output for calls separate
  --print-empty, -E      print empty lines
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

## Trivia

In keeping with my recent trend when writing utilities, there are about 1,000 lines of `golang` code. I have moved
towards having a package contain about 400 lines of code with more allowed if the package is doing one thing such as
implementing handler functions. 1,000 lines of code to define data types and variables and functions to use all of that
is not as readable.

```sh
$ gocloc . --not-match-d vendor
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                               9            222            202           1024
Markdown                         1             81              0            381
YAML                             1              3              1             29
Plain Text                       2              0              0             23
-------------------------------------------------------------------------------
TOTAL                           13            306            203           1457
-------------------------------------------------------------------------------
```