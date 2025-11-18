"""
Lab 4: Bible verse locator
Author: Alexander Taylor
Description: When the user inputs a book, chapter, and verse, the program retrieves
the corresponding verse from text file containing the entire Bible. In addition to
this, the program also allows the user to put in common abbreviations for book names,
and outputs the verse to a text file as well.
"""

import csv
import re


def main():
    # Store the abbreviations in a dictionary for easy reference
    with open("lab4/abbreviations.csv", newline="", encoding="utf-8") as abbrev_file:
        abbreviations = {
            row[0].strip().lower(): (row[1].strip().lower() if len(row) > 1 else "")
            for row in csv.reader(abbrev_file)
            if row and row[0].strip()
        }

    # Get input from the user and format it appropriately
    print("Please enter the reference of the verse you would like to retrieve")
    book = input("\tthe book: ")
    book_normalized = abbreviations.get(book.strip().lower(), book).lower()
    chapter = input("\tthe chapter: ")
    verse = input("\tthe verse: ")

    with open("lab4/bible.txt", newline="", encoding="utf-8") as bible_file:
        # Find the book
        book_match = re.search(
            f"^THE BOOK OF {book_normalized.upper()}.*$\n((?:(?!^THE BOOK OF).*\n?)*)",
            bible_file.read(),
            re.MULTILINE,
        )
        # Make sure that we found a book
        if not book_match:
            print(f'The Bible does not contain the book "{book}".')
            return 1
        book_text = book_match.group(1)

        # Find the chapter
        chapter_match = re.search(
            f"^(CHAPTER|PSALM) {chapter}.*$\n((?:(?!^(CHAPTER|PSALM) \\d+).*\n?)*)",
            book_text,
            re.MULTILINE,
        )
        # Make sure that we found a chapter
        if not chapter_match:
            print(f"The book of {book.capitalize()} does not have chapter {chapter}.")
            return 1
        chapter_text = chapter_match.group(2)

        # Find the verse
        verse_text = [
            verse_text.strip()
            for verse_text in (chapter_text.split("\n"))
            if verse_text and verse_text.strip().startswith(f"{verse} ")
        ]
        # Make sure that we found a verse
        if len(verse_text) < 1:
            print(
                f"{"Chapter" if book_normalized != "psalms" else "Psalm"} "
                f"{chapter} of {book.capitalize()} does not have verse {verse}."
            )
            return 1

        print("The verse you requested is:")
        to_print = f"{book.upper()} {chapter}:{verse_text[0]}".split(" ")
        # Format the output to be 80 characters or less per line
        output_text = to_print[0]
        for word in to_print[1:]:
            if len(output_text.split("\n")[-1]) + 1 + len(word) <= 80:
                output_text += " " + word
            else:
                output_text += "\n" + word

        # Print the text and output to a file.
        print(output_text)
        with open("lab4/verses.txt", "a", encoding="utf-8") as output_file:
            output_file.write(" ".join(to_print))
            output_file.write("\n")


if __name__ == "__main__":
    user_input = "y"
    while user_input.lower().startswith("y"):
        main()
        user_input = input("Would you like to look up another verse (y/N)? ")
