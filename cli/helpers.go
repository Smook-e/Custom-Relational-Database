package cli

import (
	"fmt"
	"strings"

	"github.com/Smook-e/Custom-Relational-Database/parser"
)

func handleMetaCommand(line string, qh *parser.QueryHandler) {
    switch {
    case line == "\\t" || line == ".tables":
        qh.PrintTables()

    case strings.HasPrefix(line, "\\d ") || strings.HasPrefix(line, ".schema "):
        parts := strings.Fields(line)
        if len(parts) < 2 {
            fmt.Println("usage: \\d <table_name>")
            return
        }
        
		err := qh.PrintTable(parts[1])
        if err != nil {
            fmt.Println(err)
            return
        }

    case line == "\\?" || line == "help":
        fmt.Println("\\t          list tables")
        fmt.Println("\\d <table>  describe table columns")
        fmt.Println("\\?          show this help")
        fmt.Println("\\q          quit")

    default:
        fmt.Println("unknown command:", line)
    }
}
func printSelectResult(columns []string, rows [][]any) {
    if len(rows) == 0 {
        fmt.Println("(0 rows)")
        return
    }

    // Compute each column's display width: max of header length and every value's length
    widths := make([]int, len(columns))
    for i, col := range columns {
        widths[i] = len(col)
    }
    strRows := make([][]string, len(rows))
    for r, row := range rows {
        strRows[r] = make([]string, len(row))
        for i, val := range row {
            s := formatValue(val)
            strRows[r][i] = s
            if len(s) > widths[i] {
                widths[i] = len(s)
            }
        }
    }

    printRow(columns, widths)
    printSeparator(widths)
    for _, row := range strRows {
        printRow(row, widths)
    }
    fmt.Printf("(%d rows)\n", len(rows))
}

func formatValue(val any) string {
    if val == nil {
        return "NULL"
    }
    return fmt.Sprintf("%v", val)
}

func printRow(values []string, widths []int) {
    for i, v := range values {
        fmt.Printf("| %-*s ", widths[i], v)
    }
    fmt.Println("|")
}

func printSeparator(widths []int) {
    for _, w := range widths {
        fmt.Print("+" + strings.Repeat("-", w+2))
    }
    fmt.Println("+")
}