package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Printf("Start server\n")
	a := App{}
	// You need to set your Username and Password here
	a.Initialize("root", "root", "mysql")

	a.Run(":3000")
}