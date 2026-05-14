package main

import (
	"flag"
	"fmt"

	"github.com/Lan-lann/go-compiler-lab/lab3/internal/grammar"
)

func main() {
	inputFile := flag.String("input", "input/grammar.txt", "指定文法文件路径")
	flag.Parse()

	g, err := grammar.LoadGrammar(*inputFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(g)

	fmt.Println(g.FindDeriveEplison())

	fmt.Println(g.GetNonTerminalFirstSet())

	fmt.Println(g.GetRightFirstSet())
}
