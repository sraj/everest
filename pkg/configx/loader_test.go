package configx

import "testing"

func TestLoader_String(t *testing.T) {
	l := New(WithLookup(mapLookup(map[string]string{
		"PORT": "9090",
		"NAME": "  ",
	})))

	if got := l.String("PORT", "8080"); got != "9090" {
		t.Errorf("String(PORT) = %q, want %q", got, "9090")
	}
	if got := l.String("MISSING", "8080"); got != "8080" {
		t.Errorf("String(MISSING) = %q, want %q", got, "8080")
	}
	if got := l.String("NAME", "default"); got != "default" {
		t.Errorf("String(NAME) = %q, want %q (whitespace treated as unset)", got, "default")
	}
}

func TestLoader_Bool(t *testing.T) {
	l := New(WithLookup(mapLookup(map[string]string{
		"DEBUG": "true",
		"BAD":   "invalid",
	})))

	if got := l.Bool("DEBUG", false); got != true {
		t.Errorf("Bool(DEBUG) = %v, want true", got)
	}
	if got := l.Bool("BAD", true); got != true {
		t.Errorf("Bool(BAD) = %v, want true (falls back on parse error)", got)
	}
	if got := l.Bool("MISSING", true); got != true {
		t.Errorf("Bool(MISSING) = %v, want true (default)", got)
	}
}

func TestLoader_Int(t *testing.T) {
	l := New(WithLookup(mapLookup(map[string]string{
		"RETRIES": "5",
		"BAD":     "nope",
	})))

	if got := l.Int("RETRIES", 3); got != 5 {
		t.Errorf("Int(RETRIES) = %d, want 5", got)
	}
	if got := l.Int("BAD", 3); got != 3 {
		t.Errorf("Int(BAD) = %d, want 3 (default on parse error)", got)
	}
	if got := l.Int("MISSING", 3); got != 3 {
		t.Errorf("Int(MISSING) = %d, want 3 (default)", got)
	}
}

func TestLoader_Duration(t *testing.T) {
	l := New(WithLookup(mapLookup(map[string]string{
		"TIMEOUT": "15s",
		"BAD":     "xxx",
	})))

	if got := l.Duration("TIMEOUT", 0); got.String() != "15s" {
		t.Errorf("Duration(TIMEOUT) = %v, want 15s", got)
	}
	if got := l.Duration("BAD", 0); got != 0 {
		t.Errorf("Duration(BAD) = %v, want 0 (default on parse error)", got)
	}
	if got := l.Duration("MISSING", 0); got != 0 {
		t.Errorf("Duration(MISSING) = %v, want 0 (default)", got)
	}
}

func TestLoader_RequiredString(t *testing.T) {
	l := New(WithLookup(mapLookup(map[string]string{
		"KEY": "val",
		"GAP": "  ",
	})))

	if v, err := l.RequiredString("KEY"); err != nil || v != "val" {
		t.Errorf("RequiredString(KEY) = (%q, %v), want (val, nil)", v, err)
	}
	if _, err := l.RequiredString("GAP"); err == nil {
		t.Error("RequiredString(GAP) should return error for whitespace-only value")
	}
	if _, err := l.RequiredString("MISSING"); err == nil {
		t.Error("RequiredString(MISSING) should return error")
	}
}

func TestLoader_WithPrefix(t *testing.T) {
	l := New(
		WithPrefix("APP"),
		WithLookup(mapLookup(map[string]string{
			"APP_PORT": "3000",
			"HOST":     "localhost",
		})),
	)

	if got := l.String("PORT", "8080"); got != "3000" {
		t.Errorf("String(PORT) with prefix APP = %q, want 3000 (reads APP_PORT)", got)
	}
	if got := l.String("HOST", "0.0.0.0"); got != "0.0.0.0" {
		t.Errorf("String(HOST) = %q, want 0.0.0.0 (HOST != APP_HOST)", got)
	}
}

func mapLookup(m map[string]string) lookupFn {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}
