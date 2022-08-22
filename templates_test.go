package main

import (
	"testing"

	"github.com/gofrs/uuid"
)

func TestUuidEq(t *testing.T) {
	tests := map[string]struct {
		s    string
		u    uuid.UUID
		want bool
	}{
		"equal": {
			s:    "05ff139b-8b9a-4341-a161-8628c3e038e7",
			u:    uuid.Must(uuid.FromString("05ff139b-8b9a-4341-a161-8628c3e038e7")),
			want: true,
		},
		"not equal": {
			s:    "05ff139b-8b9a-4341-a161-8628c3e038e7",
			u:    uuid.Must(uuid.FromString("2a0f3ad4-216d-43db-a731-fdab599e2d45")),
			want: false,
		},
		"empty": {
			s:    "",
			u:    uuid.Nil,
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := uuidEq(tc.s, tc.u)
			if tc.want != got {
				t.Errorf("want: %t, got: %t", tc.want, got)
			}
		})
	}
}
