package parser

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

func isLetter(ch rune) bool {
    return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}
func isDigit(ch rune) bool {
    return ch >= '0' && ch <= '9'
}