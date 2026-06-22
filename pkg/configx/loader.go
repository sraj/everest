package configx

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var ErrMissingRequired = errors.New("missing required config value")

type lookupFn func(string) (string, bool)

type Loader struct {
	prefix string
	lookup lookupFn
}

type Option func(*Loader)

func WithPrefix(prefix string) Option {
	return func(l *Loader) {
		p := strings.TrimSpace(prefix)
		if p == "" {
			l.prefix = ""
			return
		}
		l.prefix = strings.TrimSuffix(p, "_") + "_"
	}
}

func WithDotEnv(paths ...string) Option {
	return func(_ *Loader) {
		if len(paths) == 0 {
			_ = godotenv.Load()
			return
		}
		_ = godotenv.Load(paths...)
	}
}

func WithLookup(fn lookupFn) Option {
	return func(l *Loader) {
		if fn != nil {
			l.lookup = fn
		}
	}
}

func New(opts ...Option) *Loader {
	l := &Loader{
		lookup: os.LookupEnv,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *Loader) key(key string) string {
	if l.prefix == "" {
		return key
	}
	return l.prefix + key
}

func (l *Loader) String(key, def string) string {
	if v, ok := l.lookup(l.key(key)); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return def
}

func (l *Loader) RequiredString(key string) (string, error) {
	if v, ok := l.lookup(l.key(key)); ok && strings.TrimSpace(v) != "" {
		return v, nil
	}
	return "", errors.Join(ErrMissingRequired, errors.New(l.key(key)))
}

func (l *Loader) Bool(key string, def bool) bool {
	if v, ok := l.lookup(l.key(key)); ok && strings.TrimSpace(v) != "" {
		parsed, err := strconv.ParseBool(v)
		if err == nil {
			return parsed
		}
	}
	return def
}

func (l *Loader) Int(key string, def int) int {
	if v, ok := l.lookup(l.key(key)); ok && strings.TrimSpace(v) != "" {
		parsed, err := strconv.Atoi(v)
		if err == nil {
			return parsed
		}
	}
	return def
}

func (l *Loader) Duration(key string, def time.Duration) time.Duration {
	if v, ok := l.lookup(l.key(key)); ok && strings.TrimSpace(v) != "" {
		parsed, err := time.ParseDuration(v)
		if err == nil {
			return parsed
		}
	}
	return def
}
