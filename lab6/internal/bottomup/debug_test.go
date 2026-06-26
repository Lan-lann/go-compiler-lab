package bottomup

import "testing"

func TestLexerValues(t *testing.T) {
	l := newLexer("2+3*5")
	if l.tok != tokNum || l.val != 2 {
		t.Fatalf("expected first token num=2, got tok=%v val=%v", l.tok, l.val)
	}
	l.nextToken()
	if l.tok != tokPlus {
		t.Fatalf("expected plus, got %v", l.tok)
	}
	l.nextToken()
	if l.tok != tokNum || l.val != 3 {
		t.Fatalf("expected second token num=3, got tok=%v val=%v", l.tok, l.val)
	}
	l.nextToken()
	if l.tok != tokMul {
		t.Fatalf("expected mul, got %v", l.tok)
	}
	l.nextToken()
	if l.tok != tokNum || l.val != 5 {
		t.Fatalf("expected third token num=5, got tok=%v val=%v", l.tok, l.val)
	}
}

func TestParseExpression(t *testing.T) {
	result, err := ParseExpression("2+3*5")
	if err != nil {
		t.Fatalf("expected no parse error, got %v", err)
	}
	if result != 17 {
		t.Fatalf("expected 17, got %d", result)
	}
}
