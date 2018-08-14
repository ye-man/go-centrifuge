// +build unit

package errors

import (
	"reflect"
	"testing"
)

func TestP2PError(t *testing.T) {
	tests := []struct {
		code   Code
		msg    string
		errors map[string]string
	}{
		{
			code: AuthenticationFailed,
			msg:  "Node authentication failed",
		},

		{
			code: DocumentNotFound,
			msg:  "Invalid document",
			errors: map[string]string{
				"document_root":   "root empty",
				"next_identifier": "invalid identifier",
			},
		},

		{
			code: Code(100),
			msg:  "Unknown error",
		},
	}

	for _, c := range tests {
		err := NewWithErrors(c.code, c.msg, c.errors)
		p2perr, ok := FromError(err)
		if !ok {
			t.Fatalf("unexpected conversion error: %T", err)
		}

		if got := p2perr.Message(); got != c.msg {
			t.Fatalf("message mismatch: %s != %s", got, c.msg)
		}

		if got := p2perr.Errors(); !reflect.DeepEqual(got, c.errors) {
			t.Fatalf("errors mismatch: %v != %v", got, c.errors)
		}

		want := c.code
		if want >= maxCode {
			want = Unknown
		}

		if got := p2perr.Code(); got != want {
			t.Fatalf("code mismatch: %v != %v", got, want)
		}
	}
}
