package parser

import (
	"fmt"
	"strings"
	"github.com/Smook-e/Custom-Relational-Database/storage"
)

type Parser struct {
	tokens []Token
	position int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		position: 0,
	}
}
