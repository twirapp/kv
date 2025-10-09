package kvoptions

import (
	"time"
)

type Option func(*Options)

type Options struct {
	Expire time.Duration
}

func Construct(options ...Option) Options {
	var opts Options
	for _, o := range options {
		o(&opts)
	}

	return opts
}

func WithExpire(d time.Duration) Option {
	return func(o *Options) {
		o.Expire = d
	}
}
