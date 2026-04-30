package dfa

import (
	"bufio"
	"os"
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

func LoadDFA(filename string) (*DFA, error) {
	// 读取 dfa 文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	// 	处理每一行
	statesLine := lines[0]
	alphabetLine := lines[1]
	startStatesLine := lines[2]
	finalStatesLine := lines[3]

	states := mapset.NewSet(SplitAndTrim(statesLine, ",")...)
	alphabet := mapset.NewSet(SplitAndTrim(alphabetLine, ",")...)
	startStates := SplitAndTrim(startStatesLine, ",")[0]
	finalStates := mapset.NewSet(SplitAndTrim(finalStatesLine, ",")...)

	transitions := make(map[Pair]string)

	for i := 4; i < len(lines); i++ {

		line := lines[i]
		parts := SplitAndTrim(line, "=")
		// 函数输入
		left := parts[0]
		// 函数输出
		right := parts[1]

		leftParts := SplitAndTrim(left, ",")
		fromState := leftParts[0]
		char := leftParts[1]
		toStates := SplitAndTrim(right, ",")

		input := Pair{fromState, char}
		transitions[input] = toStates[0]
	}

	return &DFA{
		States:      states,
		Alphabet:    alphabet,
		Transitions: transitions,
		StartStates: startStates,
		FinalStates: finalStates,
	}, nil
}

func SplitAndTrim(s string, sep string) []string {
	// 分割字符串
	parts := strings.Split(s, sep)
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		// 去除空格
		t := strings.TrimSpace(p)
		if t != "" {
			res = append(res, t)
		}
	}
	return res
}
