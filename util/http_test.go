package util

import "testing"

func TestReplacePlaceholdersWithQueryParams(t *testing.T) {
	vars := map[string]string{
		"BASE_URL": "https://api.example.com",
		"CAT":      "electronics",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "{BASE_URL}/products/search?x=x",
			expected: "https://api.example.com/products/search?x=x",
		},
		{
			input:    "{BASE_URL}/products/search?category={CAT}&limit=10",
			expected: "https://api.example.com/products/search?category=electronics&limit=10",
		},
	}

	for _, tt := range tests {
		result, err := replacePlaceholders(tt.input, vars)
		if err != nil {
			t.Errorf("replacePlaceholders(%q) error: %v", tt.input, err)
		}
		if result != tt.expected {
			t.Errorf("replacePlaceholders(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
