package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckIPProxyHandling(t *testing.T) {
	cases := []struct {
		name, remote, realIP, want string
		wantErr                    bool
	}{
		{name: "direct, no header", remote: "203.0.113.7:51000", want: "203.0.113.7"},
		{name: "public peer must not be trusted", remote: "203.0.113.7:51000", realIP: "1.2.3.4", want: "203.0.113.7"},
		{name: "private peer (nginx) is trusted", remote: "172.18.0.5:41000", realIP: "198.51.100.22", want: "198.51.100.22"},
		{name: "loopback peer is trusted", remote: "127.0.0.1:41000", realIP: "198.51.100.22", want: "198.51.100.22"},
		{name: "private peer, garbage header", remote: "172.18.0.5:41000", realIP: "not-an-ip", want: "172.18.0.5"},
		{name: "private peer, no header", remote: "172.18.0.5:41000", want: "172.18.0.5"},
		{name: "blank remote addr errors", remote: "", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/api/request/id", nil)
			r.RemoteAddr = tc.remote
			if tc.realIP != "" {
				r.Header.Set("X-Real-IP", tc.realIP)
			}
			got, err := checkIP(r)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got ip %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
