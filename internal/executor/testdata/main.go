
package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "sleep" {
			time.Sleep(10 * time.Second)
		} else {
			fmt.Print(os.Args[1])
		}
	}
}
