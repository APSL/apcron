package main

import (
	"strings"
	"testing"
)

func TestCronParser(t *testing.T) {

	s := `
# Test
0 */10 * * * * echo "hello foo"
0 0 */5 * * * echo "hello bar"
`
	strReader := strings.NewReader(s)
	jobs, err := ParseCron(strReader)
	t.Logf("Jobs: %#v", jobs)
	if err != nil {
		t.Error(err)
	}
}
