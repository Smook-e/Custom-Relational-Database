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
