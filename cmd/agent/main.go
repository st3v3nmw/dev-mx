package main

import (
	"fmt"
	"os"

	"github.com/st3v3nmw/devd/internal/policy"
)

func main() {
	snap := os.Args[1]
	result := policy.Check("snaps", map[string]string{"snap": snap})
	fmt.Println(result)
}
