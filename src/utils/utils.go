package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

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
	normalized = strings.NewReplacer(" - ", "", " & ", "", ".", "", "!", "", "remix", "", "/", "", "edit", "", "from", "").Replace(normalized)

	return normalized
}
