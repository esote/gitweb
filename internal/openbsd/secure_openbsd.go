// +build openbsd

package openbsd

import (
	"log"
	"net"

	"golang.org/x/sys/unix"
)

func init() {
	// force init of lazy sysctls
	if l, err := net.Listen("tcp", "localhost:0"); err != nil {
		log.Fatal(err)
	} else {
		l.Close()
	}
}

// Secure first unveils the specified paths, then pledges the minimal system
// calls needed for gitweb.
func Secure(unveils [][2]string) error {
	if err := unveil(unveils); err != nil {
		return err
	}
	return pledge()
}

func pledge() error {
	return unix.Pledge("stdio inet rpath proc exec", "stdio rpath tty")
}

func unveil(paths [][2]string) error {
	for _, path := range paths {
		if err := unix.Unveil(path[0], path[1]); err != nil {
			return err
		}
	}

	return nil
}
