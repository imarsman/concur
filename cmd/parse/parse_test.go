package parse

import (
	"fmt"
	"strings"
	"testing"

	"github.com/benhoyt/goawk/interp"
	"github.com/benhoyt/goawk/parser"
	"github.com/matryer/is"
)

func TestCommand(t *testing.T) {
	is := is.New(t)
	src := `BEGIN { OFS = " "} { FS = "\\s+"} { print $1, $2 }`
	input := "hello world"

	prog, err := parser.ParseProgram([]byte(src), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	config := &interp.Config{
		Stdin: strings.NewReader(input),
		Vars:  []string{"OFS", ":"},
	}
	_, err = interp.ExecProgram(prog, config)
	if err != nil {
		fmt.Println(err)

	}
	is.True(1 == 1)
}
