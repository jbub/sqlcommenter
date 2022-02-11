package sqlcommenter

import (
	"bytes"
)

// AttrPairs builds Attrs from multiple key value pairs.
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

// Attr represents attribute with key and value.
type Attr struct {
	Key   string
	Value string
}

// Attrs wrap map of attributes.
type Attrs map[string]string

// Update updates map from other Attrs.
func (a Attrs) Update(other Attrs) {
	for k, v := range other {
		a[k] = v
	}
}

func (a Attrs) encode(b *bytes.Buffer) {
	total := len(a)
	keys := make([]string, 0, total)
	for k := range a {
		keys = append(keys, k)
	}
	sortKeys(keys)

	for i, key := range keys {
		writeQueryEscape(key, b)

		b.WriteByte('=')
		b.WriteByte('\'')

		writePathEscape(a[key], b)

		b.WriteByte('\'')
		if i < total-1 {
			b.WriteByte(',')
		}
	}
}

// sortKeys implements a simple insertion sort on string slice.
// We save one alloc by not using sort.Strings.
func sortKeys(keys []string) {
	for i := 1; i < len(keys); i++ {
		if keys[i] < keys[i-1] {
			j := i - 1
			temp := keys[i]
			for j >= 0 && keys[j] > temp {
				keys[j+1] = keys[j]
				j--
			}
			keys[j+1] = temp
		}
	}
}
