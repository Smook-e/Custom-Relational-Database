package parser


import (
	// "fmt"
	"github.com/Smook-e/Custom-Relational-Database/storage"
)

type SelectQuery struct {
	Columns []string
	TableName string
	Where *storage.Expression
}
type InsertQuery struct {
	TableName string
	Columns []string
	Values [][]string
}

func ParseSelectQuery(p *Parser) (*SelectQuery, error) {
	// Expect SELECT keyword
	_, err := p.Expect([]TokenType{TokenKeyword}, "SELECT")
	if err != nil {
		return nil, err
	}

	// Parse columns
	var columns []string

	if p.Peek().Value == "*" {
		columns = append(columns, "*")
		p.Get() // Consume the '*'
	} else {
		for  {
			col , err := p.Expect([]TokenType{TokenIdentifier}, "")
			if err != nil {
				return nil, err
			}
			columns = append(columns, col.Value)

			if p.Peek().Type == TokenComma {
				p.Get() // Consume the comma
			} else {
				break
			}
		}
	}
	if _, err := p.Expect([]TokenType{TokenKeyword}, "FROM"); err != nil {
		return nil, err
	}

	tableNameToken, err := p.Expect([]TokenType{TokenIdentifier}, "")
	if err != nil {
		return nil, err
	}
	query := &SelectQuery{
		Columns: columns,
		TableName: tableNameToken.Value,
	}
	if p.Peek().Type == TokenKeyword && p.Peek().Value == "WHERE" {
		p.Get() // Consume the WHERE keyword
		whereExpr, err := ParseWhereExpression(p)
		if err != nil {
			return nil, err
		}
		query.Where = whereExpr
	}
	return query, nil
}

func ParseInsertQuery(p *Parser) (*InsertQuery, error) {
	// Expect INSERT keyword
	_, err := p.Expect([]TokenType{TokenKeyword}, "INSERT")
	if err != nil {
		return nil, err
	}

	// Expect INTO keyword
	_, err = p.Expect([]TokenType{TokenKeyword}, "INTO")
	if err != nil {
		return nil, err
	}

	// Expect table name
	tableNameToken, err := p.Expect([]TokenType{TokenIdentifier}, "")
	if err != nil {
		return nil, err
	}

	// Expect opening parenthesis for columns
	_, err = p.Expect([]TokenType{TokenLParen}, "")
	if err != nil {
		return nil, err
	}

	var columns []string
	for {
		col, err := p.Expect([]TokenType{TokenIdentifier}, "")
		if err != nil {
			return nil, err
		}
		columns = append(columns, col.Value)

		if p.Peek().Type == TokenComma {
			p.Get() // Consume the comma
		} else {
			break
		}
	}

	// Expect closing parenthesis for columns
	_, err = p.Expect([]TokenType{TokenRParen}, "")
	if err != nil {
		return nil, err
	}

	// Expect VALUES keyword
	_, err = p.Expect([]TokenType{TokenKeyword}, "VALUES")
	if err != nil {
		return nil, err
	}
	var values [][]string

	for p.Peek().Type != TokenSemicolon && p.Peek().Type != TokenEOF {
		// Expect opening parenthesis for values
		_, err = p.Expect([]TokenType{TokenLParen}, "")
		if err != nil {
			return nil, err
		}
		var row []string
		for p.Peek().Type != TokenRParen {
			val, err := p.Expect([]TokenType{TokenString, TokenNumber}, "")
			if err != nil {
				return nil, err
			}
			row = append(row, val.Value)
			if p.Peek().Type == TokenComma {
				p.Get() // Consume the comma
			} else {
				break
			}
		}
		values = append(values, row)
		// Expect closing parenthesis for values
		if _, err = p.Expect([]TokenType{TokenRParen}, ""); err != nil {
			return nil, err
		}
		if p.Peek().Type == TokenComma {
			p.Get() // Consume the comma
		} else {
			break
		}
	}
	return &InsertQuery{
		TableName: tableNameToken.Value,
		Columns:   columns,
		Values:    values,
	}, nil
}