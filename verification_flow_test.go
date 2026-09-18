package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerificationRequestBoundary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/email/send" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatal("missing bearer auth")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"data":{"message_id":"m-42"}}`))
	}))
	defer server.Close()
	c := &InfraiClient{BaseURL: server.URL, Key: "test-key", HTTP: server.Client()}
	result, err := SendVerification(c, "https://app.example", Signup{Email: "dev@example.com", UserID: "u-7"})
	if err != nil || result.MessageID != "m-42" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestVerificationLinkShape(t *testing.T) {
	link, err := VerificationLink("https://app.example", Signup{UserID: "u-7"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(link, "https://app.example/verify-email?user=u-7&token=") {
		t.Fatalf("link=%s", link)
	}
}
