package reader

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type InputReader struct{}

func NewInputReader() *InputReader {
	return &InputReader{}
}

func (r *InputReader) ReadCommand() (string, bool) {
	fmt.Printf("minishell> ")

	reader := bufio.NewReader(os.Stdin)

	line, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			fmt.Println()
			return "", false
		}
		return "", false
	}

	line = strings.TrimSpace(line)
	return line, true
}

