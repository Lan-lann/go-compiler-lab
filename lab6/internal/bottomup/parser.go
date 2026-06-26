package bottomup

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type token int

const (
	tokEOF token = iota
	tokNum
	tokPlus
	tokMul
	tokLParen
	tokRParen
	tokIllegal
)

type Token struct {
	type_  token
	lexval int
}

type Lexer struct {
	input string
	pos   int
	ch    byte
	tok   token
	val   int
}

type production struct {
	left  string
	right []string
}

var productions = []production{
	{},
	{left: "E", right: []string{"E", "+", "T"}},
	{left: "E", right: []string{"T"}},
	{left: "T", right: []string{"T", "*", "F"}},
	{left: "T", right: []string{"F"}},
	{left: "F", right: []string{"(", "E", ")"}},
	{left: "F", right: []string{"d"}},
}

var actionTable = map[int]map[string]string{
	0:  {"d": "s5", "(": "s4"},
	1:  {"+": "s6", "#": "acc"},
	2:  {"+": "r2", "*": "s7", ")": "r2", "#": "r2"},
	3:  {"+": "r4", "*": "r4", ")": "r4", "#": "r4"},
	4:  {"d": "s5", "(": "s4"},
	5:  {"+": "r6", "*": "r6", ")": "r6", "#": "r6"},
	6:  {"d": "s5", "(": "s4"},
	7:  {"d": "s5", "(": "s4"},
	8:  {"+": "r1", "*": "s7", ")": "r1", "#": "r1"},
	9:  {"+": "r3", "*": "r3", ")": "r3", "#": "r3"},
	10: {"+": "s6", ")": "s11"},
	11: {"+": "r5", "*": "s7", ")": "r5", "#": "r5"},
}

var gotoTable = map[int]map[string]int{
	0: {"E": 1, "T": 2, "F": 3},
	4: {"E": 10, "T": 2, "F": 3},
	6: {"T": 8, "F": 3},
	7: {"F": 9},
}

func newLexer(input string) *Lexer {
	l := &Lexer{input: strings.TrimSpace(input)}
	l.readChar()
	l.nextToken()
	return l
}

