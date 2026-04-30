package main

import (
	"flag"
	"fmt"

	"github.com/Lan-lann/go-compiler-lab/lab1/internal/dfa"
	"github.com/Lan-lann/go-compiler-lab/lab1/internal/nfa"
	"github.com/Lan-lann/go-compiler-lab/lab1/internal/render"
	mapset "github.com/deckarep/golang-set/v2"
)

func main() {

	inputFile := flag.String("input", "input/nfa.txt", "指定NFA文件路径")
	flag.Parse()

	nfa, err := nfa.LoadNFA(*inputFile)

	if err != nil {
		panic(err)
	}

	dfa := &dfa.DFA{
		States:      mapset.NewSet[string](),
		Alphabet:    mapset.NewSet[string](),
		Transitions: map[dfa.Pair]string{},
		StartStates: "",
		FinalStates: mapset.NewSet[string](),
		StateMap:    map[string]string{},
	}
	dfa.Subset(nfa)

	fmt.Println(dfa.States.ToSlice())
	fmt.Println(dfa.Alphabet.ToSlice())
	fmt.Println(dfa.StartStates)
	fmt.Println(dfa.FinalStates.ToSlice())
	for k, v := range dfa.Transitions {
		fmt.Println(k, v)
	}

	render.DrawGraphDFA(dfa)
	render.DrawGraphNFA(nfa)
}
