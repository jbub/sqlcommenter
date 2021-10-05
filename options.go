package sqlcommenter

import (
	"context"
)

type Option func(cmt *commenter)

type AttrProvider interface {
	GetAttrs(context.Context) Attrs
}

type AttrProviderFunc func(context.Context) Attrs

func (f AttrProviderFunc) GetAttrs(ctx context.Context) Attrs {
	return f(ctx)
}

func WithAttrs(attrs Attrs) Option {
	return func(cmt *commenter) {
		cmt.providers = append(cmt.providers, AttrProviderFunc(func(ctx context.Context) Attrs {
			return attrs
		}))
	}
}

func WithAttrPairs(pairs ...string) Option {
	return func(cmt *commenter) {
		cmt.providers = append(cmt.providers, AttrProviderFunc(func(ctx context.Context) Attrs {
			return AttrPairs(pairs...)
		}))
	}
}

func WithAttrProvider(prov AttrProvider) Option {
	return func(cmt *commenter) {
		cmt.providers = append(cmt.providers, prov)
	}
}

func WithAttrFunc(fn AttrProviderFunc) Option {
	return func(cmt *commenter) {
		cmt.providers = append(cmt.providers, fn)
	}
}
