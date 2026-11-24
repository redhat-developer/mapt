// Package ssh provides helpers to work with SSH keys.
package ssh

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

// Fingerprint returns the SHA256 fingerprint of an OpenSSH public key, in
// the same format as `ssh-keygen -lf` (e.g. "SHA256:xxxxxxxx..."). It is
// meant for logging: the fingerprint identifies the key without exposing
// any secret material. If the key cannot be parsed it returns a
// description of the error instead.
func Fingerprint(publicKeyOpenssh string) string {
	pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKeyOpenssh))
	if err != nil {
		return fmt.Sprintf("unavailable: %v", err)
	}
	return ssh.FingerprintSHA256(pk)
}
