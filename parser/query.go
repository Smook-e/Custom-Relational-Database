package parser


import (
	"fmt"
	"github.com/Smook-e/Custom-Relational-Database/storage"
	"github.com/Smook-e/Custom-Relational-Database/entities"
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
type CreateTableQuery struct {
	TableName string
	Columns []entities.ColumnDefinition
	ForeignKeys []entities.ForeignKeyDefinition
}
func (q *CreateTableQuery) Parse(p *Parser) ( error) {
	return nil
}


func (q *SelectQuery) Parse(p *Parser) ( error) {
	// Expect SELECT keyword
	_, err := p.Expect([]TokenType{TokenKeyword}, "SELECT")
	if err != nil {
		return  err
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
				return  err
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
		return  err
	}

	tableNameToken, err := p.Expect([]TokenType{TokenIdentifier}, "")
	if err != nil {
		return  err
	}
	// query := &SelectQuery{
	// 	Columns: columns,
	// 	TableName: tableNameToken.Value,
	// }
	q.Columns = columns
	q.TableName = tableNameToken.Value
	if p.Peek().Type == TokenKeyword && p.Peek().Value == "WHERE" {
		p.Get() // Consume the WHERE keyword
		whereExpr, err := ParseWhereExpression(p)
		if err != nil {
			return  err
		}
		q.Where = whereExpr
	}
	return  nil
}

func (q *InsertQuery) Parse(p *Parser) ( error) {
	// Expect INSERT keyword
	_, err := p.Expect([]TokenType{TokenKeyword}, "INSERT")
	if err != nil {
		return  err
	}

	// Expect INTO keyword
	_, err = p.Expect([]TokenType{TokenKeyword}, "INTO")
	if err != nil {
		return  err
	}

	// Expect table name
	tableNameToken, err := p.Expect([]TokenType{TokenIdentifier}, "")
	if err != nil {
		return  err
	}

	// Expect opening parenthesis for columns
	_, err = p.Expect([]TokenType{TokenLParen}, "")
	if err != nil {
		return  err
	}

	var columns []string
	for {
		col, err := p.Expect([]TokenType{TokenIdentifier}, "")
		if err != nil {
			return  err
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
		return  err
	}

	// Expect VALUES keyword
	_, err = p.Expect([]TokenType{TokenKeyword}, "VALUES")
	if err != nil {
		return  err
	}
	var values [][]string

	for p.Peek().Type != TokenSemicolon && p.Peek().Type != TokenEOF && p.position < len(p.tokens) {
		// Expect opening parenthesis for values
		_, err = p.Expect([]TokenType{TokenLParen}, "")
		if err != nil {
			return  err
		}
		var row []string
		for p.Peek().Type != TokenRParen {
			val, err := p.Expect([]TokenType{TokenString, TokenNumber}, "")
			if err != nil {
				return  err
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
			return  err
		}
		if p.Peek().Type == TokenComma {
			p.Get() // Consume the comma
		} else {
			break
		}
	}
	q.TableName = tableNameToken.Value
	q.Columns = columns
	q.Values = values
	return  nil
}
func Print(q Query) {
	switch query := q.(type) {
	case *SelectQuery:
		fmt.Println("Select Query:")
		fmt.Println("Columns:", query.Columns)
		fmt.Println("Table Name:", query.TableName)
		if query.Where != nil {
			fmt.Println("Where Condition:")
			PrintExpression(query.Where, "")
		}
	case *InsertQuery:
		fmt.Println("Insert Query:")
		fmt.Println("Table Name:", query.TableName)
		fmt.Println("Columns:", query.Columns)
		fmt.Println("Values:", query.Values)
	default:
		fmt.Println("Unknown Query Type")
	}
}