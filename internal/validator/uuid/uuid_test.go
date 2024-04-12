package uuid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		name string
		u    string
		want bool
	}{
		{
			name: "Valid UUID",
			u:    "3b241101-e2bb-4255-8caf-4136c566a964",
			want: true,
		},
		{
			name: "Invalid UUID",
			u:    "invalid-uuid",
			want: false,
		},
		{
			name: "Empty string",
			u:    "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, IsValid(tt.u) == tt.want)
		})
	}
}
