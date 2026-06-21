package grammar

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type Item struct {
	Left  string
	Right string
	Dot   int
}

func (item Item) String() string {
	if item.Right == "ε" || item.Right == "" {
		if item.Dot == 0 {
			return fmt.Sprintf("%s -> .", item.Left)
		}
		return fmt.Sprintf("%s -> ε.", item.Left)
	}

	rt := []rune(item.Right)
	return fmt.Sprintf("%s -> %s.%s", item.Left, string(rt[:item.Dot]), string(rt[item.Dot:]))
}

func (item Item) SymbolAfterDot() (string, bool) {
	syms := parseSymbols(item.Right)
	if item.Dot < len(syms) {
		return syms[item.Dot], true
	}
	return "", false
}

func (item Item) IsReduceItem() bool {
	syms := parseSymbols(item.Right)
	return item.Dot == len(syms)
}

// Items 构造文法的所有 LR(0) 项。
func (g *Grammar) Items() []Item {
	items := make([]Item, 0)
	for left, rights := range g.Productions {
		for _, right := range rights {
			syms := parseSymbols(right)
			for dot := 0; dot <= len(syms); dot++ {
				items = append(items, Item{Left: left, Right: right, Dot: dot})
			}
		}
	}
	return items
}

func parseSymbols(right string) []string {
	if right == "ε" || right == "" {
		return nil
	}

	syms := make([]string, 0, len([]rune(right)))
	for _, r := range right {
		syms = append(syms, string(r))
	}
	return syms
}

func itemsToSet(items []Item) map[Item]struct{} {
	set := make(map[Item]struct{}, len(items))
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

func setToItems(set map[Item]struct{}) []Item {
	items := make([]Item, 0, len(set))
	for item := range set {
		items = append(items, item)
	}
	return items
}

func itemSetKey(items []Item) string {
	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, item.String())
	}
	sort.Strings(keys)
	return strings.Join(keys, "|")
}

// Closure 计算 LR(0) 项集的闭包。
func (g *Grammar) Closure(items []Item) []Item {
	closureSet := itemsToSet(items)
	changed := true

	for changed {
		changed = false
		for item := range closureSet {
			sym, ok := item.SymbolAfterDot()
			if !ok {
				continue
			}
			if !g.containsNonTerminal(sym) {
				continue
			}
			for _, right := range g.Productions[sym] {
				newItem := Item{Left: sym, Right: right, Dot: 0}
				if _, exists := closureSet[newItem]; !exists {
					closureSet[newItem] = struct{}{}
					changed = true
				}
			}
		}
	}

	return setToItems(closureSet)
}

// Goto 计算从给定项集经符号 X 的转移。
func (g *Grammar) Goto(items []Item, symbol string) []Item {
	moved := make([]Item, 0)
	for _, item := range items {
		sym, ok := item.SymbolAfterDot()
		if !ok || sym != symbol {
			continue
		}
		moved = append(moved, Item{Left: item.Left, Right: item.Right, Dot: item.Dot + 1})
	}
	return g.Closure(moved)
}

type ItemSet struct {
	ID       int
	Items    map[Item]struct{}
	Goto     map[string]int
	IsReduce bool
}

func newItemSet(id int, items []Item) *ItemSet {
	set := itemsToSet(items)
	isReduce := false
	for item := range set {
		if item.IsReduceItem() {
			isReduce = true
			break
		}
	}
	return &ItemSet{
		ID:       id,
		Items:    set,
		Goto:     make(map[string]int),
		IsReduce: isReduce,
	}
}

// LR0ItemSets 构造 LR(0) 项目族。
func (g *Grammar) LR0ItemSets() []*ItemSet {
	startRight := g.Productions[g.Start][0]
	startItems := g.Closure([]Item{{Left: g.Start, Right: startRight, Dot: 0}})
	collection := make([]*ItemSet, 0)
	index := make(map[string]int)
	queue := make([]*ItemSet, 0)

	initial := newItemSet(0, startItems)
	collection = append(collection, initial)
	index[itemSetString(initial)] = 0
	queue = append(queue, initial)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		symbols := g.allSymbols()
		for _, symbol := range symbols {
			moved := g.Goto(setToItems(current.Items), symbol)
			if len(moved) == 0 {
				continue
			}
			key := itemSetKey(moved)
			if id, ok := index[key]; ok {
				current.Goto[symbol] = id
				continue
			}
			newID := len(collection)
			newSet := newItemSet(newID, moved)
			collection = append(collection, newSet)
			index[key] = newID
			current.Goto[symbol] = newID
			queue = append(queue, newSet)
		}
	}

	return collection
}

