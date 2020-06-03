package goCurl

import (
	"fmt"
	"github.com/qifengzhang007/goCurl"
)

func ExampleNewClient() {
	cli := goCurl.NewClient()

	fmt.Printf("%T", cli)
	// Output: *goCurl.Request
}
