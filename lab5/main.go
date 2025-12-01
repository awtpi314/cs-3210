package main

// Alexander Taylor
// 11-18-2025
// CS 3210 Lab 5
// Bible Verse Locator
// This program allows users to search for Bible verses using fuzzy matching
// to find the book, and saves selected verses to a file.
//
// Additional Features:
// This program uses a menu-style interface in the terminal, allowing real-time
// input and dynamic updating of search results as the user types. It uses fuzzy
// matching to handle partial or misspelled book names.

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/sahilm/fuzzy"
	"golang.org/x/term"
)

// parseAbbreviations reads a CSV file containing abbreviations and their full names.
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
	// use built-in csv reader
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

// parseBible reads a text file containing the Bible and organizes it into a
// map structured by books, chapters, and verses.
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
		if after, ok := strings.CutPrefix(line, "THE BOOK OF "); ok {
			// We have a new book
			currentBook = after
			bible[currentBook] = [][]string{}
			currentChapter = -1
		} else if strings.HasPrefix(line, "CHAPTER ") ||
			// And now a new chapter
			strings.HasPrefix(line, "PSALM ") {
			parts := strings.SplitN(line, " ", 2)
			chapterNumStr := strings.TrimSpace(parts[1])
			chapterNum, err := strconv.Atoi(chapterNumStr)
			if err == nil {
				currentChapter = chapterNum - 1
			}
		} else {
			if currentBook != "" && currentChapter >= 0 {
				// It's a verse line
				verses := strings.Split(line, " ")
				if len(bible[currentBook]) <= currentChapter {
					bible[currentBook] = append(bible[currentBook], []string{})
				}
				if regexp.MustCompile(`^\d+$`).MatchString(verses[0]) {
					bible[currentBook][currentChapter] =
						append(bible[currentBook][currentChapter], strings.Join(verses[1:], " "))
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
	return bible
}

// normalizeUserInput processes user input to standardize book names and handle
// abbreviations and numeric prefixes.
func normalizeUserInput(input string, abbreviations map[string]string) string {
	normalizedString := strings.TrimSpace(input)
	if fullName, exists := abbreviations[normalizedString]; exists {
		return strings.ToUpper(fullName)
	} else {
		normalizedString = strings.ReplaceAll(normalizedString, "1", "FIRST")
		normalizedString = strings.ReplaceAll(normalizedString, "2", "SECOND")
		normalizedString = strings.ReplaceAll(normalizedString, "3", "THIRD")
	}
	return strings.ToUpper(normalizedString)
}

// getBibleBooks returns a slice of all book names in the Bible map.
func getBibleBooks(bible map[string][][]string) []string {
	books := make([]string, 0, len(bible))
	for book := range bible {
		books = append(books, book)
	}
	return books
}

// normalizeReference extracts chapter and verse numbers from a reference string.
func normalizeReference(reference string) (int, int) {
	reference = strings.TrimSpace(reference)
	// Split reference on punctuation (\p{P}) and whitespace (\s)
	verseParts := regexp.MustCompile(`[\p{P}\s]+`).Split(reference, -1)
	// There's a chapter number if there is a verse number
	chapterNum := regexp.MustCompile(`\d+`).FindString(verseParts[0])
	if len(verseParts) == 2 {
		// Both chapter and verse numbers are present
		verseNum := regexp.MustCompile(`\d+`).FindString(verseParts[1])
		chapter, _ := strconv.Atoi(chapterNum)
		verse, _ := strconv.Atoi(verseNum)
		if verse < 1 {
			verse = 1
		}
		return chapter, verse
	} else if len(verseParts) == 1 {
		// Only chapter number is present
		chapter, _ := strconv.Atoi(chapterNum)
		if chapter > 0 {
			return chapter, 1
		}
	}
	// Default to chapter 1, verse 1
	return 1, 1
}

// printWithWidth prints text to the terminal with line wrapping at the specified width.
func printWithWidth(text string, width int) {
	words := strings.Fields(text)
	currentLineLength := 0

	for _, word := range words {
		if currentLineLength+len(word)+1 > width {
			fmt.Print("\r\n")
			currentLineLength = 0
		}
		if currentLineLength > 0 {
			fmt.Print(" ")
			currentLineLength++
		}
		fmt.Print(word)
		currentLineLength += len(word)
	}
}

func main() {
	abbreviations := parseAbbreviations("abbreviations.csv")

	bible := parseBible("bible.txt")

	// Check that we are running in a terminal
	stdinFd := int(os.Stdin.Fd())
	if !term.IsTerminal(stdinFd) {
		fmt.Println("Please run the program in a terminal.")
		return
	}

	// Set terminal to raw mode for real-time input handling
	oldState, err := term.MakeRaw(stdinFd)
	if err != nil {
		fmt.Println("Error setting terminal to raw mode:", err)
		return
	}
	// Ensure terminal state is restored on exit
	defer func() {
		if err := term.Restore(stdinFd, oldState); err != nil {
			panic(err)
		}
	}()

	// Initial screen setup
	fmt.Print("\033[2J\033[H\r\n")
	defer fmt.Print("\033[2J\033[H\033[?25h")

	var search strings.Builder
	buf := make([]byte, 1)
	booksList := getBibleBooks(bible)
	currentVerse := ""
	previousSearch := ""
	fileSavedMessage := false

	// Draw initial static header
	fmt.Printf("Enter the reference (ctrl + c to quit)\r\n> ")

	for {
		currentSearch := search.String()

		if currentSearch != previousSearch {
			// Move cursor to input position and update search text
			fmt.Print("\033[?25l\033[3;3H")
			fmt.Print(currentSearch)
			fmt.Print("\033[K") // Clear to end of line

			// Clear from line 3 onwards for results
			fmt.Print("\033[4;1H")
			fmt.Print("\033[J\r\n") // Clear to end of screen

			// Ensure we can find the book based on user input
			upper := strings.ToUpper(currentSearch)
			bookMatchRaw := regexp.
				MustCompile(`^\s*(?:[1-3]\s*)?[A-Z]+(?:\s+[A-Z]+)*`).
				FindString(upper)
			bookToken := strings.TrimSpace(bookMatchRaw)
			bookSearch := normalizeUserInput(bookToken, abbreviations)

			// Everything after the matched book token is the reference (chapter/verse)
			referenceSearch := strings.TrimSpace(strings.TrimPrefix(upper, bookMatchRaw))

			if bookSearch != "" {
				// Fuzzy match the book name
				matches := fuzzy.Find(bookSearch, booksList)

				chapterNum, verseNum := normalizeReference(referenceSearch)
				if matches.Len() == 0 {
					fmt.Printf("\r\n%s does not exist in the Bible.\r\n", bookSearch)
				} else {
					if chapterNum > len(bible[matches[0].Str]) {
						// We have a book match, but invalid chapter
						fmt.Printf("\r\nChapter %d does not exist in %s\r\n",
							chapterNum, matches[0].Str)
					} else if verseNum > len(bible[matches[0].Str][chapterNum-1]) {
						// Valid book and chapter, but invalid verse
						fmt.Printf("\r\nVerse %d does not exist in %s %d\r\n",
							verseNum, matches[0].Str, chapterNum)
					} else {
						// Valid book, chapter, and verse
						verseText := bible[matches[0].Str][chapterNum-1][verseNum-1]
						currentVerse = fmt.Sprintf("%s %d:%d %s",
							matches[0].Str, chapterNum, verseNum, verseText)

						fmt.Printf("Pressing <Enter> will save %s %d:%d to the verses file.\r\n\r\n",
							matches[0].Str, chapterNum, verseNum)
						printWithWidth(currentVerse, 80)
					}
				}
			} else {
				currentVerse = ""
			}

			previousSearch = currentSearch
		}

		if fileSavedMessage {
			// Notify user that the verse was saved
			fmt.Print("Your verse saved to verses.txt!\033[J")
			fileSavedMessage = false
		}

		// Move cursor back to input position
		fmt.Printf("\033[3;%dH\033[?25h", 3+len(search.String()))
		os.Stdin.Read(buf)
		inputChar := buf[0]

		exitLoop := false
		switch {
		case inputChar == 3 || inputChar == 27:
			// Ctrl + C or Esc to exit
			exitLoop = true
		case inputChar == 127 || inputChar == 8:
			// Backspace/Delete
			if search.Len() > 0 {
				s := search.String()
				search.Reset()
				search.WriteString(s[:len(s)-1])
			}
		case inputChar >= 32 && inputChar <= 126:
			// Printable characters
			search.WriteByte(inputChar)
		case inputChar == 13:
			// Enter key
			if currentVerse != "" {
				// Append the current verse to verses.txt
				outputFile, err := os.OpenFile("verses.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Println("Error opening verses file:", err)
					return
				}
				defer outputFile.Close()

				if _, err := fmt.Fprintf(outputFile, "%s\r\n", currentVerse); err != nil {
					fmt.Println("Error writing to verses file:", err)
					return
				}

				fileSavedMessage = true
				// Clear search input and reset state
				search.Reset()
				currentVerse = ""
			}
		}
		if exitLoop {
			break
		}
	}
}
