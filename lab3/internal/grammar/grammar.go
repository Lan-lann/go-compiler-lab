package grammar

import (
	"bufio"
	"fmt"
	"log"
	"maps"
	"os"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

// 文法数据结构定义
type Grammar struct {
	Start          string
	NonTerminalSet []string
	TerminalSet    []string
	Productions    map[string][]string
}

// 定义常量代表是否可以推导出 ε
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

func (g *Grammar) FindDeriveEplison() map[string]DependEplison {

	// 初始化非终结符能否推导出 ε 的判断数组（这里用 map 实现）
	nonTerminalSetDeriveEplison := make(map[string]DependEplison)

	// 复制一份产生式集合， 避免后续删除操作影响原产生式集合
	productions := make(map[string][]string, len(g.Productions))
	maps.Copy(productions, g.Productions)

	// 遍历所有非终结符
	for _, left := range g.NonTerminalSet {
		// 取出该终结符的所有产生式
		rights := productions[left]

		// 1. 若某一非终结符的某一产生式右部为ε ,则将数组中对应该非终结符的标志置为 "是 ",
		// 并从文法中删除该非终结符的所有产生式.
		if slices.Contains(rights, "ε") {
			nonTerminalSetDeriveEplison[left] = CanDerive
			delete(productions, left)
			continue
		}

		// 2. 删除所有右部含有终结符的产生式 ,若这使得以某一非终结符为左部的所有产生式都被删除 ,
		// 则将数组中对应该非终结符的标记值改为"否",说明该非终结符不能推出 ε。

		// 遍历某一非终结符的产生式(采用倒序遍历，避免产生式前移导致直接跳过下一产生式)
		for idx := len(rights) - 1; idx >= 0; idx-- {
			// 取出该产生式
			right := rights[idx]

			// 判断该产生式是否含有终结符
			for _, t := range g.TerminalSet {
				// 若含有终结符，则删除
				if strings.Contains(right, t) {
					productions[left] = append(productions[left][:idx], productions[left][idx+1:]...)
					break
				}
			}

			// 若该非终结符的所有产生式都被删除，则将数组中对应该非终结符的标记值改为"否
			if len(productions[left]) == 0 {
				nonTerminalSetDeriveEplison[left] = NotCanDerive
				// 从产生式集合中删除该映射, 例如: A:[]为空，删除 A
				delete(productions, left)
			}
		}
	}

	// 若所扫描到的非终结符在数组中对应的标志是 “是 ",
	// 则删去该非终结符;若这使产生式右部为空 ,
	// 则将产生式左部的非终结符在数组中对应的标志改为"是 ”,
	// 并删除以该非终结符为左部的所有产生式。

	// 若所扫描到的非终结符号在数组中对应的标志是“否”,
	// 则删去该产生式;若这使产生式左部非终结符的有关产生式都被删去 ,
	// 则把在数组中该非终结符对应的标志改成“否"

	for {

		for _, left := range g.NonTerminalSet {
			// 获得一个非终结符的多个产生式
			rights := productions[left]

			for idx := len(rights) - 1; idx >= 0; idx-- {
				// 获取一个产生式
				right := []rune(rights[idx])
				// 依次遍历右部
				for j := len(right) - 1; j >= 0; j-- {
					ch := string(right[j])
					// 若能推导出 ε，则删除该非终结符
					if nonTerminalSetDeriveEplison[ch] == CanDerive {
						right = append(right[:j], right[j+1:]...)
					}
					// 若不能推导出 ε，则删除该产生式
					if nonTerminalSetDeriveEplison[ch] == NotCanDerive {
						rights = append(rights[:idx], rights[idx+1:]...)
						break
					}
				}

				// 若该产生式右部都被删除，则表示能推导出 ε，对应非终结符标记为“是“， 删除该非终结符所有产生式
				if len(right) == 0 {
					nonTerminalSetDeriveEplison[left] = CanDerive
					delete(productions, left)
					break
				}
			}

			// 若该非终结符的所有产生式均被删除，则无表示法推导出 ε， 对应非终结符标记为“否“
			if len(rights) == 0 {
				nonTerminalSetDeriveEplison[left] = NotCanDerive
				delete(productions, left)
			}

		}

		// 是否全部判断完毕
		flag := true
		for _, v := range nonTerminalSetDeriveEplison {
			if v == 0 {
				flag = false
			}
		}

		// 全部判断完毕，则退出
		if flag {
			break
		}
	}

	return nonTerminalSetDeriveEplison
}

func (g *Grammar) GetNonTerminalFirstSet() map[string]mapset.Set[string] {
	// 获得非终结符是否能推导出 ε
	nonTerminalSetDeriveEplison := g.FindDeriveEplison()

	// 初始化空的 First 集合
	firstSets := map[string]mapset.Set[string]{}
	for _, left := range g.NonTerminalSet {
		firstSets[left] = mapset.NewSet[string]()
	}

	for {
		// 依据扫描一遍是否有集合大小改变判断是否结束
		flag := true

		for left, rights := range g.Productions {

			// 记录 First 集合初始大小
			initialSize := firstSets[left].Cardinality()

			for _, right := range rights {

				rt := []rune(right)

				// 记录右部是否均能推导出 ε
				countEmpty := 0

				// 遍历每一字符
				for _, ch := range rt {
					// 若首字母为终结符， 则加入 First集合
					if slices.Contains(g.TerminalSet, string(ch)) {
						firstSets[left].Add(string(ch))
						break
					}

					// 若为非终结符，若能推出空，则将 First(ch) - ε 加入
					if nonTerminalSetDeriveEplison[string(ch)] == CanDerive {
						countEmpty++
						addSet := firstSets[string(ch)].Clone()
						addSet.Remove("ε")
						firstSets[left].Append(addSet.ToSlice()...)

					} else if nonTerminalSetDeriveEplison[string(ch)] == NotCanDerive { // 若不能推出空，则将 First(ch) 加入， 停止遍历
						addSet := firstSets[string(ch)].Clone()
						firstSets[left].Append(addSet.ToSlice()...)
						break
					}
				}

				// 若均能推导出 ε， 则产生式右部可以推导出 ε， 将 ε 加入 First 集合
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
	// 获得非终结符是否能推导出 ε
	nonTerminalSetDeriveEplison := g.FindDeriveEplison()
	// 获得非终结符的 First 集合
	terminalFirstSets := g.GetNonTerminalFirstSet()

	// 初始化产生式右部的 First 集合(初始为空)
	firstSets := map[string]mapset.Set[string]{}
	for _, rights := range g.Productions {
		for _, right := range rights {
			firstSets[right] = mapset.NewSet[string]()
		}
	}

	for {
		// 依据扫描一遍是否有集合大小改变判断是否结束
		flag := true
		// 遍历所有产生式
		for _, rights := range g.Productions {
			for _, right := range rights {
				// 记录初始集合大小
				initialSize := firstSets[right].Cardinality()

				rt := []rune(right)
				countEmpty := 0

				// 遍历一条产生式右部的所有字符
				for _, ch := range rt {
					// 若为 ε，添加 ε， 退出
					if string(ch) == "ε" {
						firstSets[right].Add(string(ch))
						break
					}

					// 若为终结符， 添加该终结符，退出
					if slices.Contains(g.TerminalSet, string(ch)) {
						firstSets[right].Add(string(ch))
						break
					}

					// 若为非终结符，如果能推导出 ε，则将 First(ch) - ε 加入，继续遍历下一个字符
					if nonTerminalSetDeriveEplison[string(ch)] == CanDerive {
						// 记录能推导出ε的字符数
						countEmpty++
						addSet := terminalFirstSets[string(ch)].Clone()
						addSet.Remove("ε")
						firstSets[right].Append(addSet.ToSlice()...)

					} else if nonTerminalSetDeriveEplison[string(ch)] == NotCanDerive { // 如果不能推导出 ε， 则将 First(ch)加入，停止遍历
						addSet := terminalFirstSets[string(ch)].Clone()
						firstSets[right].Append(addSet.ToSlice()...)
						break
					}

				}
				// 如果产生式右部字符全部可以推导出 ε， 即该右部可以推导出 ε ，则加入 ε
				if countEmpty == len(rt) {
					firstSets[right].Add("ε")
				}

				// 如果有First 集合发生改变，则本次遍历后还需继续循环
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
				// 遍历每个产生式的每一个字符
				for i, ch := range rt {
					chStr := string(ch)
					// 为非终结符
					if slices.Contains(g.NonTerminalSet, chStr) {

						pSize := followSets[chStr].Cardinality()
						// ch 是最后一个字符
						if i == l-1 {
							followSets[chStr].Append(followSets[left].ToSlice()...)
						} else { // ch 不是最后一个字符
							canAllDeriveEmpty := true
							for j := i + 1; j < l; j++ {
								next := string(rt[j])

								if next == "ε" {
									continue
								}

								// 若后继为终结符，则不能推出ε，将该字符加入 Follow 集合
								if slices.Contains(g.TerminalSet, next) {
									followSets[chStr].Add(next)
									canAllDeriveEmpty = false
									break
								}

								// 若后继为非终结符，将 First(next) - ε 加入 Follow 集合
								addSet := firstSets[next].Clone()
								addSet.Remove("ε")
								followSets[chStr].Append(addSet.ToSlice()...)
								// 若不能推出ε，则结束
								if !firstSets[next].Contains("ε") {
									canAllDeriveEmpty = false
									break
								}
							}

							// 若后继能推导出 ε, 则将 Follow(left) 加入 Follow 集合
							if canAllDeriveEmpty {
								followSets[chStr].Append(followSets[left].ToSlice()...)
							}
						}

						// 集合大小发生变化，记录Flag
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
	// 获取产生式右部的 First 集合
	rightFirstSets := g.GetRightFirstSet()
	// 获取非终结符的 Follow 集合
	followSet := g.GetFollowSet()
	selectSets := map[string][]mapset.Set[string]{}

	for left, rights := range g.Productions {
		for idx, right := range rights {

			selectSets[left] = append(selectSets[left], mapset.NewSet[string]())
			// 若右部能推导出ε，则 First 集合既含有右部的 First 集合（除去ε）也包含左部的 Follow 集合
			if rightFirstSets[right].Contains("ε") {
				addSet := rightFirstSets[right].Clone()
				addSet.Remove("ε")
				selectSets[left][idx].Append(addSet.ToSlice()...)
				selectSets[left][idx].Append(followSet[left].ToSlice()...)
			} else { // 若右部不能推导出ε，则 First 集合只含有右部的 First 集合（除去ε）
				addSet := rightFirstSets[right].Clone()
				addSet.Remove("ε")
				selectSets[left][idx].Append(addSet.ToSlice()...)
			}
		}
	}

	return selectSets
}

func (g *Grammar) IsLL1() bool {
	// 获取 SELECT 集合
	selectSets := g.GetSelectSet()

	// 遍历判断同一非终结符的所有产生式两两是否相交
	for _, rights := range selectSets {
		for i, l := 0, len(rights); i < l; i++ {
			for j := i + 1; j < l; j++ {
				// 若相交则不是 LL1 文法
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

// FirstLetterSubstitution 将产生式右侧第一个字符为非终结符的进行代入替换， 同时返回是否有产生式发生改变
// 对每个非终结符 A（按顺序），对所有在 A 之前的非终结符 B：
// 找到所有 A -> B... 的产生式，用 B 的产生式右部进行替换
func (g *Grammar) FirstLetterSubstitution() (isChanged bool) {

	// 遍历非终结符集
	for i := 0; i < len(g.NonTerminalSet); i++ {
		it1 := g.NonTerminalSet[i]
		// 取当前非终结符集前的终结符
		for j := 0; j < i; j++ {
			it2 := g.NonTerminalSet[j]
			// 初始化存放新的产生式
			var newP []string

			// 取出t1 的所有产生式
			if rights, exists := g.Productions[it1]; exists {
				for _, right := range rights {

					// 检查右侧第一个字符是否为非终结符 it2
					if strings.HasPrefix(right, it2) {
						// 找到所有以 it2 为左部的产生式
						if it2Rights, exists2 := g.Productions[it2]; exists2 {
							for _, it2Right := range it2Rights {
								// 如果不为ε， 拼接：it2的产生式右部 + 原产生式去掉第一个非终结符后的部分
								if it2Right != "ε" {
									newRight := it2Right + right[len(it2):]
									newP = append(newP, newRight)
								} else { // 如果是 ε，则只保留原产生式去掉第一个非终结符后的部分
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
					isChanged = true
				}

				// 用新产生式替换旧产生式
				g.Productions[it1] = newP
			}
		}
	}
	return
}

// FirstLetterSubstitutionForCommonFactor 按产生式来处理代入替换，不按顺序
// 直到没有非终结符开头的产生式为止
func (g *Grammar) FirstLetterSubstitutionForCommonFactor() {

	for {
		// 判断是否有产生式发生改变
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
							// 若产生式右部不为 ε
							if ntRight != "ε" {
								// 若首字符为非终结符，则不替换
								if g.containsNonTerminal(string(ntRight[0])) {
									continue
								}
								// 替换产生式右部非终结符
								newRight := ntRight + right[len(firstChar):]
								newP = append(newP, newRight)
							} else { // 若产生式右部为 ε, 则去掉右部非终结符
								newRight := right[len(firstChar):]

								// 如果新产生式为空，则形如 A -> ε
								if newRight == "" {
									newRight = "ε"
								}
								newP = append(newP, newRight)
							}
						}

						// 记录进行了修改
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
	// 提取隐藏的左公共因子
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

				// 若存在两个产生式含有左公共因子，则返回 True
				if firstI == string([]rune(rights[j])[0]) {
					return true
				}
			}
		}
	}
	return false
}

func (g *Grammar) HaveLeftRecursion() bool {

	// 克隆原文法，避免
	g2 := g.clone()

	// 消除间接左递归
	g2.FirstLetterSubstitution()

	// 判断是否存在直接左递归
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

	// 从开始符出发，遍历栈中元素
	for len(stack) > 0 {
		ch := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// 遍历该非终结符的产生式右部中的所有非终结符
		for _, right := range g.Productions[ch] {
			for _, ch := range right {
				s := string(ch)
				// 如果为非终结符且非添加到可到达集合，则添加
				if g.containsNonTerminal(s) && !reachable.Contains(s) {
					reachable.Add(s)
					stack = append(stack, s)
				}
			}
		}
	}

	// 删除不可到达的非终结符
	newNonTerminals := make([]string, 0, len(g.NonTerminalSet))
	for _, nt := range g.NonTerminalSet {
		if reachable.Contains(nt) {
			newNonTerminals = append(newNonTerminals, nt)
		}
	}
	g.NonTerminalSet = newNonTerminals

	// 删除不可到达的非终结符的产生式
	for left := range g.Productions {
		if !reachable.Contains(left) {
			delete(g.Productions, left)
		}
	}
}

func (g *Grammar) ParsingCommonFactor() {
	for {
		// 记录是否有产生式变化
		changed := false
		for _, left := range g.NonTerminalSet {
			rights, ok := g.Productions[left]
			if !ok || len(rights) < 2 {
				continue
			}

			// 记录产生式右部最长长度
			maxLen := 0
			for _, right := range rights {
				if l := len([]rune(right)); l > maxLen {
					maxLen = l
				}
			}
			// 遍历右部长度，先匹配最长的
			for length := maxLen; length >= 1 && !changed; length-- {
				prefixGroups := map[string][]int{}
				for i, right := range rights {
					rr := []rune(right)
					if len(rr) < length {
						continue
					}
					// 将具有相同前缀的记录到prefixGroups中
					prefix := string(rr[:length])
					prefixGroups[prefix] = append(prefixGroups[prefix], i)
				}

				// 枚举前缀
				for prefix, indices := range prefixGroups {
					// 若只有一个产生式，则不用提取左公共因子
					if len(indices) <= 1 {
						continue
					}

					newNT := g.getNewNonTerminal()
					g.NonTerminalSet = append(g.NonTerminalSet, newNT)

					newNTProductions := make([]string, 0, len(indices))
					remainingRights := make([]string, 0, len(rights)-len(indices))
					remove := map[int]struct{}{}

					// 记录需要提取左公共因子的下标
					for _, idx := range indices {
						remove[idx] = struct{}{}
					}

					// 遍历需要提取左公共因子的产生式
					for idx, right := range rights {
						if _, skip := remove[idx]; skip {
							rr := []rune(right)
							// 若公因子部分==产生式长度，则提取后为ε
							if len(rr) == length {
								newNTProductions = append(newNTProductions, "ε")
							} else { // 否则，提取后为公因子右部
								newNTProductions = append(newNTProductions, string(rr[length:]))
							}
						} else {
							// 保留无需提取的产生式
							remainingRights = append(remainingRights, right)
						}
					}

					// 原非终结符添加到新非终结符的产生式
					g.Productions[left] = append(remainingRights, prefix+newNT)

					// 添加新非终结符产生式
					g.Productions[newNT] = append(g.Productions[newNT], newNTProductions...)

					// 记录产生式发生变化
					changed = true
					break
				}

				if changed {
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
		// 产生式集合是否发生改变
		changed := false

		// 遍历所有产生式
		for _, left := range g.NonTerminalSet {
			rights, ok := g.Productions[left]
			if !ok || len(rights) == 0 {
				continue
			}

			// 记录形如A -> Ab的产生式
			recurIndices := make([]int, 0)
			// 记录形如A -> b 的产生式
			endIndices := make([]int, 0)

			for i, right := range rights {
				// 若形如A -> Ab， 则记录在recurIndices中
				if strings.HasPrefix(right, left) {
					recurIndices = append(recurIndices, i)
				} else { // 若形如A -> b， 则记录在endIndices中
					endIndices = append(endIndices, i)
				}
			}

			// 如无直接左递归，则遍历下一个非终结符
			if len(recurIndices) == 0 {
				continue
			}

			// 获取新的非终结符表示
			newNT := g.getNewNonTerminal()
			g.NonTerminalSet = append(g.NonTerminalSet, newNT)

			newNTProductions := make([]string, 0, len(recurIndices)+1)

			// 构造新非终结符的产生式, 例如A -> Ab, 构造B -> bB
			for _, idx := range recurIndices {
				right := rights[idx]
				rr := []rune(right)

				// 取出除去非终结符部分的右部作为前缀
				suffix := string(rr[len([]rune(left)):])

				// 构造产生式
				newNTProductions = append(newNTProductions, suffix+newNT)
			}

			// 添加 -> ε 的产生式
			newNTProductions = append(newNTProductions, "ε")

			// 构造原非终结符到新终结符的产生式，例如A -> b, 构造A -> bB
			newLeftProductions := make([]string, 0, len(endIndices))
			for _, idx := range endIndices {
				right := rights[idx]
				// 若原产生式右部为 ε, 则新右部为新非终结符
				if right == "ε" {
					newLeftProductions = append(newLeftProductions, newNT)
				} else { // 否则为原右部 + 新非终结符
					newLeftProductions = append(newLeftProductions, right+newNT)
				}
			}

			if len(newLeftProductions) == 0 {
				newLeftProductions = append(newLeftProductions, newNT)
			}

			// 将新产生式加入文法中
			g.Productions[left] = newLeftProductions
			g.Productions[newNT] = append(g.Productions[newNT], newNTProductions...)
			changed = true
			break
		}

		// 若本次遍历无新产生式产生，则结束循环
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
