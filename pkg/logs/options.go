package logs

const (
	TraceLevel = "trace"
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
	PanicLevel = "panic"
	FatalLevel = "fatal"
)

const (
	TextFormat = "text"
	JsonFormat = "json"
)

// Logger options
type Options struct {
	level  string
	format string
}

type Option func(*Options)

// WithLevel sets log level
func WithLevel(level string) Option {
	return func(o *Options) {
		o.level = level
	}
}

// WithFormat sets log format
func WithFormat(format string) Option {
	return func(o *Options) {
		o.format = format
	}
}
