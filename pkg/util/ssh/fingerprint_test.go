package ssh

import "testing"

func TestFingerprint(t *testing.T) {
	// ssh-keygen -t ed25519 public key with known fingerprint
	pub := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIK83XlVobqlz7JuvhORJs+xdFnxg1PR9pPWerUVFBl3s test@example"
	want := "SHA256:PTBgfmDJ1OAdXXvJAb9uMlAbc82bpHpEDKW9hr2uIyU"
	if got := Fingerprint(pub); got != want {
		t.Errorf("Fingerprint() = %q, want %q", got, want)
	}
}

func TestFingerprintInvalidKey(t *testing.T) {
	got := Fingerprint("not a public key")
	if got == "" {
		t.Error("Fingerprint() on invalid input should describe the error, not return empty")
	}
}
