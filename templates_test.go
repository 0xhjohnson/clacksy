package main

import (
	"testing"
	"time"

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

func TestHumanDate(t *testing.T) {
	tests := map[string]struct {
		time time.Time
		want string
	}{
		"January 2, 2021 at 1:04pm": {
			time: time.Date(2021, time.January, 2, 13, 4, 0, 0, time.UTC),
			want: "02 Jan 2021 at 1:04PM",
		},
		"December 30, 2020 at 11:59pm": {
			time: time.Date(2020, time.December, 30, 23, 59, 0, 0, time.UTC),
			want: "30 Dec 2020 at 11:59PM",
		},
		"February 14, 2022 at 2:30AM": {
			time: time.Date(2022, time.February, 14, 2, 30, 0, 0, time.UTC),
			want: "14 Feb 2022 at 2:30AM",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := humanDate(tc.time)
			if tc.want != got {
				t.Errorf("want: %v, got %v", tc.want, got)
			}
		})
	}
}
