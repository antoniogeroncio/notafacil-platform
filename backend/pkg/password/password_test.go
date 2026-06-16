package password

import "testing"

func TestHashProducesVerifiableHash(t *testing.T) {
	hash, err := Hash("s3nh4forte")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" || hash == "s3nh4forte" {
		t.Fatalf("hash must not be empty nor equal to the plaintext, got %q", hash)
	}
	if !Verify(hash, "s3nh4forte") {
		t.Errorf("Verify should accept the correct password")
	}
}

func TestVerifyRejectsWrongPassword(t *testing.T) {
	hash, err := Hash("s3nh4forte")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if Verify(hash, "outra-senha") {
		t.Errorf("Verify should reject an incorrect password")
	}
}

func TestHashIsSaltedPerCall(t *testing.T) {
	a, _ := Hash("mesma-senha")
	b, _ := Hash("mesma-senha")
	if a == b {
		t.Errorf("two hashes of the same password must differ (salt)")
	}
}
