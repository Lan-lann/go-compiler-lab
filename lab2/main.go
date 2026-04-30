package main

import (
	"fmt"

	"github.com/Lan-lann/go-compiler-lab/lab2/internal/dfa"
	"github.com/Lan-lann/go-compiler-lab/lab2/internal/render"
)

func main() {
	dfa, err := dfa.LoadDFA("./input/dfa.txt")
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

	render.DrawGraphDFA(newDFA)
}
