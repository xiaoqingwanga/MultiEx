package main

import (
	"fmt"
	"time"
)

func main() {
	// server.Main()
	t := time.Now().Unix
	fmt.Print(fmt.Sprintf("%d", t))

}
