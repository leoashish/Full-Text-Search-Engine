package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	snowballeng "github.com/kljensen/snowball/english"
)

type document struct {
	Title string `xml:"title"`
	URL   string `xml:"url"`
	Text  string `xml:"abstract"`
	ID    int
}

func loadDocuments(path string) ([]document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	dec := xml.NewDecoder(f)

	dump := struct {
		Documents []document `xml:"doc"`
	}{}

	if err := dec.Decode(&dump); err != nil {
		return nil, err
	}

	docs := dump.Documents

	for i := range docs {
		docs[i].ID = i
	}
	return docs, nil

}

// Tokenizer
// The tokenizer is the first step of text analysis.
// Its job is to convert text into a list of tokens.
// Our implementation splits the text on a word boundary and removes punctuation marks:
func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

// Filteration
// Lowercase
func lowercaseFilters(tokens []string) []string {
	r := make([]string, len(tokens))

	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}

	return r
}

// Stopwords needs to be filtered
var stopwords = map[string]struct{}{ // I wish Go had built-in sets.
	"a": {}, "and": {}, "be": {}, "have": {}, "i": {},
	"in": {}, "of": {}, "that": {}, "the": {}, "to": {}}

func filterStopwords(tokens []string) []string {
	r := make([]string, 0, len(tokens))

	for _, token := range tokens {
		if _, ok := stopwords[token]; !ok {
			r = append(r, token)
		}
	}

	return r
}

// Stemming
// It involves converting the various forms of a word to a single form
func stemmerFilter(tokens []string) []string {
	r := make([]string, len(tokens))

	for i, token := range tokens {
		r[i] = snowballeng.Stem(token, false)
	}
	return r
}

// Building the index
type index map[string][]int

func (idx index) add(docs []document) {
	for _, doc := range docs {
		tokens := analyze(doc.Text)

		for _, token := range tokens {
			ids := idx[token]
			if ids != nil && ids[len(ids)-1] == doc.ID {
				continue
			}
			idx[token] = append(ids, doc.ID)
		}
	}
}

// Analyzer
func analyze(text string) []string {
	tokens := tokenize(text)
	tokens = lowercaseFilters(tokens)
	tokens = filterStopwords(tokens)
	tokens = stemmerFilter(tokens)

	return tokens
}

// Searching the content
// Attempt one
func search(docs []document, text string) []document {
	var r []document

	for _, doc := range docs {
		if strings.Contains(doc.Text, text) {
			r = append(r, doc)
		}
	}
	return r
}

// Intersection
func intersection(a []int, b []int) []int {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	r := make([]int, 0, maxLen)

	i := 0
	j := 0

	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			i++
		} else if b[j] < a[i] {
			j++
		} else {
			r = append(r, a[i])
			i++
			j++
		}
	}

	return r
}

// Searching using Regex
// Attempt two
func searchRegex(docs []document, term string) []document {
	re := regexp.MustCompile(`(?i)\b` + term + `\b`)

	var r []document
	for _, doc := range docs {
		if re.MatchString(doc.Text) {
			r = append(r, doc)
		}
	}

	return r
}

// Attempt three of search
func (idx index) search(text string) []int {
	var r []int

	tokens := analyze(text)
	for _, token := range tokens {
		if ids, ok := idx[token]; ok {
			if r == nil {
				r = ids
			} else {
				r = intersection(r, ids)
			}
		}
	}
	return r
}
func main() {
	docs, err := loadDocuments("enwiki-latest-abstract1.xml")

	if err != nil {
		panic(err)
	}

	idx := make(index)
	idx.add(docs)

	start := time.Now()
	result := idx.search("Small wild cat")

	elapsed := time.Since(start)
	fmt.Println(elapsed)

	fmt.Println(result)

	for _, r := range result {
		fmt.Println(docs[r].ID, " ", docs[r].Text)
	}

	result = idx.search("Catopuma")
	for _, r := range result {
		fmt.Println(docs[r].ID, " ", docs[r].Text)
	}

}