func itemSetString(set *ItemSet) string {
	return itemSetKey(setToItems(set.Items))
}

func (g *Grammar) allSymbols() []string {
	symbols := make([]string, 0, len(g.NonTerminalSet)+len(g.TerminalSet))
	symbols = append(symbols, g.NonTerminalSet...)
	symbols = append(symbols, g.TerminalSet...)
	return symbols
}

// AugmentGrammar 拓广文法，将文法开始符号扩展为一个新的非终结符。
// 例如，原始文法开始符号为 S，则添加 S' -> S，并将 Start 修改为 S'.
func (g *Grammar) AugmentGrammar() {
	newStart := g.getAugmentedStartSymbol()
	g.NonTerminalSet = append([]string{newStart}, g.NonTerminalSet...)
	g.Productions[newStart] = []string{g.Start}
	g.Start = newStart
}

func (g *Grammar) getAugmentedStartSymbol() string {
	candidates := []string{
		"Z", "Y", "X", "W", "V", "U", "T", "R", "Q", "P", "O", "N", "M", "L", "K", "J", "I", "H", "G", "F", "E", "D", "C", "B", "A",
		"z", "y", "x", "w", "v", "u", "t", "r", "q", "p", "o", "n", "m", "l", "k", "j", "i", "h", "g", "f", "e", "d", "c", "b", "a",
		"@", "$", "%", "&", "*", "!", "?",
	}

	for _, sym := range candidates {
		if !g.containsSymbol(sym) {
			return sym
		}
	}

	// 若所有单字符符号都被占用，则继续生成以 S 开头的唯一符号。
	suffix := 0
	for {
		sym := fmt.Sprintf("S%d", suffix)
		if !g.containsSymbol(sym) {
			return sym
		}
		suffix++
	}
}

func (g *Grammar) containsSymbol(sym string) bool {
	for _, nt := range g.NonTerminalSet {
		if nt == sym {
			return true
		}
	}
	for _, t := range g.TerminalSet {
		if t == sym {
			return true
		}
	}
	return false
}

// PrintItemSets 打印 LR(0) 项目族的详细信息
func (g *Grammar) PrintItemSets() {
	sets := g.LR0ItemSets()
	for _, s := range sets {
		fmt.Printf("ItemSet %d:\n", s.ID)

		// 收集并排序项目以保证输出稳定
		items := make([]string, 0, len(s.Items))
		for it := range s.Items {
			items = append(items, it.String())
		}
		sort.Strings(items)
		for _, itstr := range items {
			fmt.Printf("  %s\n", itstr)
		}

		fmt.Printf("  IsReduce: %v\n", s.IsReduce)

		if len(s.Goto) > 0 {
			syms := make([]string, 0, len(s.Goto))
			for sym := range s.Goto {
				syms = append(syms, sym)
			}
			sort.Strings(syms)
			fmt.Printf("  Goto:\n")
			for _, sym := range syms {
				fmt.Printf("    %s -> %d\n", sym, s.Goto[sym])
			}
		}

		fmt.Println()
	}
}

// DrawItemSetsGraph 将 LR(0) 项目族以有向图形式导出为 PNG，输出路径由 output 指定
func (g *Grammar) DrawItemSetsGraph(output string) error {
	sets := g.LR0ItemSets()

	ctx := context.Background()
	gv, err := graphviz.New(ctx)
	if err != nil {
		return err
	}
	graph, err := gv.Graph()
	if err != nil {
		gv.Close()
		return err
	}
	defer func() {
		_ = graph.Close()
		gv.Close()
	}()

	graph.SetRankDir(cgraph.LRRank)

	nodeMap := make(map[int]*cgraph.Node)

	for _, s := range sets {
		name := fmt.Sprintf("I%d", s.ID)
		n, err := graph.CreateNodeByName(name)
		if err != nil {
			return err
		}

		// 构造节点标签：ID 与项目列表
		items := make([]string, 0, len(s.Items))
		for it := range s.Items {
			items = append(items, it.String())
		}
		sort.Strings(items)
		label := fmt.Sprintf("%s\n", name)
		for _, it := range items {
			label += it + "\\l"
		}
		n.SetLabel(label)

		// 标记含归约项目的集合
		if s.IsReduce {
			n.SetShape(cgraph.BoxShape)
		} else {
			n.SetShape(cgraph.BoxShape)
		}

		nodeMap[s.ID] = n
	}

	// 添加转移边
	for _, s := range sets {
		from := nodeMap[s.ID]
		for sym, toID := range s.Goto {
			to := nodeMap[toID]
			if to == nil || from == nil {
				continue
			}
			e, err := graph.CreateEdgeByName(sym, from, to)
			if err != nil {
				return err
			}
			e.SetLabel(sym)
		}
	}

	if err := gv.RenderFilename(ctx, graph, graphviz.PNG, output); err != nil {
		return err
	}
	return nil
}

