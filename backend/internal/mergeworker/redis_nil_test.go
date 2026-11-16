package mergeworker

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestIsRedisNilRecognizesRESPAbsenceOnly(t *testing.T) {
	_, missing := readRESP(bufio.NewReader(strings.NewReader("$-1\r\n")))
	if !IsRedisNil(missing) || !IsRedisNil(fmt.Errorf("read: %w", missing)) {
		t.Fatal("missing RESP value must remain identifiable")
	}
	if IsRedisNil(nil) || IsRedisNil(errors.New("redis nil")) || IsRedisNil(errors.New("connection failed")) {
		t.Fatal("other errors must not become missing values")
	}
}
