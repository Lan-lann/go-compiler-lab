package topdown

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

type token int

const (
	tokenEOF token = iota
	tokenPlus
	tokenMinus
	tokenNum
	tokenIllegal
)

type ASTNode struct {
	ID       string
	Label    string
	Children []*ASTNode
}

var nodeCounter int

func newASTNode(label string, children ...*ASTNode) *ASTNode {
	nodeCounter++
	return &ASTNode{
		ID:       fmt.Sprintf("node%d", nodeCounter),
		Label:    label,
		Children: children,
	}
}

type Lexer struct {
	input string
	pos   int
	ch    byte
	token token
	num   int
}

func newLexer(input string) *Lexer {
	l := &Lexer{input: strings.TrimSpace(input)}
	l.readChar()
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

func (l *Lexer) nextToken() {
	l.skipWhitespace()

	if l.ch == 0 {
		l.token = tokenEOF
		return
	}

	switch l.ch {
	case '+':
		l.token = tokenPlus
		l.readChar()
	case '-':
		l.token = tokenMinus
		l.readChar()
	default:
		if isDigit(l.ch) {
			l.token = tokenNum
			l.num = l.readNumber()
			return
		}
		l.token = tokenIllegal
		l.readChar()
	}
}

func (l *Lexer) matchToken(expected token) error {
	if l.token != expected {
		return syntaxError()
	}
	l.nextToken()
	return nil
}

func syntaxError() error {
	return fmt.Errorf("syntax error")
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func (l *Lexer) readNumber() int {
	start := l.pos - 1
	for isDigit(l.ch) {
		l.readChar()
	}

	end := l.pos - 1
	if l.ch == 0 {
		end = l.pos
	}

	numString := l.input[start:end]
	num, err := strconv.Atoi(numString)
	if err != nil {
		return 0
	}
	return num
}

func (l *Lexer) tokenString() string {
	switch l.token {
	case tokenEOF:
		return "EOF"
	case tokenPlus:
		return "+"
	case tokenMinus:
		return "-"
	case tokenNum:
		return fmt.Sprintf("num(%d)", l.num)
	default:
		return "ILLEGAL"
	}
}

// ParseExpression 解析简单表达式并计算语义结果。
// 语法：
//
//	E -> T R
//	R -> + T R | - T R | ε
//	T -> num
func ParseExpression(input string) (int, *ASTNode, error) {
	nodeCounter = 0
	lexer := newLexer(input)
	lexer.nextToken()

	result, tree, err := parseE(lexer)
	if err != nil {
		return 0, nil, err
	}

	if lexer.token != tokenEOF {
		return 0, nil, syntaxError()
	}

	return result, tree, nil
}

func parseE(l *Lexer) (int, *ASTNode, error) {
	fmt.Println("E -> T R")
	tVal, tNode, err := parseT(l)
	if err != nil {
		return 0, nil, err
	}

	fmt.Printf("R.in := %d\n", tVal)
	rVal, rNode, err := parseR(l, tVal)
	if err != nil {
		return 0, nil, err
	}

	root := newASTNode(fmt.Sprintf("E\nval=%d", rVal), tNode, rNode)
	fmt.Printf("E.val := %d\n", rVal)
	return rVal, root, nil
}

func parseR(l *Lexer, in int) (int, *ASTNode, error) {
	switch l.token {
	case tokenPlus:
		fmt.Println("R -> + T R")
		if err := l.matchToken(tokenPlus); err != nil {
			return 0, nil, err
		}
		tVal, tNode, err := parseT(l)
		if err != nil {
			return 0, nil, err
		}

		r1In := in + tVal
		fmt.Printf("R1.in := %d\n", r1In)
		rVal, rNode, err := parseR(l, r1In)
		if err != nil {
			return 0, nil, err
		}
		fmt.Printf("R.val := %d\n", rVal)
		node := newASTNode(fmt.Sprintf("R\nin=%d\nval=%d", in, rVal), newASTNode("+"), tNode, rNode)
		return rVal, node, nil

	case tokenMinus:
		fmt.Println("R -> - T R")
		if err := l.matchToken(tokenMinus); err != nil {
			return 0, nil, err
		}
		tVal, tNode, err := parseT(l)
		if err != nil {
			return 0, nil, err
		}

		r1In := in - tVal
		fmt.Printf("R1.in := %d\n", r1In)
		rVal, rNode, err := parseR(l, r1In)
		if err != nil {
			return 0, nil, err
		}
		fmt.Printf("R.val := %d\n", rVal)
		node := newASTNode(fmt.Sprintf("R\nin=%d\nval=%d", in, rVal), newASTNode("-"), tNode, rNode)
		return rVal, node, nil

	case tokenEOF:
		fmt.Println("R -> ε")
		fmt.Printf("R.val := %d\n", in)
		node := newASTNode(fmt.Sprintf("R\nval=%d", in), newASTNode("ε"))
		return in, node, nil

	default:
		return 0, nil, syntaxError()
	}
}

func parseT(l *Lexer) (int, *ASTNode, error) {
	if l.token != tokenNum {
		return 0, nil, syntaxError()
	}

	currentNum := l.num
	fmt.Printf("T -> num (%d)\n", currentNum)
	tVal := currentNum
	fmt.Printf("T.val := %d\n", tVal)

	leaf := newASTNode(fmt.Sprintf("num=%d", currentNum))
	root := newASTNode(fmt.Sprintf("T\nval=%d", tVal), leaf)
	if err := l.matchToken(tokenNum); err != nil {
		return 0, nil, err
	}
	return tVal, root, nil
}

func DrawAST(root *ASTNode, output string) error {
	ctx := context.Background()
	gv, err := graphviz.New(ctx)
	if err != nil {
		return err
	}
	defer gv.Close()

	graph, err := gv.Graph()
	if err != nil {
		return err
	}
	defer func() {
		_ = graph.Close()
	}()

	nodeMap := make(map[string]*cgraph.Node)
	var addNodes func(n *ASTNode) error
	addNodes = func(n *ASTNode) error {
		node, ok := nodeMap[n.ID]
		if !ok {
			var err error
			node, err = graph.CreateNodeByName(n.ID)
			if err != nil {
				return err
			}
			node.SetLabel(n.Label)
			nodeMap[n.ID] = node
		}

		for _, child := range n.Children {
			childNode, ok := nodeMap[child.ID]
			if !ok {
				var err error
				childNode, err = graph.CreateNodeByName(child.ID)
				if err != nil {
					return err
				}
				childNode.SetLabel(child.Label)
				nodeMap[child.ID] = childNode
			}

			if _, err := graph.CreateEdgeByName(n.ID+"->"+child.ID, node, childNode); err != nil {
				return err
			}
			if err := addNodes(child); err != nil {
				return err
			}
		}
		return nil
	}

	graph.SetRankDir(cgraph.TBRank)
	if err := addNodes(root); err != nil {
		return err
	}

	return gv.RenderFilename(ctx, graph, graphviz.PNG, output)
}
