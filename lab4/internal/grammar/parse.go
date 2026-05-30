package grammar

import (
	"fmt"
	"slices"
	"strings"
)

type Pair struct {
	First  string
	Second string
}

func (g *Grammar) ConstructTable() map[Pair]string {

	selectSets := g.GetSelectSet()

	chList := append(g.TerminalSet, "#")
	predictTable := make(map[Pair]string)
	for _, nonTerminal := range g.NonTerminalSet {
		for _, ch := range chList {
			selectSet := selectSets[nonTerminal]
			for i, set := range selectSet {
				if set.Contains(ch) {
					predictTable[Pair{nonTerminal, ch}] = g.Productions[nonTerminal][i]
					break
				}
			}
		}
	}

	return predictTable
}

func (g *Grammar) Parse(s string) {

	table := g.ConstructTable()

	stack := []string{}
	stack = append(stack, "#", g.Start)

	idx := 0
	cnt := 1
	fmt.Printf("%-4s | %-12s | %-10s | %-20s\n", "步骤", "分析栈", "剩余输入串", "推导所用产生式或匹配")
	fmt.Println("-------|-----------------|-----------------|----------------------")

	for {
		t := stack[len(stack)-1]

		ch := string(s[idx])

		if slices.Contains(g.TerminalSet, t) {
			if string(ch) == t {
				idx++
				stack = stack[:len(stack)-1]
				fmt.Printf("%-6d | %-15v | %-15v | %-20s\n", cnt, strings.Join(stack, ""), s[idx:], string(ch)+"匹配")
			} else {
				fmt.Println("出错")
				return
			}

		} else {
			if t == "#" {
				if ch == "#" {
					fmt.Printf("%-6d | %-15v | %-15v | %-20s\n", cnt, strings.Join(stack, ""), s[idx:], "接受")
					return
				} else {
					fmt.Println("出错")
					return
				}
			}

			p, ok := table[Pair{t, ch}]

			if !ok {
				fmt.Println("出错")
				return
			}

			fmt.Printf("%-6d | %-15v | %-15v | %-20s\n", cnt, strings.Join(stack, ""), s[idx:], string(t)+"->"+p)
			stack = stack[:len(stack)-1]
			if string(p) != "ε" {
				for j := len(p) - 1; j >= 0; j-- {
					stack = append(stack, string(p[j]))
				}
			}

		}

		cnt++
	}

}
