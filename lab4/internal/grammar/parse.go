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

// ConstructTable 根据文法的 SELECT 集构造预测分析表。
// 返回值为 map[Pair]string，键是 (非终结符, 输入符号) 对，值是对应的产生式。
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

// Parse 对输入串进行 LL(1) 预测分析，并打印每一步的分析过程。
func (g *Grammar) Parse(s string) {
	// 输入串必须以 '#' 结尾，否则无法进行分析。
	if s[len(s)-1] != '#' {
		fmt.Println("句子末尾不含#，请重新输入")
		return
	}

	table := g.ConstructTable()

	// 分析栈初始状态为 #S，其中 S 为文法开始符号。
	stack := []string{}
	stack = append(stack, "#", g.Start)

	idx := 0
	cnt := 1
	fmt.Printf("%-4s | %-12s | %-10s | %-20s\n", "步骤", "分析栈", "剩余输入串", "推导所用产生式或匹配")
	fmt.Println("-------|-----------------|-----------------|----------------------")

	for {
		// 当前栈顶符号
		t := stack[len(stack)-1]
		ch := string(s[idx])

		// 如果栈顶是终结符，则进行匹配操作。
		if slices.Contains(g.TerminalSet, t) {
			if string(ch) == t {
				fmt.Printf("%-6d | %-15v | %-15v | %-20s\n", cnt, strings.Join(stack, ""), s[idx:], string(ch)+"匹配")
				idx++
				stack = stack[:len(stack)-1]
			} else {
				fmt.Printf("%-6d | %-15v | %-15v | %-20s\n", cnt, strings.Join(stack, ""), s[idx:], "出错")
				return
			}

		} else {
			// 栈顶是非终结符或者结束符号
			if t == "#" {
				if ch == "#" {
					fmt.Printf("%-6d | %-15v | %-15v | %-20s\n", cnt, strings.Join(stack, ""), s[idx:], "接受")
					return
				} else {
					fmt.Printf("%-6d | %-15v | %-15v | %-20s\n", cnt, strings.Join(stack, ""), s[idx:], "出错")
					return
				}
			}

			// 查询预测分析表，获取当前非终结符和输入符号对应的产生式。
			p, ok := table[Pair{t, ch}]
			// 若为空，则表示出错
			if !ok {
				fmt.Printf("%-6d | %-15v | %-15v | %-20s\n", cnt, strings.Join(stack, ""), s[idx:], "出错")
				return
			}
			// 否则，选择对应产生式进行推导
			fmt.Printf("%-6d | %-15v | %-15v | %-20s\n", cnt, strings.Join(stack, ""), s[idx:], string(t)+"->"+p)
			stack = stack[:len(stack)-1]

			// 如果产生式不是 ε，则将右部符号逆序入栈。
			if string(p) != "ε" {
				for j := len(p) - 1; j >= 0; j-- {
					stack = append(stack, string(p[j]))
				}
			}
		}

		cnt++
	}
}
