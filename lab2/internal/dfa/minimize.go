package dfa

import (
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
)

func (dfa *DFA) NewStatesSet() []mapset.Set[string] {
	nonFinalStateSet := mapset.NewSet[string]()
	finalStateSet := mapset.NewSet[string](dfa.FinalStates.ToSlice()...)

	for _, state := range dfa.States.ToSlice() {
		if !finalStateSet.Contains(state) {
			nonFinalStateSet.Add(state)
		}
	}

	StatesSetList := []mapset.Set[string]{nonFinalStateSet, finalStateSet}

	for {

		fmt.Println(StatesSetList)

		length := len(StatesSetList)

		for i, stateSet := range StatesSetList {
			// 是否进行了划分
			flag := false
			// 集合只有一个元素
			if stateSet.Cardinality() == 1 {
				continue
			}

			for _, ch := range dfa.Alphabet.ToSlice() {

				// 每次尝试均生成空转移
				NoneStatesSetList := make([]mapset.Set[string], length)
				for i := range NoneStatesSetList {
					NoneStatesSetList[i] = mapset.NewSet[string]()
				}

				// 对集合里每一个元素
				for _, state := range stateSet.ToSlice() {
					nextState := dfa.Transitions[Pair{state, ch}]
					// 划分子集
					for i, statesSet := range StatesSetList {
						if statesSet.Contains(nextState) {
							NoneStatesSetList[i].Add(state)
						}
					}
				}

				newStatesSetList := []mapset.Set[string]{}

				for _, stateset := range NoneStatesSetList {
					if stateset.Cardinality() > 0 {
						newStatesSetList = append(newStatesSetList, stateset)
					}
				}

				if len(newStatesSetList) >= 2 {
					StatesSetList = append(StatesSetList[:i], append(newStatesSetList, StatesSetList[i+1:]...)...)
					flag = true
					break
				}

			}
			if flag {
				break
			}
		}

		if len(StatesSetList) == length {
			break
		}

	}
	return StatesSetList
}

func (dfa *DFA) Minimize() *DFA {

	newDFA := &DFA{
		States:      mapset.NewSet[string](),
		Alphabet:    mapset.NewSet[string](),
		Transitions: map[Pair]string{},
		StartStates: "",
		FinalStates: mapset.NewSet[string](),
		StateMap:    map[string]string{},
	}

	newDFA.Alphabet = dfa.Alphabet

	statesSetList := dfa.NewStatesSet()

	stateList := []mapset.Set[string]{}

	//	找初态
	for _, set := range statesSetList {
		if set.Contains(dfa.StartStates) {
			setByName := getStateName(setToKey(set), &newDFA.StateMap)
			newDFA.StartStates = setByName
		}
	}

	//	找中间状态
	for _, set := range statesSetList {
		setByName := getStateName(setToKey(set), &newDFA.StateMap)
		newDFA.States.Add(setByName)
		stateList = append(stateList, set)
	}

	// 找终态
	for _, set := range statesSetList {
		for _, state := range dfa.FinalStates.ToSlice() {
			if set.Contains(state) {
				setByName := getStateName(setToKey(set), &newDFA.StateMap)
				newDFA.FinalStates.Add(setByName)
			}
		}
	}

	for input, toState := range dfa.Transitions {

		fromState, ch := input.First, input.Second
		var newFromState mapset.Set[string]
		var newToState mapset.Set[string]

		for _, state := range stateList {

			if state.Contains(fromState) {
				newFromState = state
			}

			if state.Contains(toState) {
				newToState = state
			}
		}

		newFromStateByName := getStateName(setToKey(newFromState), &newDFA.StateMap)
		newToStateByName := getStateName(setToKey(newToState), &newDFA.StateMap)

		newDFA.Transitions[Pair{newFromStateByName, ch}] = newToStateByName
	}

	return newDFA
}
