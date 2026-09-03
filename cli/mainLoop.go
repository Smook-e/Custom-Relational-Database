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

	fmt.Println("SQL Engine initialized. Enter your queries (type 'exit' to quit):")
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "exit" || line == "quit" {
            break
        }
        if line == "" {
            continue
        }
		_ , err := qh.ExecuteQuery(line)
		if err != nil {
			fmt.Printf("Error executing query: %v\n", err)
			continue
		}
		
	}
}