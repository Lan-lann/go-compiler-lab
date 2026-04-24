package dfa

import (
	"github.com/Lan-lann/go-compiler-lab/lab1/internal/nfa"
	mapset "github.com/deckarep/golang-set/v2"
)

func (dfa *DFA) Subset(nfa *nfa.NFA) {
	dfa.Alphabet = nfa.Alphabet

	startStates := nfa.EpsilonClosure(nfa.StartStates)

	dfa.StartStates = getStateName(setToKey(startStates), &dfa.StateMap)
	dfa.States.Add(dfa.StartStates)

	if s := startStates.Intersect(nfa.FinalStates); s.Cardinality() != 0 {
		dfa.FinalStates.Add(dfa.StartStates)
	}

	unmarkedStates := []mapset.Set[string]{startStates}

	for len(unmarkedStates) > 0 {
		currentState := unmarkedStates[0]
		currentStateByName := getStateName(setToKey(currentState), &dfa.StateMap)

		unmarkedStates = unmarkedStates[1:]

		for _, char := range nfa.Alphabet.ToSlice() {
			nextStates := nfa.EpsilonClosure(nfa.Move(currentState, char))

			// 如果nextSatates 为空，则跳过!!
			if nextStates.Cardinality() == 0 {
				continue
			}

			// 判断DFA 状态集合中是否已经有该状态

			nextStatesByName := getStateName(setToKey(nextStates), &dfa.StateMap)
			if !dfa.States.Contains(nextStatesByName) {

				dfa.States.Add(nextStatesByName)

				unmarkedStates = append(unmarkedStates, nextStates)

				if s := nextStates.Intersect(nfa.FinalStates); s.Cardinality() != 0 {
					dfa.FinalStates.Add(nextStatesByName)
				}
			}
			dfa.Transitions[Pair{currentStateByName, char}] = nextStatesByName
		}
	}

}
