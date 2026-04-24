package dfa

import (
	"sort"
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

type Pair struct {
	First  string
	Second string
}

type DFA struct {
	States      mapset.Set[string]
	Alphabet    mapset.Set[string]
	Transitions map[Pair]string
	StartStates string
	FinalStates mapset.Set[string]
	StateMap    map[string]string
}

func setToKey(s mapset.Set[string]) string {
	slice := s.ToSlice()
	sort.Strings(slice)
	return strings.Join(slice, ",")
}

var nextName = 0

func getStateName(key string, stateMap *map[string]string) string {
	if name, ok := (*stateMap)[key]; ok {
		return name
	}

	name := strconv.Itoa(nextName)

	(*stateMap)[key] = name

	nextName++

	return name
}
