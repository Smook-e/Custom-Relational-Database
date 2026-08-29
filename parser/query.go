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
type DeleteQuery struct {
	TableName string
	Where *storage.Expression
}
type UpdateQuery struct {
	TableName string
	SetClauses map[string]string
	Where *storage.Expression
}
type CreateTableQuery struct {
	TableName string
	Columns []entities.ColumnDefinition
	ForeignKeys []entities.ForeignKeyDefinition
}
func (q *CreateTableQuery) Parse(p *Parser) ( error) {
	// Expect CREATE keyword
	_, err := p.Expect([]TokenType{TokenKeyword}, "CREATE")
	if err != nil {
		return  err
	}

	// Expect TABLE keyword
	_, err = p.Expect([]TokenType{TokenKeyword}, "TABLE")
	if err != nil {
		return  err
	}

	// Expect table name
	tableNameToken, err := p.Expect([]TokenType{TokenIdentifier}, "")
	if err != nil {
		return  err
	}
	q.TableName = tableNameToken.Value

	// Expect opening parenthesis for columns
	_, err = p.Expect([]TokenType{TokenLParen}, "")
	if err != nil {
		return  err
	}

	var columns []entities.ColumnDefinition
	var foreignKeys []entities.ForeignKeyDefinition

	for {
		if p.Peek().Type == TokenKeyword && p.Peek().Value == "FOREIGN" {
			p.Get() // Consume the FOREIGN keyword

			// Expect KEY keyword
			_, err = p.Expect([]TokenType{TokenKeyword}, "KEY")
			if err != nil {
				return  err
			}

			// Expect opening parenthesis for foreign key column
			_, err = p.Expect([]TokenType{TokenLParen}, "")
			if err != nil {
				return  err
			}

			// Expect foreign key column name
			fkColumnToken, err := p.Expect([]TokenType{TokenIdentifier}, "")
			if err != nil {
				return  err
			}

			// Expect closing parenthesis for foreign key column
			_, err = p.Expect([]TokenType{TokenRParen}, "")
			if err != nil {
				return  err
			}

			// Expect REFERENCES keyword
			_, err = p.Expect([]TokenType{TokenKeyword}, "REFERENCES")
			if err != nil {
				return  err
			}

			// Expect referenced table name
			refTableToken, err := p.Expect([]TokenType{TokenIdentifier}, "")
			if err != nil {
				return  err
			}

			// Expect opening parenthesis for referenced column
			_, err = p.Expect([]TokenType{TokenLParen}, "")
			if err != nil {
				return  err
			}

			// Expect referenced column name
			refColumnToken, err := p.Expect([]TokenType{TokenIdentifier}, "")
			if err != nil {
				return  err
			}

			// Expect closing parenthesis for referenced column
			_, err = p.Expect([]TokenType{TokenRParen}, "")
			if err != nil {
				return  err
			}

			foreignKeys = append(foreignKeys, entities.ForeignKeyDefinition{
				ColumnName: fkColumnToken.Value,
				ReferencedTableName: refTableToken.Value,
				ReferencedColumnName: refColumnToken.Value,
			})
		}else {
			// Parse column definition
			columnDef := entities.ColumnDefinition{}
			// Expect column name
			colNameToken, err := p.Expect([]TokenType{TokenIdentifier}, "")
			if err != nil {
				return  err
			}
			columnDef.Name = colNameToken.Value
			// Expect data type
			dataTypeToken, err := p.Expect([]TokenType{TokenKeyword}, "")
			if err != nil {
				return  err
			}
			columnDef.DataType = dataTypeToken.Value
			if dataTypeToken.Value == "VARCHAR" {
				// Expect opening parenthesis for VARCHAR length
				if p.Peek().Type == TokenLParen {
					p.Get() // Consume '('
					// Expect length value
					lengthToken, err := p.Expect([]TokenType{TokenNumber}, "")
					if err != nil {
						return  err
					}
					columnDef.DataType = fmt.Sprintf("VARCHAR(%s)", lengthToken.Value)
					// Expect closing parenthesis for VARCHAR length
					_, err = p.Expect([]TokenType{TokenRParen}, "")
					if err != nil {
						return  err
					}
				}
			}
			// Handle optional constraints 
			var constraints []string
			for p.position < len(p.tokens) && p.Peek().Type != TokenComma && p.Peek().Type != TokenRParen {
				// Expect constraint keyword
				constraintToken, err := p.Expect([]TokenType{TokenKeyword}, "")
				if err != nil {
					return  err
				}
				switch constraintToken.Value {
				case "NOT":
					// Expect NULL keyword
					_, err = p.Expect([]TokenType{TokenKeyword}, "NULL")
					if err != nil {
						return  err
					}
					constraints = append(constraints, "NOTNULL")
				case "PRIMARY":
					// Expect KEY keyword
					_, err = p.Expect([]TokenType{TokenKeyword}, "KEY")
					if err != nil {
						return  err
					}
					constraints = append(constraints, "PRIMARYKEY")
				case "DEFAULT":
					// Expect default value (string or number)
					defaultValueToken, err := p.Expect([]TokenType{TokenString, TokenNumber}, "")
					if err != nil {
						return  err
					}
					constraints = append(constraints, "DEFAULT")
					columnDef.Default = defaultValueToken.Value
				default:
					constraints = append(constraints, constraintToken.Value)
				}
				columnDef.Constraints = constraints
			}
			columns = append(columns, columnDef)
		}
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
	q.Columns = columns
	q.ForeignKeys = foreignKeys
	return  nil
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
	case *CreateTableQuery:
		fmt.Println("Create Table Query:")
		fmt.Println("Table Name:", query.TableName)
		fmt.Println("Columns:")
		for _, col := range query.Columns {
			fmt.Printf("  Name: %s, DataType: %s, Constraints: %v, Default: %s\n", col.Name, col.DataType, col.Constraints, col.Default)
		}
		fmt.Println("Foreign Keys:")
		for _, fk := range query.ForeignKeys {
			fmt.Printf("  Column: %s, References: %s(%s)\n", fk.ColumnName, fk.ReferencedTableName, fk.ReferencedColumnName)
		}
	default:
		fmt.Println("Unknown Query Type")
	}
}