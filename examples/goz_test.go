package goz

import (
	"fmt"
	"github.com/qifengzhang007/goCurl"
)

func ExampleNewClient() {
	cli := goz.NewClient()

	fmt.Printf("%T", cli)
	// Output: *goz.Request
}
