package main

import (
	"github.com/alexflint/go-arg"
)

// args CLI args
type args struct {
}

func main() {
	var callArgs args // initialize call args structure
	arg.MustParse(&callArgs)

}
