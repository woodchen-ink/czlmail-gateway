package mail

import (
	"encoding/base64"
	"strings"
	"testing"
)

// base64Lines encodes n bytes the way mail clients do: 76 characters per line.
func base64Lines(n int) string {
	enc := base64.StdEncoding.EncodeToString(make([]byte, n))
	var b strings.Builder
	for len(enc) > 76 {
		b.WriteString(enc[:76] + "\r\n")
		enc = enc[76:]
	}
	b.WriteString(enc + "\r\n")
	return b.String()
}

// The sizes of gateway#2: base64 padding makes the BODYSTRUCTURE estimate up to
// 2 bytes large; a locally parsed message gets the exact decoded size.
func TestPartSize(t *testing.T) {
	for _, n := range []int{82, 204800, 34202, 0, 3} {
		raw := "From: a@example.test\r\nContent-Type: multipart/mixed; boundary=b\r\n\r\n" +
			"--b\r\nContent-Type: text/plain\r\n\r\nhi\r\n" +
			"--b\r\nContent-Type: application/octet-stream\r\nContent-Transfer-Encoding: base64\r\n\r\n" + base64Lines(n) +
			"--b--\r\n"
		rm := parseRawMessage([]byte(raw))
		pt := buildBody(rm.root).byID["2"]
		if pt == nil {
			t.Fatalf("%d bytes: no part 2", n)
		}
		if got := pt.size(rm); got != int64(n) {
			t.Errorf("%d bytes: exact size %d", n, got)
		}
		if est := pt.size(nil); est < int64(n) || est > int64(n)+2 {
			t.Errorf("%d bytes: estimate %d, want %d..%d", n, est, n, n+2)
		}
	}
}
