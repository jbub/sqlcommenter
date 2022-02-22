package sqlcommenter

import (
	"bytes"
	"context"
	"strings"
	"sync"
)

const (
	commentStart = "/*"
	commentEnd   = "*/"
)

// Comment adds comments to query using provided options.
func Comment(ctx context.Context, query string, opts ...Option) string {
	if len(opts) == 0 {
		return query
	}
	if strings.Contains(query, commentStart) {
		return query
	}
	return newCommenter(opts...).comment(ctx, query)
}

func newCommenter(opts ...Option) *commenter {
	cmt := &commenter{}
	for _, opt := range opts {
		opt(cmt)
	}
	return cmt
}

type commenter struct {
	providers []AttrProvider
}

func (c *commenter) comment(ctx context.Context, query string) string {
	attrs := c.attrs(ctx)
	if len(attrs) == 0 {
		return query
	}

	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()

	buf.WriteString(query)
	buf.WriteByte(' ')
	buf.WriteString(commentStart)
	attrs.encode(buf)
	buf.WriteString(commentEnd)
	return buf.String()
}

func (c *commenter) attrs(ctx context.Context) Attrs {
	switch len(c.providers) {
	case 0:
		return nil
	case 1:
		return c.providers[0].GetAttrs(ctx)
	default:
		attrs := make(Attrs)
		for _, prov := range c.providers {
			attrs.Update(prov.GetAttrs(ctx))
		}
		return attrs
	}
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 100))
	},
}
