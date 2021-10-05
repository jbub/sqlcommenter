package sqlcommenter

import (
	"net/url"
	"sort"
	"strings"
)

func AttrPairs(pairs ...string) Attrs {
	if len(pairs)%2 == 1 {
		panic("got odd number of pairs")
	}
	attrs := make(Attrs, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		attrs[pairs[i]] = pairs[i+1]
	}
	return attrs
}

type Attr struct {
	Key   string
	Value string
}

type Attrs map[string]string

func (a Attrs) Encode() string {
	total := len(a)
	keys := make([]string, 0, total)
	for k := range a {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, key := range keys {
		b.WriteString(encodeKey(key))
		b.WriteByte('=')
		b.WriteString(encodeValue(a[key]))
		if i < total-1 {
			b.WriteByte(',')
		}
	}
	return b.String()
}

func (a Attrs) Update(other Attrs) {
	for k, v := range other {
		a[k] = v
	}
}

func encodeKey(k string) string {
	return url.QueryEscape(k)
}

func encodeValue(v string) string {
	return "'" + strings.ReplaceAll(url.PathEscape(v), "+", "%20") + "'"
}
