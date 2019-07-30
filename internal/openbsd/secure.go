// +build !openbsd

package openbsd

// Secure is a no-op for non-OpenBSD systems.
func Secure(unveils [][2]string) error {
	return nil
}
