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
				"key":  "DROP TABLE FOO",
				"2key": "/param first",
				"name": "1234",
			},
			want: "2key='%2Fparam%20first',key='DROP%20TABLE%20FOO',name='1234'",
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
