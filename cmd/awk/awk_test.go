package awk

import (
	"testing"

	"github.com/matryer/is"
)

type awkCommand struct {
	command string
	payload string
}

func newAwkCommand(command, payload string) (cmd awkCommand) {
	cmd = awkCommand{}
	cmd.command = command
	cmd.payload = payload

	return
}

func getCommands() []awkCommand {
	command := `BEGIN { OFS = " "} { FS = "\\s+"} {ORS = ""} { printf "%s" $1; $1=$2=""; print $0 }`
	payload := "prefix hello world"

	commands := []awkCommand{}
	commands = append(commands, newAwkCommand(command, payload))
	commands = append(commands, newAwkCommand(command, "prefix godbye cruel world"))
	commands = append(commands, newAwkCommand(command, "prefix tomorrow and tomorrow and tomorrow"))

	return commands
}

func TestCommand(t *testing.T) {
	commands := getCommands()

	is := is.New(t)

	for _, c := range commands {
		awk, err := NewAwk(c.command)
		is.NoErr(err)
		output, err := awk.Execute(c.payload)
		is.NoErr(err)
		t.Log("payload:", c.payload)
		t.Logf("%s\n", output)
	}
}

// go test -bench=. -benchmem
func BenchmarkFib10(b *testing.B) {
	is := is.New(b)

	command := `BEGIN { OFS = " "} { FS = "\\s+"} { print $1, $2 }`
	payload := "hello world"

	var err error
	awk, err := NewAwk(command)
	is.NoErr(err)

	var output string
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		output, err = awk.Execute(payload)
	}

	b.Log(output)
}
