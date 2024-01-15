package test

import (
	"fmt"
	"strings"
	"testing"
)

func TestJoinStringEachLine(t *testing.T) {
	lineStr := []string{"1", "1", "google.com", "com", "475587", "2213358", "google.com", "com", "1", "1", "475902", "2214498"}
	resultJoin := strings.Join(lineStr, ",")
	fmt.Println(resultJoin)
}