// BuildParsingTable 构造 LR(0) 的 ACTION 和 GOTO 表
func (g *Grammar) BuildParsingTable() (map[int]map[string]string, map[int]map[string]int, []string) {
	sets := g.LR0ItemSets()

	// 产生式列表（保持顺序），以及从产生式到编号的映射
	prodList := make([]string, 0)
	prodIndex := make(map[string]int)
	for _, left := range g.NonTerminalSet {
		for _, right := range g.Productions[left] {
			key := left + "->" + right
			prodIndex[key] = len(prodList)
			prodList = append(prodList, key)
		}
	}

	// 表格初始化
	action := make(map[int]map[string]string)
	gotoTable := make(map[int]map[string]int)

	// 列表化终结符与非终结符（用于打印/遍历）
	termList := append([]string{}, g.TerminalSet...)
	termList = append(termList, "#")

	for _, s := range sets {
		i := s.ID
		action[i] = make(map[string]string)
		gotoTable[i] = make(map[string]int)

		// shift 操作 & goto（来自 s.Goto）
		for sym, to := range s.Goto {
			if containsString(g.TerminalSet, sym) {
				action[i][sym] = "s" + strconv.Itoa(to)
			} else if containsString(g.NonTerminalSet, sym) {
				gotoTable[i][sym] = to
			}
		}

		// reduce / accept
		for it := range s.Items {
			if !it.IsReduceItem() {
				continue
			}

			// augmented start 的归约即为接受
			// augmented start 的左部等于 g.Start
			if it.Left == g.Start {
				action[i]["#"] = "ACC"
				continue
			}

			// 否则为归约：为所有终结符（及#）设置 rK
			key := it.Left + "->" + it.Right
			idx, ok := prodIndex[key]
			if !ok {
				// should not happen
				continue
			}
			for _, a := range termList {
				// 若已有 shift 冲突则优先保留 shift
				if cur, exists := action[i][a]; exists && strings.HasPrefix(cur, "s") {
					continue
				}
				action[i][a] = "r" + strconv.Itoa(idx)
			}
		}
	}

	return action, gotoTable, prodList
}

func containsString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// PrintParsingTable 打印 ACTION 与 GOTO 表，参照图片表格格式
func (g *Grammar) PrintParsingTable() {
	action, gotoTable, prodList := g.BuildParsingTable()

	terms := append([]string{}, g.TerminalSet...)
	terms = append(terms, "#")
	nonterms := []string{}

	for _, non := range g.NonTerminalSet {
		if non != g.Start {
			nonterms = append(nonterms, non)
		}
	}

	// 创建表格
	writer := table.NewWriter()
	writer.SetTitle("LR(0)分析表")

	header2 := make(table.Row, 0, 1+len(terms)+len(nonterms))
	header2 = append(header2, "状态")
	for i := 0; i < len(terms); i++ {
		header2 = append(header2, "ACTION")
	}
	for i := 0; i < len(nonterms); i++ {
		header2 = append(header2, "GOTO")
	}
	writer.AppendRow(header2, table.RowConfig{AutoMerge: true})
	writer.AppendSeparator()
	header1 := make(table.Row, 0, 1+len(terms)+len(nonterms))
	header1 = append(header1, "状态")
	for _, t := range terms {
		header1 = append(header1, t)
	}
	for _, nt := range nonterms {
		header1 = append(header1, nt)
	}
	writer.AppendRow(header1, table.RowConfig{AutoMerge: true})

	writer.AppendSeparator()

	writer.SetColumnConfigs([]table.ColumnConfig{
		{
			Number:    1,
			AutoMerge: true,
			Align:     text.AlignCenter,
		},
	})

	// states sorted
	n := 0
	sets := g.LR0ItemSets()
	for _, s := range sets {
		if s.ID >= n {
			n = s.ID + 1
		}
	}

	for i := 0; i < n; i++ {
		row := make(table.Row, 0, 1+len(terms)+len(nonterms))
		row = append(row, i)
		for _, t := range terms {
			if v, ok := action[i][t]; ok {
				row = append(row, v)
			} else {
				row = append(row, "")
			}
		}
		for _, nt := range nonterms {
			if v, ok := gotoTable[i][nt]; ok {
				row = append(row, v)
			} else {
				row = append(row, "")
			}
		}
		writer.AppendRow(row)
	}

	writer.SetStyle(table.StyleRounded)
	fmt.Println(writer.Render())

	fmt.Println("Productions:")
	for idx, p := range prodList {
		fmt.Printf("%d: %s\n", idx, p)
	}
}

