package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/term"
)

func parseAbbreviations(filename string) map[string]string {
	abbreviations := make(map[string]string)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return abbreviations
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()
	csvReader := csv.NewReader(bufio.NewReader(file))
	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV file:", err)
		return abbreviations
	}
	for _, record := range records {
		if len(record) >= 2 {
			abbreviation := strings.TrimSpace(record[0])
			fullName := strings.TrimSpace(record[1])
			abbreviations[abbreviation] = fullName
		}
	}
	return abbreviations
}

func parseBible(filename string) map[string][][]string {
	bible := make(map[string][][]string)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return bible
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()
	scanner := bufio.NewScanner(file)
	currentBook := ""
	currentChapter := -1
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "THE BOOK OF ") {
			currentBook = strings.TrimPrefix(line, "THE BOOK OF ")
			bible[currentBook] = [][]string{}
			currentChapter = -1
		} else if strings.HasPrefix(line, "CHAPTER ") || strings.HasPrefix(line, "PSALM ") {
			parts := strings.SplitN(line, " ", 2)
			chapterNumStr := strings.TrimSpace(parts[1])
			chapterNum, err := strconv.Atoi(chapterNumStr)
			if err == nil {
				currentChapter = chapterNum - 1
			}
		} else {
			if currentBook != "" && currentChapter >= 0 {
				verses := strings.Split(line, " ")
				if len(bible[currentBook]) <= currentChapter {
					bible[currentBook] = append(bible[currentBook], []string{})
				}
				bible[currentBook][currentChapter] = append(bible[currentBook][currentChapter], strings.Join(verses, " "))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
	return bible
}

func normalizeUserInput(input string, abbreviations map[string]string) string {
	normalizedString := strings.TrimSpace(input)
	if fullName, exists := abbreviations[normalizedString]; exists {
		return strings.ToUpper(fullName)
	} else {
		normalizedString = strings.Replace(normalizedString, "1", "FIRST", -1)
		normalizedString = strings.Replace(normalizedString, "2", "SECOND", -1)
		normalizedString = strings.Replace(normalizedString, "3", "THIRD", -1)
	}
	return strings.ToUpper(normalizedString)
}

func getBibleBooks(bible map[string][][]string) []string {
	books := make([]string, 0, len(bible))
	for book := range bible {
		books = append(books, book)
	}
	return books
}

func main() {
	abbreviations := parseAbbreviations("abbreviations.csv")

	bible := parseBible("bible.txt")

	stdinFd := int(os.Stdin.Fd())
	if !term.IsTerminal(stdinFd) {
		fmt.Println("Please run the program in a terminal.")
		return
	}

	oldState, err := term.MakeRaw(stdinFd)
	if err != nil {
		fmt.Println("Error setting terminal to raw mode:", err)
		return
	}
	defer func() {
		if err := term.Restore(stdinFd, oldState); err != nil {
			panic(err)
		}
	}()

	_, height, err := term.GetSize(stdinFd)
	if err != nil {
		fmt.Println("Error getting terminal height:", err)
		fmt.Println("Defaulting to a height of 10.")
		height = 10
	}

	fmt.Print("\033[2J\033[H\033[?25l")
	defer fmt.Print("\033[?25h")

	var search strings.Builder
	buf := make([]byte, 1)
	booksList := getBibleBooks(bible)

	for {
		fmt.Print("\033[H\033[2J")

		fmt.Printf("Enter the reference (ctrl + c to quit)\r\n> %s\r\n", search.String())
		bookSearch := normalizeUserInput(search.String(), abbreviations)
		bookSearch = regexp.MustCompile(`^[^0-9]*`).FindString(bookSearch)

		if bookSearch != "" {
			matches := fuzzy.Find(bookSearch, booksList)

			maxResults := max(height-4, 0)
			if len(matches) > maxResults {
				matches = matches[:maxResults]
			}

			for _, match := range matches {
				fmt.Printf("  - %s\r\n", match)
			}
		}

		os.Stdin.Read(buf)
		inputChar := buf[0]

		if inputChar == 3 {
			break
		} else if inputChar == 127 || inputChar == 8 {
			if search.Len() > 0 {
				s := search.String()
				search.Reset()
				search.WriteString(s[:len(s)-1])
			}
		} else if inputChar >= 32 && inputChar <= 126 {
			search.WriteByte(inputChar)
		}
	}
}
