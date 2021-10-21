package sqlcommenter

import (
	"bytes"
	"testing"
)

func TestAttrsEncode(t *testing.T) {
	cases := []struct {
		name  string
		attrs Attrs
		want  string
	}{
		{
			name: "no attrs",
		},
		{
			name: "single attr",
			attrs: map[string]string{
				"key": "value",
			},
			want: "key='value'",
		},
		{
			name: "multiple attrs",
			attrs: map[string]string{
				"key":  "value",
				"2key": "value 33",
				"key3": "44  value",
			},
			want: "2key='value%2033',key='value',key3='44%20%20value'",
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			var b bytes.Buffer
			cs.attrs.encode(&b)
			got := b.String()
			if want := cs.want; want != got {
				t.Errorf("got '%v', want '%v'", got, want)
			}
		})
	}
}
