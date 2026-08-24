package parser

import (
	"fmt"
	"strings"
)
type TokenType int

// Represents the different types of tokens that can be identified by the tokenizer.
const (
    TokenKeyword TokenType = iota
    TokenIdentifier
    TokenNumber
    TokenString
    TokenOperator
    TokenComma
    TokenLParen
    TokenRParen
    TokenEOF
	TokenSemicolon
)
type Token struct {
    Type  TokenType
    Value string
}
func DecodeTokenType(t TokenType) string {
	switch t {
	case TokenKeyword:
		return "Keyword"
	case TokenIdentifier:
		return "Identifier"
	case TokenNumber:
		return "Number"
	case TokenString:
		return "String"
	case TokenOperator:
		return "Operator"
	case TokenComma:
		return "Comma"
	case TokenLParen:
		return "Left Parenthesis"
	case TokenRParen:
		return "Right Parenthesis"
	case TokenEOF:
		return "EOF"
	case TokenSemicolon:
		return "Semicolon"
	default:
		return ""
	}
}
func (token *Token) Decode() string {
	switch token.Type {
	case TokenKeyword:
		return fmt.Sprintf("Keyword(%q)", token.Value)
	case TokenIdentifier:
		return "Identifier"
	case TokenNumber:
		return "Number"
	case TokenString:
		return "String"
	case TokenOperator:
		return "Operator"
	case TokenComma:
		return "Comma"
	case TokenLParen:
		return "Left Parenthesis"
	case TokenRParen:
		return "Right Parenthesis"
	case TokenEOF:
		return "EOF"
	case TokenSemicolon:
		return "Semicolon"
	default:
		return ""
	}
}
var keywords = map[string]bool{
    "SELECT": true, "FROM": true, "WHERE": true,
    "AND": true, "OR": true, "INSERT": true,
    "INTO": true, "VALUES": true,
}
func Tokenize(input string) ([]Token, error) {
	var tokens []Token
	i := 0
	runes := []rune(input)
	for i < len(runes) {
		ch := runes[i]
		switch {
			case ch == ' ' || ch == '\t' || ch == '\n':
				i++
			case ch == ',':
				tokens = append(tokens, Token{Type: TokenComma, Value: ","})
				i++
			case ch == '(':
				tokens = append(tokens, Token{Type: TokenLParen, Value: "("})
				i++
			case ch == ')':
				tokens = append(tokens, Token{Type: TokenRParen, Value: ")"})
				i++
			case ch == '=' || ch == '<' || ch == '>':
				start := i
				i++
				// handle >=, <=, != if next char extends the operator
				if i < len(runes) && runes[i] == '=' {
					i++
				}
				tokens = append(tokens, Token{TokenOperator, string(runes[start:i])})

			case ch == '\'':// Start of a string literal
				i++
				start := i // Skip the opening quote
				for i < len(runes) && runes[i] != '\'' {
					i++
				}
				if i >= len(runes) {
					return nil, fmt.Errorf("unterminated string literal")
				}
				tokens = append(tokens, Token{Type: TokenString, Value: string(runes[start:i])})
				i++ // Skip the closing quote
			case isLetter(ch):
				start := i
				for i < len(runes) && (isLetter(runes[i]) || isDigit(runes[i]) || runes[i] == '_') {
					i++
				}
				word := string(runes[start:i])
				upper := strings.ToUpper(word)
				if keywords[upper] {
					tokens = append(tokens, Token{Type: TokenKeyword, Value: upper})
				} else {
					tokens = append(tokens, Token{Type: TokenIdentifier, Value: word})
				}
			case isDigit(ch):
				start := i
				for i < len(runes) && isDigit(runes[i]) {
					i++
				}
				tokens = append(tokens, Token{Type: TokenNumber, Value: string(runes[start:i])})
			case ch == ';':
				tokens = append(tokens, Token{Type: TokenSemicolon, Value: ";"})
				i++
			case ch == '*':
				tokens = append(tokens, Token{TokenIdentifier, "*"})
				i++
			default:
				return nil, fmt.Errorf("unexpected character: %q at position %d", ch, i)
		}
	}
	return tokens, nil
}

func isLetter(ch rune) bool {
    return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch == '_') 
}
func isDigit(ch rune) bool {
    return ch >= '0' && ch <= '9'
}