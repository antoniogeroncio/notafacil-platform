package token

import (
	"testing"
	"time"
)

func TestSessionRoundTrip(t *testing.T) {
	m := NewManager("super-secret", 30*time.Minute)
	issued, err := m.Issue(SessionClaims{UserID: "u1", TenantID: "t1", Role: "Admin"})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	got, err := m.Parse(issued)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.UserID != "u1" || got.TenantID != "t1" || got.Role != "Admin" {
		t.Errorf("claims mismatch: %+v", got)
	}
}

func TestSessionExpired(t *testing.T) {
	m := NewManager("super-secret", -1*time.Minute) // already expired
	issued, err := m.Issue(SessionClaims{UserID: "u1", TenantID: "t1", Role: "Editor"})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if _, err := m.Parse(issued); err == nil {
		t.Errorf("expected expired token to fail parsing")
	}
}

func TestSessionTamperedSignature(t *testing.T) {
	m := NewManager("super-secret", 30*time.Minute)
	other := NewManager("different-secret", 30*time.Minute)
	issued, _ := other.Issue(SessionClaims{UserID: "u1", TenantID: "t1", Role: "Viewer"})
	if _, err := m.Parse(issued); err == nil {
		t.Errorf("expected token signed with another secret to be rejected")
	}
}

func TestInviteTokenHashing(t *testing.T) {
	raw, hash, err := GenerateInviteToken()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if raw == "" || hash == "" {
		t.Fatal("raw and hash must be non-empty")
	}
	if raw == hash {
		t.Error("raw token must never equal its stored hash (Princípio VI)")
	}
	if HashInviteToken(raw) != hash {
		t.Error("HashInviteToken(raw) must reproduce the stored hash")
	}
}

func TestInviteTokenIsUnique(t *testing.T) {
	_, h1, _ := GenerateInviteToken()
	_, h2, _ := GenerateInviteToken()
	if h1 == h2 {
		t.Error("two generated invite tokens must differ")
	}
}
