package utils

import (
	"unicode/utf8"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// GetScore calculates the score of a guess based on the reference string,
// using the Levenshtein distance algorithm.
// The score is a float32 between 0 and 1
func GetScore(guess, reference string) float32 {
	guessLen := utf8.RuneCountInString(guess)
	titleLen := utf8.RuneCountInString(reference)
	score := fuzzy.LevenshteinDistance(guess, reference)
	return 100 * (float32(max(guessLen, titleLen)) - float32(score)) / float32(max(guessLen, titleLen))
}
