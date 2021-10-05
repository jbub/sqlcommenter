package sqlcommenter

import (
	"testing"
)

func TestComment(t *testing.T) {
	cases := []struct {
		name  string
		query string
		opts  []Option
		want  string
	}{
		{
			name: "empty query",
		},
		{
			name:  "empty query with whitespace",
			query: "  ",
			want:  "  ",
		},
		{
			name:  "query with comment",
			query: "SELECT 1  /* comment */",
			want:  "SELECT 1  /* comment */",
		},
		{
			name:  "query without attrs",
			query: "SELECT 1",
			want:  "SELECT 1",
		},
		{
			name:  "query with single attr",
			query: "SELECT 1",
			opts:  []Option{WithAttrPairs("key", "value")},
			want:  "SELECT 1 /* key='value' */",
		},
		{
			name:  "query with multiple attrs",
			query: "SELECT 1",
			opts:  []Option{WithAttrPairs("key", "1value", "key2", "  value 2")},
			want:  "SELECT 1 /* key='1value',key2='%20%20value%202' */",
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			got := Comment(cs.query, cs.opts...)
			if want := cs.want; want != got {
				t.Fatalf("got '%v', want '%v'", got, want)
			}
		})
	}
}
