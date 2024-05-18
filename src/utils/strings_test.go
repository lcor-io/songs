package utils_test

import (
	"reflect"
	"testing"

	"lcor.io/songs/src/utils"
)

func TestNormalize(t *testing.T) {
	testcases := []struct {
		in, want string
	}{
		{"Bonjour", "bonjour"},
		{"Hello World", "hello world"},
		{"Hello World (Remix)", "hello world"},
		{"Hello World [Remix]", "hello world"},
		{"Hello World - Remix", "hello world"},
		{"Hello World & Friends", "hello world friends"},
		{"Héllò Wo̧rld - Remix", "hello world"},
		{"Héllò - Wo̧rld - Remix", "hello"},
	}

	for _, tc := range testcases {
		normalized := utils.Normalize(tc.in)
		if normalized != tc.want {
			t.Errorf("Normalize(%q) = %q; want %q", tc.in, normalized, tc.want)
		}
	}
}

func TestPermutations(t *testing.T) {
	testscases := []struct {
		in, want []string
	}{
		{[]string{"Hello", "world"}, []string{"Hello", "world", "Hello world"}},
		{[]string{"Hello"}, []string{"Hello"}},
	}

	for _, tc := range testscases {
		permutations := utils.Permutations(tc.in)
		if !reflect.DeepEqual(permutations, tc.want) {
			t.Errorf("Permutations(%q) = %q; want %q", tc.in, permutations, tc.want)
		}
	}
}
