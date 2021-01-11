package log

// A Level is a logging priority. Higher levels are more important.
// This is here as a convenience when using the various log options.
type Level int

const (
	// DebugLevel logs are typically voluminous,
	// and are usually disabled in production.
	DebugLevel Level = iota - 1

	// InfoLevel is the default logging priority.
	InfoLevel

	// WarnLevel logs are more important than Info,
	// but don't need individual human review.
	WarnLevel

	// ErrorLevel logs are high-priority. If an application runs
	// smoothly, it shouldn't generate any error-level logs.
	ErrorLevel

	// PanicLevel logs a message, then panics.
	PanicLevel

	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel
)