func (l *Lexer) readChar() {
	if l.pos >= len(l.input) {
		l.ch = 0
		return
	}
	l.ch = l.input[l.pos]
	l.pos++
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readNumber() int {
	start := l.pos - 1
	for l.ch >= '0' && l.ch <= '9' {
		l.readChar()
	}
	end := l.pos
	if l.ch != 0 {
		end = l.pos - 1
	}
	numText := l.input[start:end]
	num, err := strconv.Atoi(numText)
	if err != nil {
		return 0
	}
	return num
}

func (l *Lexer) nextToken() {
	l.skipWhitespace()
	if l.ch == 0 {
		l.tok = tokEOF
		return
	}

	switch l.ch {
	case '+':
		l.tok = tokPlus
		l.val = 0
		l.readChar()
	case '*':
		l.tok = tokMul
		l.val = 0
		l.readChar()
	case '(':
		l.tok = tokLParen
		l.val = 0
		l.readChar()
	case ')':
		l.tok = tokRParen
		l.val = 0
		l.readChar()
	default:
		if l.ch >= '0' && l.ch <= '9' {
			l.tok = tokNum
			l.val = l.readNumber()
			return
		}
		l.tok = tokIllegal
		l.val = 0
		l.readChar()
	}
}

func tokenToSymbol(t token) string {
	switch t {
	case tokNum:
		return "d"
	case tokPlus:
		return "+"
	case tokMul:
		return "*"
	case tokLParen:
		return "("
	case tokRParen:
		return ")"
	case tokEOF:
		return "#"
	default:
		return "?"
	}
}

func ParseExpression(input string) (int, error) {
	lexer := newLexer(input)
	inputSymbols := parseInputSymbols(input)
	inputPos := 0
	stateStack := []int{0}
	symbolStack := []string{"#"}
	semanticStack := []int{0}
	step := 0

	writer := table.NewWriter()
	writer.SetTitle(fmt.Sprintf("对输入串 %s 的 LR(0) 语法分析过程", input))
	writer.AppendHeader(table.Row{"步骤", "状态栈", "符号栈", "语义栈", "剩余输入", "ACTION", "语义动作"})

	for {
		state := stateStack[len(stateStack)-1]
		sym := tokenToSymbol(lexer.tok)
		action := actionTable[state][sym]
		if action == "" {
			return 0, fmt.Errorf("syntax error at %q", sym)
		}

		semAction := ""
		if action == "acc" {
			writer.AppendRow(table.Row{step, formatStateStack(stateStack), strings.Join(symbolStack, " "), formatSemanticStack(semanticStack), strings.Join(inputSymbols[inputPos:], ""), action, semAction})
			break
		}

		if strings.HasPrefix(action, "s") {
			writer.AppendRow(table.Row{step, formatStateStack(stateStack), strings.Join(symbolStack, " "), formatSemanticStack(semanticStack), strings.Join(inputSymbols[inputPos:], ""), action, semAction})
			nextState, _ := strconv.Atoi(action[1:])
			symbolStack = append(symbolStack, sym)
			if lexer.tok == tokNum {
				semanticStack = append(semanticStack, lexer.val)
			} else {
				semanticStack = append(semanticStack, 0)
			}
			stateStack = append(stateStack, nextState)
			lexer.nextToken()
			inputPos++
		} else if strings.HasPrefix(action, "r") {
			prodIdx, _ := strconv.Atoi(action[1:])
			prod := productions[prodIdx]
			count := len(prod.right)

			values := make([]int, count)
			for i := count - 1; i >= 0; i-- {
				values[i] = semanticStack[len(semanticStack)-count+i]
			}
			semAction := formatSemanticAction(prodIdx, values)
			writer.AppendRow(table.Row{step, formatStateStack(stateStack), strings.Join(symbolStack, " "), formatSemanticStack(semanticStack), strings.Join(inputSymbols[inputPos:], ""), action, semAction})

			for i := count - 1; i >= 0; i-- {
				semanticStack = semanticStack[:len(semanticStack)-1]
				symbolStack = symbolStack[:len(symbolStack)-1]
				stateStack = stateStack[:len(stateStack)-1]
			}

			result := evaluate(prodIdx, values)
			semanticStack = append(semanticStack, result)
			symbolStack = append(symbolStack, prod.left)

			gotoState, ok := gotoTable[stateStack[len(stateStack)-1]][prod.left]
			if !ok {
				return 0, fmt.Errorf("syntax error: no goto from state %d by %s", stateStack[len(stateStack)-1], prod.left)
			}
			stateStack = append(stateStack, gotoState)
		} else {
			return 0, fmt.Errorf("invalid action %s", action)
		}

		step++
	}

	writer.SetStyle(table.StyleRounded)
	writer.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignCenter},
		{Number: 2, Align: text.AlignLeft},
		{Number: 3, Align: text.AlignLeft},
		{Number: 4, Align: text.AlignLeft},
		{Number: 5, Align: text.AlignRight},
		{Number: 6, Align: text.AlignCenter},
	})
	fmt.Println(writer.Render())

	return semanticStack[len(semanticStack)-1], nil
}

func parseInputSymbols(input string) []string {
	lexer := newLexer(input)
	symbols := []string{}
	for {
		symbols = append(symbols, tokenToSymbol(lexer.tok))
		if lexer.tok == tokEOF {
			break
		}
		lexer.nextToken()
	}
	return symbols
}

func formatStateStack(stack []int) string {
	parts := make([]string, len(stack))
	for i, v := range stack {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, " ")
}

func formatSemanticStack(stack []int) string {
	parts := make([]string, len(stack))
	for i, v := range stack {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, " ")
}

func formatSemanticAction(prodIdx int, values []int) string {
	result := evaluate(prodIdx, values)
	var action string
	switch prodIdx {
	case 1:
		action = fmt.Sprintf("E = %d + %d", values[0], values[2])
	case 2:
		action = fmt.Sprintf("E = %d", values[0])
	case 3:
		action = fmt.Sprintf("T = %d * %d", values[0], values[2])
	case 4:
		action = fmt.Sprintf("T = %d", values[0])
	case 5:
		action = fmt.Sprintf("F = (%d)", values[1])
	case 6:
		action = fmt.Sprintf("F = %d", values[0])
	default:
		action = ""
	}
	if action == "" {
		return ""
	}
	return fmt.Sprintf("%s => %d", action, result)
}

func tokenString(t token) string {
	switch t {
	case tokNum:
		return "d"
	case tokPlus:
		return "+"
	case tokMul:
		return "*"
	case tokLParen:
		return "("
	case tokRParen:
		return ")"
	case tokEOF:
		return "#"
	default:
		return "?"
	}
}

func evaluate(prodIdx int, values []int) int {
	switch prodIdx {
	case 1:
		return values[0] + values[2]
	case 2:
		return values[0]
	case 3:
		return values[0] * values[2]
	case 4:
		return values[0]
	case 5:
		return values[1]
	case 6:
		return values[0]
	default:
		return 0
	}
}
