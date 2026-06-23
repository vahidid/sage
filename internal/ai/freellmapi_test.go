package ai

import "testing"

func TestFreeLLMAPIResponsesURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{
			name:    "base host",
			baseURL: "http://65.109.176.81:3001",
			want:    "http://65.109.176.81:3001/v1/responses",
		},
		{
			name:    "v1 base",
			baseURL: "http://65.109.176.81:3001/v1",
			want:    "http://65.109.176.81:3001/v1/responses",
		},
		{
			name:    "full endpoint",
			baseURL: "http://65.109.176.81:3001/v1/responses",
			want:    "http://65.109.176.81:3001/v1/responses",
		},
		{
			name:    "trailing slash",
			baseURL: "http://65.109.176.81:3001/",
			want:    "http://65.109.176.81:3001/v1/responses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := freeLLMAPIResponsesURL(tt.baseURL); got != tt.want {
				t.Fatalf("freeLLMAPIResponsesURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
