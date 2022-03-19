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
	command := `BEGIN { OFS = " "} { FS = "\\s+"} {ORS = ""} { print $1; $1=$2=""; sub("  "," "); print $0 }`
	payload := "prefix hello world"

	commands := []awkCommand{}
	commands = append(commands, newAwkCommand(command, payload))
	commands = append(commands, newAwkCommand(command, "first goodbye cruel world"))
	commands = append(commands, newAwkCommand(command, "first tomorrow and tomorrow and tomorrow"))

	// cat /var/log/system.log|goawk 'BEGIN { OFS = " "} {printf "%s", $" "1$2" "$3; $1=$2=$3=""; sub("  ", " "); print $0}'

	return commands
}

func TestCommand(t *testing.T) {
	commands := getCommands()

	is := is.New(t)

	for _, c := range commands {
		// fmt.Println(c.command, c.payload)
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
