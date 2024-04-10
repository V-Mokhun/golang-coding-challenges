package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

func main() {
	var byteCount int64
	var scanner *bufio.Scanner
	var fileName string
	arguments := []string{}

	if len(os.Args) > 1 {
		fileName = os.Args[len(os.Args)-1]
		arguments = os.Args[1 : len(os.Args)-1]
	}

	file, err := os.Open(fileName)
	if err != nil {
		stdinStat, _ := os.Stdin.Stat()
		byteCount = stdinStat.Size()
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		fileStat, _ := os.Stat(fileName)
		byteCount = fileStat.Size()
		scanner = bufio.NewScanner(file)
	}
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	lineCount, wordCount, charCount := 0, 0, 0
	for scanner.Scan() {
		line := scanner.Text()

		charCount += utf8.RuneCountInString(line) + 2
		for range strings.Fields(line) {
			wordCount++
		}

		lineCount++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Invalid input: %s", err)
	}

	if len(arguments) == 0 {
		fmt.Println(lineCount, wordCount, byteCount, fileName)
	} else {
		for _, arg := range arguments {
			switch arg {
			case "-c":
				fmt.Println(byteCount, fileName)
			case "-l":
				fmt.Println(lineCount, fileName)
			case "-w":
				fmt.Println(wordCount, fileName)
			case "-m":
				fmt.Println(charCount, fileName)
			}
		}
	}

}
