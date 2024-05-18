package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Normalize sanitizes Spotigy songs names by removing diacritics, special
// characters and lowercasing the input. It also slice the stings to remove
// unimportant characters.
func Normalize(s string) string {
	// Runs a transformer to remove all diacritics and lowercase the input
	t := transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn)), norm.NFC, runes.Map(unicode.ToLower))
	normalized, _, _ := transform.String(t, s)

	// Cut the string after not important characters
	terminationStrings := []string{" (", " [", " - "}
	for _, r := range terminationStrings {
		normalized, _, _ = strings.Cut(normalized, r)
	}

	// Remove forbidden characters
	normalized = strings.NewReplacer(" - ", " ", " & ", " ", ".", " ", "!", " ", "remix", " ", "/", " ", "edit", " ", "from", " ").Replace(normalized)

	return normalized
}

func Permutations(input []string) []string {
	result := make([]string, 0, len(input)*(len(input)+1)/2)

	for length := 1; length <= len(input); length++ {
		for index := 0; index <= len(input)-length; index++ {
			result = append(result, strings.Join(input[index:index+length], " "))
		}
	}

	return result
}
