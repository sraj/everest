# configx

`configx` is a small reusable environment configuration loader for Go applications.

It supports:
- Optional `.env` loading
- Optional key prefixing (e.g. `APP_`)
- Typed getters with defaults
- Required values with explicit errors
- Custom lookup sources for tests

## Install/Use

This package is internal to this repository at [pkg/configx](pkg/configx).

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/sraj/everest/pkg/configx"
)

type Config struct {
	Port   string
	Debug  bool
	Retries int
}

func Load() *Config {
	l := configx.New(configx.WithDotEnv())

	return &Config{
		Port:    l.String("PORT", "8080"),
		Debug:   l.Bool("DEBUG", false),
		Retries: l.Int("RETRIES", 3),
	}
}

func main() {
	cfg := Load()
	fmt.Println(cfg.Port, cfg.Debug, cfg.Retries)
}
```

## API

### `New(opts ...Option) *Loader`
Creates a loader with default lookup from `os.LookupEnv`.

### Options

- `WithDotEnv(paths ...string)`
  - Loads `.env` values before lookups.
  - If no path is provided, defaults to `.env`.

- `WithPrefix(prefix string)`
  - Prefixes keys automatically.
  - `WithPrefix("APP")` makes `String("PORT")` read `APP_PORT`.

- `WithLookup(fn func(string) (string, bool))`
  - Overrides environment lookup source.
  - Useful for tests and custom providers.

### Getters

- `String(key, def string) string`
- `Bool(key string, def bool) bool`
- `Int(key string, def int) int`
- `Duration(key string, def time.Duration) time.Duration`
- `RequiredString(key string) (string, error)`

Notes:
- Empty values are treated as unset and fall back to default.
- Invalid bool/int/duration values fall back to default.

## Examples

### 1) Using a Prefix

```go
l := configx.New(
	configx.WithDotEnv(),
	configx.WithPrefix("APP"),
)

port := l.String("PORT", "8080") // reads APP_PORT
```

### 2) Required Value

```go
dsn, err := l.RequiredString("DATABASE_URL")
if err != nil {
	// handle missing required config
}
```

### 3) Duration Parsing

```go
timeout := l.Duration("HTTP_TIMEOUT", 5*time.Second)
// env accepts values like: "250ms", "3s", "1m"
```

### 4) Test-Friendly Lookup

```go
fake := map[string]string{
	"PORT":  "9090",
	"DEBUG": "true",
}

lookup := func(k string) (string, bool) {
	v, ok := fake[k]
	return v, ok
}

l := configx.New(configx.WithLookup(lookup))
```

## Recommended Pattern

Keep `pkg/configx` generic, and define app-specific config structs in your app layer (e.g. `internal/config`):

```go
type Config struct {
	Port        string
	DatabaseURL string
}

func Load() *Config {
	l := configx.New(configx.WithDotEnv())
	return &Config{
		Port:        l.String("PORT", "8080"),
		DatabaseURL: l.String("DATABASE_URL", ""),
	}
}
```

This keeps `pkg` reusable across multiple applications.
