package awk

import (
	"testing"

	"github.com/matryer/is"
)

func TestCommand(t *testing.T) {
	is := is.New(t)
	command := `BEGIN { OFS = " "} { FS = "\\s+"} { print $1, $2 }`
	payload := "hello world"

	awk, err := NewAwk(command)
	is.NoErr(err)

	output, err := awk.Execute(payload)
	is.NoErr(err)
	t.Log(output)
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
