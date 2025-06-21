package main

import (
	"fmt"
	"os"

	"github.com/st3v3nmw/devd/internal/engine"
)

func main() {
	snap := os.Args[1]
	result := engine.CheckSnapPolicy(snap)
	fmt.Println(result)
}
