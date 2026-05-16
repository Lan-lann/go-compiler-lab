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
				// 处理空右部或明确的 ε 标记
				if right == "" || right == "ε" || strings.TrimSpace(right) == "" {
					firstSets[left].Add("ε")
					continue
				}

				// 使用 rune 切片以支持多字节符号，先判断是否为空
				rt0 := []rune(right)
				if len(rt0) == 0 {
					firstSets[left].Add("ε")
					continue
				}

				// 首字母为终结符
				if slices.Contains(g.TerminalSet, string(rt0[0])) {
					firstSets[left].Add(string(rt0[0]))
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

				// 处理空右部或 ε
				if right == "" || right == "ε" || strings.TrimSpace(right) == "" {
					firstSets[right].Add("ε")
					if firstSets[right].Cardinality() > initialSize {
						flag = false
					}
					continue
				}

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
					chStr := string(ch)
					if chStr != "ε" && !slices.Contains(g.TerminalSet, chStr) {
						if _, ok := followSets[chStr]; !ok {
							followSets[chStr] = mapset.NewSet[string]()
						}
						pSize := followSets[chStr].Cardinality()
						// ch 不是最后一个字符
						if i != l-1 {
							bch := string(rt[i+1])

							if slices.Contains(g.TerminalSet, bch) {
								followSets[chStr].Add(bch)
							} else {
								if _, ok := firstSets[bch]; !ok {
									firstSets[bch] = mapset.NewSet[string]()
								}
								addSet := firstSets[bch].Clone()
								addSet.Remove("ε")
								followSets[chStr].Append(addSet.ToSlice()...)
								if firstSets[bch].Contains("ε") {
									followSets[chStr].Append(followSets[left].ToSlice()...)
								}
							}
						} else {
							followSets[chStr].Append(followSets[left].ToSlice()...)
						}
						if followSets[chStr].Cardinality() > pSize {
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

// FirstLetterSubstitution 将产生式右侧第一个字符为非终结符的进行代入替换
// 对每个非终结符 A（按顺序），对所有在 A 之前的非终结符 B：
// 找到所有 A -> B... 的产生式，用 B 的产生式右侧代替 B 进行替换
func (g *Grammar) FirstLetterSubstitution() bool {
	changed := false
	for i := 0; i < len(g.NonTerminalSet); i++ {
		it1 := g.NonTerminalSet[i]
		for j := 0; j < i; j++ {
			it2 := g.NonTerminalSet[j]
			var newP []string

			if rights, exists := g.Productions[it1]; exists {
				for _, right := range rights {
					// 检查右侧第一个字符是否为非终结符 it2
					if strings.HasPrefix(right, it2) {
						// 找到所有以 it2 为左部的产生式
						if it2Rights, exists2 := g.Productions[it2]; exists2 {
							for _, it2Right := range it2Rights {
								// 拼接：it2的产生式 + 原产生式去掉第一个非终结符后的部分
								if it2Right != "ε" {
									newRight := it2Right + right[len(it2):]
									newP = append(newP, newRight)
								} else {
									// 如果是 ε，则只保留原产生式去掉第一个非终结符后的部分
									newRight := right[len(it2):]
									newP = append(newP, newRight)
								}
							}
						}
					} else {
						// 保留不需要替换的产生式
						newP = append(newP, right)
					}
				}

				// 如果产生式集合发生变化，标记为已修改
				if !stringSlicesEqual(g.Productions[it1], newP) {
					changed = true
				}

				// 用新产生式替换旧产生式
				g.Productions[it1] = newP
			}
		}
	}

	return changed
}

// FirstLetterSubstitutionForCommonFactor 按产生式来处理代入替换，不按顺序
// 直到没有非终结符开头的产生式为止
func (g *Grammar) FirstLetterSubstitutionForCommonFactor() {

	for {
		localChanged := false

		for left, rights := range g.Productions {
			var newP []string

			for _, right := range rights {
				if right == "" || right == "ε" {
					newP = append(newP, right)
					continue
				}

				// 获取产生式首字符
				firstChar := string([]rune(right)[0])

				// 检查是否为非终结符
				if g.containsNonTerminal(firstChar) {
					// 找到该非终结符的所有产生式进行替换
					if ntRights, exists := g.Productions[firstChar]; exists {
						for _, ntRight := range ntRights {
							if ntRight != "ε" {
								// 若首字符为非终结符，则不替换
								if g.containsNonTerminal(string(ntRight[0])) {
									continue
								}

								newRight := ntRight + right[len(firstChar):]
								newP = append(newP, newRight)
							} else {
								newRight := right[len(firstChar):]
								if newRight == "" {
									newRight = "ε"
								}
								newP = append(newP, newRight)
							}
						}
						localChanged = true
					}
				} else {
					// 保留不需要替换的产生式
					newP = append(newP, right)
				}
			}

			// 更新产生式
			if localChanged {
				g.Productions[left] = newP
			}
		}

		// 如果本轮没有任何变化，则退出循环
		if !localChanged {
			break
		}
	}

}

func (g *Grammar) HaveCommonFactor() bool {
	// 使用当前文法副本进行判断，避免修改原文法
	g2 := g.clone()
	g2.FirstLetterSubstitutionForCommonFactor()

	for _, left := range g2.NonTerminalSet {
		rights, ok := g2.Productions[left]
		if !ok {
			continue
		}
		for i := 0; i < len(rights); i++ {
			if len(rights[i]) == 0 {
				continue
			}
			firstI := string([]rune(rights[i])[0])
			for j := i + 1; j < len(rights); j++ {
				if len(rights[j]) == 0 {
					continue
				}
				if firstI == string([]rune(rights[j])[0]) {
					return true
				}
			}
		}
	}
	return false
}

func (g *Grammar) HaveLeftRecursion() bool {
	g2 := g.clone()
	g2.FirstLetterSubstitution()

	for left, rights := range g2.Productions {
		for _, right := range rights {
			if right == "" || right == "ε" {
				continue
			}
			if strings.HasPrefix(right, left) {
				return true
			}
		}
	}
	return false
}

func (g *Grammar) DelUnreachableProduction() {
	reachable := mapset.NewSet[string]()
	stack := []string{g.Start}

	for len(stack) > 0 {
		ch := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if reachable.Contains(ch) {
			continue
		}
		reachable.Add(ch)
		for _, right := range g.Productions[ch] {
			for _, ch := range right {
				sym := string(ch)
				if g.containsNonTerminal(sym) && !reachable.Contains(sym) {
					stack = append(stack, sym)
				}
			}
		}
	}

	newNonTerminals := make([]string, 0, len(g.NonTerminalSet))
	for _, nt := range g.NonTerminalSet {
		if reachable.Contains(nt) {
			newNonTerminals = append(newNonTerminals, nt)
		}
	}
	g.NonTerminalSet = newNonTerminals

	for left := range g.Productions {
		if !reachable.Contains(left) {
			delete(g.Productions, left)
		}
	}
}

func (g *Grammar) ParsingCommonFactor() {
	for {
		changed := false
		for _, left := range g.NonTerminalSet {
			rights, ok := g.Productions[left]
			if !ok || len(rights) < 2 {
				continue
			}

			maxLen := 0
			for _, right := range rights {
				if l := len([]rune(right)); l > maxLen {
					maxLen = l
				}
			}

			for length := maxLen; length >= 1 && !changed; length-- {
				prefixGroups := map[string][]int{}
				for i, right := range rights {
					rr := []rune(right)
					if len(rr) < length {
						continue
					}
					prefix := string(rr[:length])
					prefixGroups[prefix] = append(prefixGroups[prefix], i)
				}

				for prefix, indices := range prefixGroups {
					if len(indices) <= 1 {
						continue
					}

					newNT := g.getNewNonTerminal()
					g.NonTerminalSet = append(g.NonTerminalSet, newNT)

					newNTProductions := make([]string, 0, len(indices))
					remainingRights := make([]string, 0, len(rights)-len(indices))
					remove := map[int]struct{}{}
					for _, idx := range indices {
						remove[idx] = struct{}{}
					}

					for idx, right := range rights {
						if _, skip := remove[idx]; skip {
							rr := []rune(right)
							if len(rr) == length {
								newNTProductions = append(newNTProductions, "ε")
							} else {
								newNTProductions = append(newNTProductions, string(rr[length:]))
							}
						} else {
							remainingRights = append(remainingRights, right)
						}
					}

					g.Productions[left] = append(remainingRights, prefix+newNT)
					g.Productions[newNT] = append(g.Productions[newNT], newNTProductions...)
					changed = true
					break
				}
			}

			if changed {
				break
			}
		}

		if !changed {
			break
		}
	}
}

func (g *Grammar) ParsingDirectLeftRecursion() {
	for {
		changed := false
		for _, left := range g.NonTerminalSet {
			rights, ok := g.Productions[left]
			if !ok || len(rights) == 0 {
				continue
			}

			recurIndices := make([]int, 0)
			endIndices := make([]int, 0)
			for i, right := range rights {
				if right == "" || right == "ε" {
					endIndices = append(endIndices, i)
					continue
				}
				if strings.HasPrefix(right, left) {
					recurIndices = append(recurIndices, i)
				} else {
					endIndices = append(endIndices, i)
				}
			}

			if len(recurIndices) == 0 {
				continue
			}

			newNT := g.getNewNonTerminal()
			g.NonTerminalSet = append(g.NonTerminalSet, newNT)

			newNTProductions := make([]string, 0, len(recurIndices)+1)
			for _, idx := range recurIndices {
				right := rights[idx]
				rr := []rune(right)
				suffix := ""
				if len(rr) > len([]rune(left)) {
					suffix = string(rr[len([]rune(left)):])
				}
				newNTProductions = append(newNTProductions, suffix+newNT)
			}
			newNTProductions = append(newNTProductions, "ε")

			newLeftProductions := make([]string, 0, len(endIndices))
			for _, idx := range endIndices {
				right := rights[idx]
				if right == "ε" {
					newLeftProductions = append(newLeftProductions, newNT)
				} else {
					newLeftProductions = append(newLeftProductions, right+newNT)
				}
			}
			if len(newLeftProductions) == 0 {
				newLeftProductions = append(newLeftProductions, newNT)
			}

			g.Productions[left] = newLeftProductions
			g.Productions[newNT] = append(g.Productions[newNT], newNTProductions...)
			changed = true
			break
		}

		if !changed {
			break
		}
	}
}

func (g *Grammar) clone() *Grammar {
	copyG := &Grammar{
		Start:          g.Start,
		NonTerminalSet: append([]string(nil), g.NonTerminalSet...),
		TerminalSet:    append([]string(nil), g.TerminalSet...),
		Productions:    map[string][]string{},
	}
	for left, rights := range g.Productions {
		copyG.Productions[left] = append([]string(nil), rights...)
	}
	return copyG
}

func (g *Grammar) containsNonTerminal(sym string) bool {
	return slices.Contains(g.NonTerminalSet, sym)
}

func (g *Grammar) getNewNonTerminal() string {
	for c := 'Z'; c >= 'A'; c-- {
		sym := string(c)
		if !g.containsNonTerminal(sym) {
			return sym
		}
	}

	suffix := 1
	for {
		for c := 'Z'; c >= 'A'; c-- {
			sym := fmt.Sprintf("%s%d", string(c), suffix)
			if !g.containsNonTerminal(sym) {
				return sym
			}
		}
		suffix++
	}
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
