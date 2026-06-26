package topdown

import (
	"fmt"
	"strconv"
	"strings"
)

type token int

const (
	tokenEOF token = iota
	tokenPlus
	tokenMinus
	tokenNum
	tokenIllegal
)

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
func ParseExpression(input string) (int, error) {
	lexer := newLexer(input)
	lexer.nextToken()

	result, err := parseE(lexer)
	if err != nil {
		return 0, err
	}

	if lexer.token != tokenEOF {
		return 0, syntaxError()
	}

	return result, nil
}

func parseE(l *Lexer) (int, error) {
	fmt.Println("E -> T R")
	tVal, err := parseT(l)
	if err != nil {
		return 0, err
	}

	rVal, err := parseR(l, tVal)
	if err != nil {
		return 0, err
	}

	fmt.Printf("E.val := %d\n", rVal)
	return rVal, nil
}

func parseR(l *Lexer, in int) (int, error) {
	switch l.token {
	case tokenPlus:
		fmt.Println("R -> + T R")
		if err := l.matchToken(tokenPlus); err != nil {
			return 0, err
		}
		tVal, err := parseT(l)
		if err != nil {
			return 0, err
		}

		r1In := in + tVal
		fmt.Printf("R1.in := %d\n", r1In)
		rVal, err := parseR(l, r1In)
		if err != nil {
			return 0, err
		}
		fmt.Printf("R.val := %d\n", rVal)
		return rVal, nil

	case tokenMinus:
		fmt.Println("R -> - T R")
		if err := l.matchToken(tokenMinus); err != nil {
			return 0, err
		}
		tVal, err := parseT(l)
		if err != nil {
			return 0, err
		}

		r1In := in - tVal
		fmt.Printf("R1.in := %d\n", r1In)
		rVal, err := parseR(l, r1In)
		if err != nil {
			return 0, err
		}
		fmt.Printf("R.val := %d\n", rVal)
		return rVal, nil

	case tokenEOF:
		fmt.Println("R -> ε")
		fmt.Printf("R.val := %d\n", in)
		return in, nil

	default:
		return 0, syntaxError()
	}
}

func parseT(l *Lexer) (int, error) {
	currentNum := l.num
	if err := l.matchToken(tokenNum); err != nil {
		return 0, err
	}

	fmt.Printf("T -> num (%d)\n", currentNum)
	tVal := currentNum
	fmt.Printf("T.val := %d\n", tVal)

	return tVal, nil
}
