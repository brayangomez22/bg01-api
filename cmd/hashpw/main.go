// Command hashpw prints a bcrypt hash for use as ADMIN_PASSWORD_HASH.
//
//	go run ./cmd/hashpw 'my-secret-password'
package main

import (
	"fmt"
	"os"

	"github.com/brayangomez22/bg01-api/internal/auth"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: hashpw <password>")
		os.Exit(1)
	}
	hash, err := auth.HashPassword(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println(hash)
}
