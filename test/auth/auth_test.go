package auth_test

import (
	"net/http"
	"testing"

	"github.com/sacortesro/a-que-voto/internal/auth"
)

func TestGetAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{name: "missing header", header: "", wantErr: true},
		{name: "wrong scheme", header: "ApiKey abc", wantErr: true},
		{name: "scheme only", header: "Bearer", wantErr: true},
		{name: "too many parts", header: "Bearer abc def", wantErr: true},
		{name: "valid", header: "Bearer abc", want: "abc"},
		{name: "lowercase scheme", header: "bearer abc", want: "abc"},
		{name: "extra spaces", header: "Bearer   abc ", want: "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			if tt.header != "" {
				h.Set("Authorization", tt.header)
			}

			got, err := auth.GetAPIKey(h)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got key %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
