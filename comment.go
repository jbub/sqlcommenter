package sqlcommenter

import (
	"context"
	"strings"
)

const (
	commentStart = "/*"
	commentEnd   = "*/"
)

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
	return query + " " + commentStart + " " + attrs.encode() + " " + commentEnd
}

func (c *commenter) attrs(ctx context.Context) Attrs {
	if len(c.providers) == 0 {
		return nil
	}
	attrs := make(Attrs)
	for _, prov := range c.providers {
		attrs.Update(prov.GetAttrs(ctx))
	}
	return attrs
}
