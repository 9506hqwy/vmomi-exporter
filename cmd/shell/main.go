package main

import (
	"fmt"

	"github.com/9506hqwy/template-go-module/pkg/example"
)

var version = "<version>"
var commit = "<commit>"

func main() {
	ret := example.Add(2, 5)
	fmt.Printf("result = %d\n", ret)
	fmt.Printf("version = %s, commit = %s\n", version, commit)
}
