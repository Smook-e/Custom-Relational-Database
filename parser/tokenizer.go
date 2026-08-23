package parser
import (
	"fmt"
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
)
type Token struct {
    Type  TokenType
    Value string
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
		}
	}
	return tokens, nil
}

func isLetter(ch rune) bool {
    return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}
func isDigit(ch rune) bool {
    return ch >= '0' && ch <= '9'
}