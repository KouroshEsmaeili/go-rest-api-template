package posts

import (
	"strings"
	"testing"
)

func TestInputValidate(t *testing.T) {
	tests := []struct {
		name  string
		input Input
		want  string
	}{
		{name: "valid", input: Input{Title: "A post"}},
		{name: "blank title", input: Input{Title: "  "}, want: "title is required"},
		{name: "long title", input: Input{Title: strings.Repeat("a", MaxTitleLength+1)}, want: "title must be 200 characters or fewer"},
		{name: "long content", input: Input{Title: "Title", Content: strings.Repeat("a", MaxContentLength+1)}, want: "content must be 10000 characters or fewer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.Validate(); got != tt.want {
				t.Fatalf("Validate() = %q, want %q", got, tt.want)
			}
		})
	}
}
