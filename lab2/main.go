package main

import (
	"flag"
	"fmt"

	"github.com/Lan-lann/go-compiler-lab/lab2/internal/dfa"
	"github.com/Lan-lann/go-compiler-lab/lab2/internal/render"
)

func main() {
	inputFile := flag.String("input", "input/dfa.txt", "指定NFA文件路径")
	flag.Parse()

	dfa, err := dfa.LoadDFA(*inputFile)

	if err != nil {
		panic(err)
	}

	res := dfa.NewStatesSet()

	fmt.Println(res)

	newDFA := dfa.Minimize()

	fmt.Println(newDFA.States.ToSlice())
	fmt.Println(newDFA.Alphabet.ToSlice())
	fmt.Println(newDFA.StartStates)
	fmt.Println(newDFA.FinalStates.ToSlice())
	for k, v := range newDFA.Transitions {
		fmt.Println(k, v)
	}

	render.DrawGraphDFA(dfa, "output/graph_original_dfa.png")
	render.DrawGraphDFA(newDFA, "output/graph_minimal_dfa.png")
}