// PrintParseProcess 逐步打印 LR(0) 语法分析过程
func (g *Grammar) PrintParseProcess(input string) {
	action, gotoTable, prodList := g.BuildParsingTable()

	inputSymbols := parseSymbols(input)
	inputSymbols = append(inputSymbols, "#")

	stateStack := []int{0}
	symbolStack := []string{"#"}

	writer := table.NewWriter()
	writer.SetTitle(fmt.Sprintf("对输入串%s的LR(0)分析过程", input))
	header := table.Row{"步骤", "状态栈", "符号栈", "输入串", "ACTION", "GOTO"}
	writer.AppendHeader(header)

	step := 1
	for {
		currentState := stateStack[len(stateStack)-1]
		a := inputSymbols[0]
		actionStr := action[currentState][a]
		if actionStr == "" {
			actionStr = "ERR"
		}

		gotoStr := ""

		// 预计算若为 reduce 时会跳转到的 GOTO 目标（用于打印），但不执行修改
		if strings.HasPrefix(actionStr, "r") {
			reduceIdx, err := strconv.Atoi(actionStr[1:])
			if err == nil && reduceIdx >= 0 && reduceIdx < len(prodList) {
				prod := prodList[reduceIdx]
				parts := strings.SplitN(prod, "->", 2)
				left := parts[0]
				right := parts[1]

				// 模拟弹出以找到归约后栈顶状态
				tmpStateStack := append([]int(nil), stateStack...)
				if right != "" && right != "ε" {
					rhs := parseSymbols(right)
					for range rhs {
						if len(tmpStateStack) > 0 {
							tmpStateStack = tmpStateStack[:len(tmpStateStack)-1]
						}
					}
				}
				if len(tmpStateStack) > 0 {
					fromState := tmpStateStack[len(tmpStateStack)-1]
					if nextState, ok := gotoTable[fromState][left]; ok {
						gotoStr = strconv.Itoa(nextState)
					} else {
						gotoStr = "ERR"
					}
				}
			}
		}

		// 先打印将要执行的 ACTION / GOTO
		writer.AppendRow(table.Row{fmt.Sprintf("(%d)", step), joinStateStack(stateStack), joinSymbolStack(symbolStack), strings.Join(inputSymbols, ""), actionStr, gotoStr})

		// 再执行动作
		if strings.HasPrefix(actionStr, "s") {
			j, err := strconv.Atoi(actionStr[1:])
			if err != nil {
				break
			}
			symbolStack = append(symbolStack, a)
			stateStack = append(stateStack, j)
			inputSymbols = inputSymbols[1:]
		} else if strings.HasPrefix(actionStr, "r") {
			reduceIdx, err := strconv.Atoi(actionStr[1:])
			if err != nil {
				break
			}
			if reduceIdx < 0 || reduceIdx >= len(prodList) {
				break
			}
			prod := prodList[reduceIdx]
			parts := strings.SplitN(prod, "->", 2)
			left := parts[0]
			right := parts[1]

			if right != "" && right != "ε" {
				rhs := parseSymbols(right)
				for range rhs {
					if len(symbolStack) > 0 {
						symbolStack = symbolStack[:len(symbolStack)-1]
					}
					if len(stateStack) > 0 {
						stateStack = stateStack[:len(stateStack)-1]
					}
				}
			}

			fromState := stateStack[len(stateStack)-1]
			nextState, ok := gotoTable[fromState][left]
			if !ok {
				break
			}
			symbolStack = append(symbolStack, left)
			stateStack = append(stateStack, nextState)
		} else if actionStr == "ACC" {
			// already printed
			break
		} else {
			// unknown/empty action -> error
			break
		}

		step++
	}

	writer.SetStyle(table.StyleRounded)
	writer.SetColumnConfigs([]table.ColumnConfig{
		{
			Number:    4,
			AutoMerge: true,
			Align:     text.AlignRight,
			VAlign:    text.VAlignDefault,
		},
	})
	fmt.Println(writer.Render())
}

func joinStateStack(states []int) string {
	var sb strings.Builder
	for _, s := range states {
		sint := strconv.Itoa(s)
		if len(sint) >= 2 {
			sb.WriteString(fmt.Sprintf("(%s)", sint))
		} else {
			sb.WriteString(fmt.Sprintf("%s", sint))
		}

	}
	return sb.String()
}

func joinSymbolStack(symbols []string) string {
	return strings.Join(symbols, "")
}
