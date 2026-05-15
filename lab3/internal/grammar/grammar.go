package grammar

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

type Grammar struct {
	Start          string
	NonTerminalSet []string
	TerminalSet    []string
	Productions    map[string][]string
}

type DependEplison int

const (
	CanDerive DependEplison = 1 + iota
	NotCanDerive
)

func LoadGrammar(filename string) (*Grammar, error) {
	// 读取 文法 文件
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

func (g *Grammar) FindDeriveEplison() map[string]DependEplison {
	nonTerminalSetDeriveEplison := make(map[string]DependEplison)

	// 复制一份产生式集合
	productions := make(map[string][]string, len(g.Productions))
	for left, rights := range g.Productions {
		productions[left] = rights
	}

	keyLeft := make([]string, len(productions))
	for key := range g.Productions {
		keyLeft = append(keyLeft, key)
	}

	for _, left := range keyLeft {
		rights := productions[left]
		// 某一非终结符的某一产生式右部为ε ,则将数组中对应该非终结符的标志置为 "是 ",
		// 并 从 文 法 中 删 除 该 非 终 结 符 的 所 有 产 生 式 .
		if slices.Contains(rights, "ε") {
			nonTerminalSetDeriveEplison[left] = CanDerive
			delete(productions, left)
			continue
		}

		// 删 除 所 有 右 部 含 有 终 结 符 的 产 生 式 ,若 这 使 得 以 某 一 非 终 结 符 为 左 部 的 所 有 产 生 式 都 被 删 除 ,
		// 则 将 数 组 中 对 应 该 非 终 结 符 的 标 记 值 改 为 "否 ",说 明 该 非 终 结 符 不 能 推 出 6。
		// 遍历某一非终结符的所有产生式
		for idx := len(rights) - 1; idx >= 0; idx-- {
			right := rights[idx]
			// 判断某产生式是否含有终结符
			for _, t := range g.TerminalSet {
				// 有则删除
				if strings.Contains(right, t) {
					productions[left] = append(productions[left][:idx], productions[left][idx+1:]...)
					break
				}
			}
			if len(productions[left]) == 0 {
				nonTerminalSetDeriveEplison[left] = NotCanDerive
				delete(productions, left)
			}
		}
	}

	// 若 所 扫 描 到 的 非 终 结 符 在 数 组 中 对 应 的 标 志 是 “是 ",
	// 则 删 去 该 非 终 结 符 ;若 这 使 产 生 式 右 部 为 空 ,
	// 则 将 产 生 式 左 部 的 非 终 结 符 在 数 组 中 对 应 的 标 志 改 为 "是 ”,
	// 并删除以该非终 结符为左部的所有产生式。

	// 若 所 扫 描 到 的 非 终 结 符 号 在 数 组 中 对 应 的 标 志 是 “否 ”,
	// 则 删 去 该 产 生 式 ;若这使产 生 式 左 部 非 终 结 符 的 有 关 产 生 式 都 被 删 去 ,
	// 则 把 在 数 组 中 该 非 终 结 符 对 应 的 标 志 改 成 “否 "

	secondkeyLeft := make([]string, len(productions))
	for key := range productions {
		secondkeyLeft = append(secondkeyLeft, key)
	}
	for {
		for _, left := range secondkeyLeft {
			// 多个产生式
			rights := productions[left]

			for idx := len(rights) - 1; idx >= 0; idx-- {
				// 一个产生式
				right := []rune(rights[idx])
				for j := len(right) - 1; j >= 0; j-- {
					ch := string(right[j])
					if nonTerminalSetDeriveEplison[ch] == CanDerive {
						right = append(right[:j], right[j+1:]...)
					}

					if nonTerminalSetDeriveEplison[ch] == NotCanDerive {
						rights = append(rights[:idx], rights[idx+1:]...)
						break
					}
				}

				if len(right) == 0 {
					nonTerminalSetDeriveEplison[left] = CanDerive
					delete(productions, left)
					break
				}

				if len(rights) == 0 {
					nonTerminalSetDeriveEplison[left] = NotCanDerive
					delete(productions, left)
				}
			}

		}

		// 是否全部判断完毕
		flag := true
		for _, v := range nonTerminalSetDeriveEplison {
			if v == 0 {
				flag = false
			}
		}

		if flag {
			break
		}
	}

	return nonTerminalSetDeriveEplison
}

func (g *Grammar) GetNonTerminalFirstSet() map[string]mapset.Set[string] {

	nonTerminalSetDeriveEplison := g.FindDeriveEplison()

	firstSets := map[string]mapset.Set[string]{}
	for _, left := range g.NonTerminalSet {
		firstSets[left] = mapset.NewSet[string]()
	}
	for {
		// 依据扫描一遍是否有集合大小改变判断是否结束
		flag := true

		for left, rights := range g.Productions {
			initialSize := firstSets[left].Cardinality()
			for _, right := range rights {
				// 推导出ε
				if strings.Contains(right, "ε") {
					firstSets[left].Add("ε")
					continue
				}

				// 首字母为终结符
				if slices.Contains(g.TerminalSet, string(right[0])) {
					firstSets[left].Add(string(right[0]))
					continue
				}

				// 循环判断
				rt := []rune(right)
				countEmpty := 0
				for _, ch := range rt {
					// 为终结符
					if slices.Contains(g.TerminalSet, string(ch)) {
						firstSets[left].Add(string(ch))
						break
					}

					// 为非终结符
					if nonTerminalSetDeriveEplison[string(ch)] == CanDerive {
						countEmpty++
						addSet := firstSets[string(ch)].Clone()
						addSet.Remove("ε")
						firstSets[left].Append(addSet.ToSlice()...)

					} else if nonTerminalSetDeriveEplison[string(ch)] == NotCanDerive {
						addSet := firstSets[string(ch)].Clone()
						firstSets[left].Append(addSet.ToSlice()...)
						break
					}
				}

				if countEmpty == len(rt) {
					firstSets[left].Add("ε")
				}
			}

			// 如果发生改变，则本次遍历后还需继续循环
			if firstSets[left].Cardinality() > initialSize {
				flag = false
			}
		}

		// 如果遍历一遍后First 集合不再改变，则结束
		if flag {
			break
		}
	}

	return firstSets
}

func (g *Grammar) GetRightFirstSet() map[string]mapset.Set[string] {

	nonTerminalSetDeriveEplison := g.FindDeriveEplison()
	terminalFirstSets := g.GetNonTerminalFirstSet()

	firstSets := map[string]mapset.Set[string]{}
	for _, rights := range g.Productions {
		for _, right := range rights {
			firstSets[right] = mapset.NewSet[string]()
		}
	}

	for {
		// 依据扫描一遍是否有集合大小改变判断是否结束
		flag := true
		for _, rights := range g.Productions {
			for _, right := range rights {

				initialSize := firstSets[right].Cardinality()

				rt := []rune(right)
				countEmpty := 0

				for _, ch := range rt {

					// ε
					if string(ch) == "ε" {
						firstSets[right].Add(string(ch))
						break
					}

					// 为终结符
					if slices.Contains(g.TerminalSet, string(ch)) {
						firstSets[right].Add(string(ch))
						break
					}

					// 为非终结符
					if nonTerminalSetDeriveEplison[string(ch)] == CanDerive {
						countEmpty++
						addSet := terminalFirstSets[string(ch)].Clone()
						addSet.Remove("ε")
						firstSets[right].Append(addSet.ToSlice()...)

					} else if nonTerminalSetDeriveEplison[string(ch)] == NotCanDerive {
						addSet := terminalFirstSets[string(ch)].Clone()
						firstSets[right].Append(addSet.ToSlice()...)
						break
					}

				}

				if countEmpty == len(rt) {
					firstSets[right].Add("ε")
				}

				// 如果发生改变，则本次遍历后还需继续循环
				if firstSets[right].Cardinality() > initialSize {
					flag = false
				}
			}

		}

		// 如果遍历一遍后First 集合不再改变，则结束
		if flag {
			break
		}
	}
	return firstSets

}

func (g *Grammar) GetFollowSet() map[string]mapset.Set[string] {

	firstSets := g.GetNonTerminalFirstSet()
	followSets := map[string]mapset.Set[string]{}

	for _, left := range g.NonTerminalSet {
		followSets[left] = mapset.NewSet[string]()
	}

	followSets[g.Start].Add("#")

	for {
		// 依据扫描一遍是否有集合大小改变判断是否结束
		flag := true

		for left, rights := range g.Productions {

			for _, right := range rights {

				// 循环判断
				rt := []rune(right)
				l := len(rt)
				for i, ch := range rt {
					// 为非终结符
					if string(ch) != "ε" && !slices.Contains(g.TerminalSet, string(ch)) {
						pSize := followSets[string(ch)].Cardinality()
						// ch 不是最后一个字符
						if i != l-1 {
							bch := string(rt[i+1])

							if slices.Contains(g.TerminalSet, bch) {
								followSets[string(ch)].Add(bch)
							} else {
								addSet := firstSets[bch].Clone()
								addSet.Remove("ε")
								followSets[string(ch)].Append(addSet.ToSlice()...)
								if firstSets[bch].Contains("ε") {
									followSets[string(ch)].Append(followSets[left].ToSlice()...)
								}
							}
						} else {
							followSets[string(ch)].Append(followSets[left].ToSlice()...)
						}
						if followSets[string(ch)].Cardinality() > pSize {
							flag = false
						}
					}
				}
			}
		}
		// 如果遍历一遍后First 集合不再改变，则结束
		if flag {
			break
		}
	}

	return followSets
}

func (g *Grammar) GetSelectSet() map[string][]mapset.Set[string] {

	rightFirstSets := g.GetRightFirstSet()

	followSet := g.GetFollowSet()
	selectSets := map[string][]mapset.Set[string]{}
	// for _, left := range g.NonTerminalSet {
	// 	selectSets[left] = mapset.NewSet[string]()
	// }

	for left, rights := range g.Productions {
		for idx, right := range rights {
			// fmt.Println(left, right)
			selectSets[left] = append(selectSets[left], mapset.NewSet[string]())

			if rightFirstSets[right].Contains("ε") {
				addSet := rightFirstSets[right].Clone()
				addSet.Remove("ε")
				selectSets[left][idx].Append(addSet.ToSlice()...)
				selectSets[left][idx].Append(followSet[left].ToSlice()...)
			} else {
				addSet := rightFirstSets[right].Clone()
				addSet.Remove("ε")
				selectSets[left][idx].Append(addSet.ToSlice()...)
			}
		}
	}

	return selectSets
}

func (g *Grammar) IsLL1() bool {

	selectSets := g.GetSelectSet()
	for _, rights := range selectSets {
		for i, l := 0, len(rights); i < l; i++ {
			for j := i + 1; j < l; j++ {
				if rights[i].Intersect(rights[j]).Cardinality() > 0 {
					return false
				}
			}
		}
	}
	return true
}

func (g *Grammar) ShowSelectSet() {

	selectSets := g.GetSelectSet()
	for left, rights := range selectSets {
		for _, right := range rights {
			fmt.Printf("SELECT(%s) = {%s}\n", left, strings.Join(right.ToSlice(), ","))
		}
	}
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
