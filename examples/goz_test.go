package goCurl

import (
	"fmt"
	"github.com/qifengzhang007/goCurl"
)

func ExampleCreateCurlClient() {
	cli := goCurl.CreateCurlClient()

	fmt.Printf("%T", cli)
	// Output: *goCurl.Request
}
