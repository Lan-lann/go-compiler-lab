package grammar

import (
	"bufio"
	"log"
	"os"
	"slices"
	"strings"
)

type Grammar struct {
	Start          string
	NonTerminalSet []string
	TerminalSet    []string
	Productions    map[string][]string
}

func LoadGrammar(filename string) (*Grammar, error) {
	// 读取 文法 文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
		log.Fatalf("读取文件失败: %v", err)
	}

	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	// 	处理每一行
	nonTerminalSetLine := lines[0]
	terminalSetLine := lines[1]

	nonTerminalSet := SplitAndTrim(nonTerminalSetLine, ",")
	terminalSet := SplitAndTrim(terminalSetLine, ",")
	start := nonTerminalSet[0]

	productions := map[string][]string{}

	for i := 2; i < len(lines); i++ {
		line := lines[i]
		parts := SplitAndTrim(line, "->")
		productions[parts[0]] = append(productions[parts[0]], parts[1])
	}

	return &Grammar{
		Start:          start,
		NonTerminalSet: nonTerminalSet,
		TerminalSet:    terminalSet,
		Productions:    productions,
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

func (g *Grammar) containsNonTerminal(sym string) bool {
	return slices.Contains(g.NonTerminalSet, sym)
}
