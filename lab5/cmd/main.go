package main

import (
	"flag"
	"fmt"

	"github.com/Lan-lann/go-compiler-lab/lab5/internal/grammar"
)

func main() {
	inputFile := flag.String("inputGrammar", "input/g1.txt", "指定文法文件路径")
	inputString := flag.String("inputString", "bccd", "指定文法文件路径")

	flag.Parse()

	g, err := grammar.LoadGrammar(*inputFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(g)

	g.AugmentGrammar()

	fmt.Println(g)

	items := g.Items()

	for idx, x := range items {
		fmt.Println(idx+1, x.String())
	}

	// g.PrintItemSets()
	g.DrawItemSetsGraph("output/dfa.png")
	g.PrintParsingTable()

	g.PrintParseProcess(*inputString)
}
