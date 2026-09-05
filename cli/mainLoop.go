package cli
import (
	"github.com/Smook-e/Custom-Relational-Database/parser"
	"fmt"
	"strings"
	"os"
	"golang.org/x/term"
)


func MainLoop() {
	qh, err := parser.InitializeQueryHandler("database.bin")
	if err != nil {
		fmt.Printf("Error: Failed to initialize query handler: %v\n", err)
		return
	}
	defer qh.Close()
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("Error setting raw mode: %v\n", err)
		return
	}
	// Crucial: Restore original terminal settings when the program exits!
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// 2. Use term.Terminal to handle line editing, cursor movement, and pasting natively
	t := term.NewTerminal(os.Stdin, "\r\n> ")
	t.SetSize(500, 24)
	var buffer []string

	fmt.Print("SQL Engine initialized. Enter your SQL commands below.\r\nType '\\q' to exit.\r\nType '\\w' or '.commit' to write changes to disk.\r\nType '\\h' for help.", "\r\n")
	for {
		
		line, err := t.ReadLine()
		if err != nil {// Ctrl + c
			break 
		}
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
			
			buffer = buffer[:0] // Clear the buffer
			qr, err := qh.ExecuteQuery(fullQuery)
			if err != nil {
				fmt.Print(err, "\r\n")
				continue
			}
			printQueryResult(qr)
			t.SetPrompt("\r\n> ") 
		}else {
			// Change prompt to continuation prompt for multi-line inputs
			t.SetPrompt("-> ")
		}
		
	}
}