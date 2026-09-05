package cli
import (
	"github.com/Smook-e/Custom-Relational-Database/parser"
	"fmt"
	"bufio"
	"strings"
	"os"
)


func MainLoop() {
	qh, err := parser.InitializeQueryHandler("database.db")
	if err != nil {
		fmt.Printf("Error: Failed to initialize query handler: %v\n", err)
		return
	}
	defer qh.Close()
	scanner := bufio.NewScanner(os.Stdin)
	var buffer []string

	fmt.Println("SQL Engine initialized. Enter your SQL commands below.\nType '\\q' to exit.\nType '\\w' or '.commit' to write changes to disk.\nType '\\h' for help.")
	for {
		if len(buffer) == 0 {
			fmt.Print("> ")
		} else {
			fmt.Print("-> ")
		}
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}
		if trimmedLine == "exit" || trimmedLine == "quit" {
            break
        }
		if strings.HasPrefix(trimmedLine, "\\") || strings.HasPrefix(trimmedLine, ".") {
			quit := strings.HasPrefix(trimmedLine, "\\q") || strings.HasPrefix(trimmedLine, ".quit")
			if quit {
				break
			}
			handleMetaCommand(trimmedLine, qh)
			continue
		}
		buffer = append(buffer, line)
		// Execute the query if the line ends with a semicolon
		if strings.HasSuffix(trimmedLine, ";") {
			fullQuery := strings.Join(buffer, "\n")
			fullQuery = strings.TrimSuffix(fullQuery, ";")
			
			buffer = buffer[:0] // Clear the buffer
			qr, err := qh.ExecuteQuery(fullQuery)
			if err != nil {
				fmt.Println(err)
				continue
			}
			printQueryResult(qr)
		}
		
	}
}