package parser

import (
	"fmt"
	// "strings"
	"slices"
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
func (p *Parser) Expect(expectedType []TokenType, val string) (Token, error) {
	token := p.Peek()
	if !slices.Contains(expectedType, token.Type) {
		return token, fmt.Errorf("expected token of type %v, got %v", expectedType, token.Type)
	}
	if val != "" && token.Value != val {
		return token, fmt.Errorf("expected token with value %v, got %v", val, token.Value)
	}
	return p.Get(), nil
}
func ParseWhereCondition(p *Parser) (*storage.SearchCondition, error) {
	columnNameToken, err := p.Expect([]TokenType{TokenIdentifier}, "")
	if err != nil {
		return nil, err
	}
	operatorToken, err := p.Expect([]TokenType{TokenOperator}, "")
	if err != nil {
		return nil, err
	}
	valueToken, err := p.Expect([]TokenType{TokenString, TokenNumber}, "")
	if err != nil {
		return nil, err
	}
	return &storage.SearchCondition{
		ColumnName: columnNameToken.Value,
		Operator: operatorToken.Value,
		Value: []byte(valueToken.Value),
	}, nil
}

func ParseWhereExpression(p *Parser) (*storage.Expression, error) {
	root := &storage.Expression{}
	// Parse the first condition
	if p.Peek().Type == TokenLParen {
		p.Get() // consume '('
		subExpr, err := ParseWhereExpression(p)
		if err != nil {
			return nil, err
		}
		root = subExpr
		_, err = p.Expect([]TokenType{TokenRParen}, "")
		if err != nil {
			
			return nil, err
		}
	} else {
		condition, err := ParseWhereCondition(p)
		if err != nil {
			
			return nil, err
		}
		root.Type = storage.NodeCondition
		root.Condition = condition
	}
	lastNode := root

	for p.Peek().Type != TokenEOF && p.Peek().Type != TokenSemicolon && p.Peek().Type != TokenRParen {
		// Parse the logical operator (AND/OR)
		logicalOpToken, err := p.Expect([]TokenType{TokenKeyword}, "")
		if err != nil {	
			return nil, err
		}
		var nextExpr *storage.Expression
		// Check for parentheses to handle precedence
		if p.Peek().Type == TokenLParen {
			p.Get() // consume '('
			// Treat the expression inside parentheses as a new root
			subExpr, err := ParseWhereExpression(p)
			if err != nil {
				return nil, err
			}
			// Expect a closing parenthesis
			_, err = p.Expect([]TokenType{TokenRParen}, "")
			if err != nil {
				return nil, err
			}
			nextExpr = subExpr
		}else {
			// Read the next condition
			nextCondition, err := ParseWhereCondition(p)
			if err != nil {
				return nil, err
			}
			nextExpr = &storage.Expression{
				Type: storage.NodeCondition,
				Condition: nextCondition,
			}
		}

		// Attach the next condition to the last node based on the logical operator
		switch logicalOpToken.Value {
		case "AND":
			if lastNode.Type == storage.NodeCondition && lastNode.Condition != nil {
				lastNode.Type = storage.NodeAnd
				lastNode.Left = &storage.Expression{
					Type: storage.NodeCondition,
					Condition: lastNode.Condition,
				}
				lastNode.Right = nextExpr
				lastNode.Condition = nil
				lastNode = lastNode.Right
			}else {
				// If the last node was a sub-expression, we need to create a new AND node
				newRoot := &storage.Expression{
					Type: storage.NodeAnd,
					Left: lastNode,
					Right: nextExpr,
				}
				root = newRoot
				lastNode = root.Right
			}
		case "OR" :
			newRoot := &storage.Expression{
				Type: storage.NodeOr,
				Left: nextExpr,
				Right: root,
			}
			root = newRoot
			lastNode = root.Left
		default:
			return nil, fmt.Errorf("unexpected logical operator: %s", logicalOpToken.Value)		
		}				
	}
	return root, nil
	
}
func PrintExpression(expr *storage.Expression, indent string) {
	if expr == nil {
		return
	}
	switch expr.Type {
	case storage.NodeCondition:
		fmt.Printf("%sCondition: %s %s %s\n", indent, expr.Condition.ColumnName, expr.Condition.Operator, string(expr.Condition.Value))
	case storage.NodeAnd:
		fmt.Printf("%sAND:\n", indent)
		PrintExpression(expr.Left, indent+"  ")
		PrintExpression(expr.Right, indent+"  ")
	case storage.NodeOr:
		fmt.Printf("%sOR:\n", indent)
		PrintExpression(expr.Left, indent+"  ")
		PrintExpression(expr.Right, indent+"  ")
	default:
		fmt.Printf("%sUnknown expression type: %v\n", indent, expr.Type)
	}
}