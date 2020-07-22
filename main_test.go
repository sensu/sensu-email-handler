package main

import (
	"fmt"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
)

var tcRcpts = []struct {
	name  string
	to    []string
	expTo string
}{
	// -t email1@sensu.com
	{"single", []string{"email1@example.com"}, "To: email1@example.com"},
	// -t email1@sensu.com,email2@sensu.com
	{"single_comma", []string{"email1@example.com,email2@example.com"}, "To: email1@example.com,email2@example.com"},
	// -t "email1@sensu.com, email2@sensu.com"
	{"single_comma_space", []string{"email1@example.com, email2@example.com"}, "To: email1@example.com,email2@example.com"},
	// -t email1@sensu.com -t email2@sensu.com
	{"multiple_flag", []string{"email1@example.com", "email2@example.com"}, "To: email1@example.com,email2@example.com"},
	// -t " email1@example.com\r\n, email2@example.com" -t email3@sensu.com
	// note invalid line endings removed, and order is changed
	{"multiple_flag_comma", []string{" email1@example.com\r\n, email2@example.com", "email3@example.com"}, "To: email1@example.com,email3@example.com,email2@example.com"},
	// -t email1@example.com -t "email2@example.com, email3@example.com" -t email4@example.com
	{"multiple_flag_comma2", []string{"email1@example.com", "email2@example.com, email3@example.com", "email4@example.com"}, "To: email1@example.com,email2@example.com,email4@example.com,email3@example.com"},
}

func TestNewRcpts(t *testing.T) {
	for _, tc := range tcRcpts {
		t.Run(tc.name, func(t *testing.T) {

			r := newRcpts(tc.to)
			assert.Equal(t, tc.expTo, fmt.Sprintf("To: %s", r), "receipients should be equal")
		})
	}
}

func TestResolveTemplate(t *testing.T) {
	event := corev2.FixtureEvent("foo", "bar")
	executed := time.Unix(event.Check.Executed, 0)
	executedFormatted := executed.Format("2 Jan 2006 15:04:05")
	template := "Entity: {{.Entity.Name}} Check: {{.Check.Name}} Executed: {{(UnixTime .Check.Executed).Format \"2 Jan 2006 15:04:05\"}}"
	templout, err := resolveTemplate(template, event, "text/plain")
	assert.NoError(t, err)
	expected := fmt.Sprintf("Entity: foo Check: bar Executed: %s", executedFormatted)
	assert.Equal(t, templout, expected)
	template = "<html>Entity: {{.Entity.Name}} Check: {{.Check.Name}} Executed: {{(UnixTime .Check.Executed).Format \"2 Jan 2006 15:04:05\"}}</html>"
	templout, err = resolveTemplate(template, event, "text/html")
	assert.NoError(t, err)
	expected = fmt.Sprintf("<html>Entity: foo Check: bar Executed: %s</html>", executedFormatted)
	assert.Equal(t, templout, expected)
}
