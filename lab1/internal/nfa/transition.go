package nfa

import mapset "github.com/deckarep/golang-set/v2"

func (nfa *NFA) EpsilonClosure(states mapset.Set[string]) mapset.Set[string] {

	closure := mapset.Set[string](states)

	q := states.ToSlice()

	for len(q) > 0 {
		p := q[0]
		q = q[1:]

		epsTrans := nfa.Transitions[Pair{p, "ε"}]

		if epsTrans != nil {
			for _, nextState := range epsTrans.ToSlice() {
				if !closure.Contains(nextState) {
					closure.Add(nextState)
					q = append(q, nextState)
				}
			}
		}

	}

	return closure
}

func (nfa *NFA) Move(states mapset.Set[string], char string) mapset.Set[string] {

	nextStates := mapset.NewSet[string]()
	statesList := states.ToSlice()

	for _, state := range statesList {
		next := nfa.Transitions[Pair{state, char}]

		if next != nil {
			nextStatesList := next.ToSlice()
			for _, s := range nextStatesList {
				nextStates.Add(s)
			}
		}
	}
	return nextStates
}
