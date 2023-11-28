package e2e_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

func TestApp(t *testing.T) {
	resp, err := http.Get(mustAddr())
	if err != nil {
		t.Errorf("get: %v", err)
	}

	code := resp.StatusCode
	if code < 200 || code > 299 {
		t.Errorf("non-ok code: %d", code)
	}
}

func mustAddr() string {
	addr := os.Getenv("SAMPLE_ADDR")
	if addr == "" {
		panic(fmt.Errorf("`SAMPLE_ADDR` env var must be set"))
	}

	return addr
}
