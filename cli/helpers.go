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