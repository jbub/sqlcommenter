package sqlcommenter

import (
	"context"
	"sync"
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
			got := Comment(context.Background(), cs.query, cs.opts...)
			if want := cs.want; want != got {
				t.Fatalf("got '%v', want '%v'", got, want)
			}
		})
	}
}

func TestCommentConcurrent(t *testing.T) {
	var wg sync.WaitGroup

	ctx := context.Background()
	cmt := newCommenter(WithAttrs(map[string]string{
		"key":  "value",
		"2key": "value 33",
		"key3": "44  value",
	}))

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmt.comment(ctx, "SELECT * FROM my_table WHERE column IS NOT NULL")
		}()
	}

	wg.Wait()
}

func BenchmarkComment(b *testing.B) {
	ctx := context.Background()
	cmt := newCommenter(WithAttrs(map[string]string{
		"key":  "value",
		"2key": "value 33",
		"key3": "44  value",
	}))

	b.ReportAllocs()
	b.SetBytes(2)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmt.comment(ctx, "SELECT * FROM my_table WHERE column IS NOT NULL")
	}
}
