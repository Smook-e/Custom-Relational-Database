package parser

import (
	"fmt"
	// "strings"
	"github.com/Smook-e/Custom-Relational-Database/storage"
)

type Parser struct {
	tokens []Token
	position int
}
type SelectStatement struct {
	Columns []string
	Tablename string
	Where *storage.Expression
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		position: 0,
	}
}
func (p *Parser) Peek() Token {	
	return p.tokens[p.position]
}

func (p *Parser) Get() Token {
	token := p.tokens[p.position]
	p.position++
	return token
}
func (p *Parser) Expect(expectedType TokenType, val string) (Token, error) {
	token := p.Peek()
	if token.Type != expectedType {
		return token, fmt.Errorf("expected token of type %v, got %v", expectedType, token.Type)
	}
	if val != "" && token.Value != val {
		return token, fmt.Errorf("expected token with value %v, got %v", val, token.Value)
	}
	return p.Get(), nil
}
