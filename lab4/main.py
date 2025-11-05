"""
Lab 4: Bible verse locator
Author: Alexander Taylor
Description: When the user inputs a book, chapter, and verse, the program retrieves
the corresponding verse from text file containing the entire Bible. In addition to
this, the program also allows the user to put in common abbreviations for book names,
and outputs the verse to a text file as well.
"""

import csv


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
    book_normalized = abbreviations.get(book.strip().lower(), book)
    if book_normalized == "psalm":
        book_normalized = "psalms"
    chapter = input("\tthe chapter: ")
    verse = input("\tthe verse: ")

    with open("lab4/bible.txt", newline="", encoding="utf-8") as bible_file:
        # Find the book
        book_text = [
            book_text.strip()
            for book_text in bible_file.read().split("THE BOOK OF")
            if book_text
            and book_text.strip().lower().startswith(book_normalized)
        ]
        # Make sure that we found a book
        if len(book_text) < 1:
            print(f'The Bible does not contain the book "{book}".')
            return 1
        book_text = book_text[0]

        # Find the chapter
        chapter_text = [
            chapter_text.strip()
            for chapter_text in (
                book_text.split("CHAPTER")
                if book_normalized != "psalms"
                else book_text.split("PSALM")
            )
            if chapter_text
            and chapter_text.strip().startswith(f"{chapter}\n")
        ]
        # Make sure that we found a chapter
        if len(chapter_text) < 1:
            print(f"The book of {book.capitalize()} does not have chapter {chapter}.")
            return 1

        # Find the verse
        verse_text = [
            verse_text.strip()
            for verse_text in (chapter_text[0].split("\n"))
            if verse_text
            and verse_text.strip().startswith(f"{verse} ")
        ]
        # Make sure that we found a verse
        if len(verse_text) < 1:
            print(
                f"{"Chapter" if book_normalized != "psalms" else "Psalm"} "
                f"{chapter} of {book.capitalize()} does not have verse {verse}."
            )
            return 1

        print("The verse you requested is:")
        to_print = f"{book.capitalize()} {chapter}:{verse_text[0]}".split(" ")
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


if __name__ == "__main__":
    user_input = "y"
    while user_input.lower().startswith("y"):
        main()
        user_input = input("Would you like to look up another verse (y/N)? ")
