package sqlcommenter

import (
	"context"
)

// Option configures commenter.
type Option func(cmt *commenter)

// AttrProvider provides Attrs from context.Context.
type AttrProvider interface {
	GetAttrs(context.Context) Attrs
}

// AttrProviderFunc adapts func to AttrProvider.
type AttrProviderFunc func(context.Context) Attrs

// GetAttrs returns Attrs.
func (f AttrProviderFunc) GetAttrs(ctx context.Context) Attrs {
	return f(ctx)
}

// WithAttrs configures commenter with Attrs.
func WithAttrs(attrs Attrs) Option {
	return func(cmt *commenter) {
		cmt.providers = append(cmt.providers, AttrProviderFunc(func(ctx context.Context) Attrs {
			return attrs
		}))
	}
}

// WithAttrPairs configures commenter with attr pairs.
func WithAttrPairs(pairs ...string) Option {
	return func(cmt *commenter) {
		cmt.providers = append(cmt.providers, AttrProviderFunc(func(ctx context.Context) Attrs {
			return AttrPairs(pairs...)
		}))
	}
}

// WithAttrProvider configures commenter with AttrProvider.
func WithAttrProvider(prov AttrProvider) Option {
	return func(cmt *commenter) {
		cmt.providers = append(cmt.providers, prov)
	}
}

// WithAttrFunc configures commenter with AttrProviderFunc.
func WithAttrFunc(fn AttrProviderFunc) Option {
	return func(cmt *commenter) {
		cmt.providers = append(cmt.providers, fn)
	}
}
