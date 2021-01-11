package storage

import "github.com/jrapoport/chestnut/log"

// StoreOptions provides a default implementation for common storage Options stores should support.
type StoreOptions struct {
	log log.Logger
}

// Logger returns the configured logger for the store.
func (o StoreOptions) Logger() log.Logger {
	return o.log
}

// DefaultStoreOptions represents the recommended default StoreOptions for a store.
var DefaultStoreOptions = StoreOptions{
	log: log.Log,
}

// A StoreOption sets options such disabling overwrite, and other parameters, etc.
type StoreOption interface {
	apply(*StoreOptions)
}

// EmptyStoreOption does not alter the store configuration.
// It can be embedded in another structure to build custom options.
type EmptyStoreOption struct{}

func (EmptyStoreOption) apply(*StoreOptions) {}

// funcOption wraps a function that modifies StoreOptions
// into an implementation of the StoreOption interface.
type funcOption struct {
	f func(*StoreOptions)
}

// Apply applies an StoreOption to StoreOptions.
func (fdo *funcOption) apply(do *StoreOptions) {
	fdo.f(do)
}

func newFuncOption(f func(*StoreOptions)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// ApplyOptions accepts an StoreOptions struct and applies the StoreOption(s) to it.
func ApplyOptions(opts StoreOptions, opt ...StoreOption) StoreOptions {
	for _, o := range opt {
		o.apply(&opts)
	}
	return opts
}

// WithLogger returns a StoreOption which sets the logger to use for the encrypted store.
func WithLogger(l log.Logger) StoreOption {
	return newFuncOption(func(o *StoreOptions) {
		o.log = l
	})
}

// WithStdLogger is a convenience that returns a StoreOption for a standard err logger.
func WithStdLogger(lvl log.Level) StoreOption {
	return WithLogger(log.NewStdLoggerWithLevel(lvl))
}

// WithLogrusLogger is a convenience that returns a StoreOption for a default logrus logger.
func WithLogrusLogger(lvl log.Level) StoreOption {
	return WithLogger(log.NewLogrusLoggerWithLevel(lvl))
}

// WithZapLogger is a convenience that returns a StoreOption for a production zap logger.
func WithZapLogger(lvl log.Level) StoreOption {
	return WithLogger(log.NewZapLoggerWithLevel(lvl))
}
