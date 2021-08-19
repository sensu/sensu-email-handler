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

func TestStringLines(t *testing.T) {
	lines, err := StringLines("Line1\nLine2")
	assert.Equal(t, 2, len(lines))
	assert.NoError(t, err)
	lines, err = StringLines("Line1\r\nLine2\r\nLine3")
	assert.Equal(t, 3, len(lines))
	assert.NoError(t, err)
}
func TestResolveTemplate(t *testing.T) {
	event := corev2.FixtureEvent("foo", "bar")
	executed := time.Unix(event.Check.Executed, 0)
	executedFormatted := executed.Format("2 Jan 2006 15:04:05")

	template := "Entity: {{.Entity.Name}} Check: {{.Check.Name}} Executed: {{(UnixTime .Check.Executed).Format \"2 Jan 2006 15:04:05\"}}"
	templout, err := resolveTemplate(template, event, "text/plain")
	assert.NoError(t, err)
	expected := fmt.Sprintf("Entity: foo Check: bar Executed: %s", executedFormatted)
	assert.Equal(t, expected, templout)

	template = "<html>Entity: {{.Entity.Name}} Check: {{.Check.Name}} Executed: {{(UnixTime .Check.Executed).Format \"2 Jan 2006 15:04:05\"}}</html>"
	templout, err = resolveTemplate(template, event, "text/html")
	assert.NoError(t, err)
	expected = fmt.Sprintf("<html>Entity: foo Check: bar Executed: %s</html>", executedFormatted)
	assert.Equal(t, expected, templout)
	event.Check.Output = "Test Unix newline\nSecond Line"
	template = "<html>Entity: {{.Entity.Name}} Check: {{.Check.Name}} Output: {{range $element := StringLines .Check.Output}}{{$element}}<br>{{end}}</html>"
	templout, err = resolveTemplate(template, event, "text/html")
	assert.NoError(t, err)
	expected = "<html>Entity: foo Check: bar Output: Test Unix newline<br>Second Line<br></html>"
	assert.Equal(t, expected, templout)

	t.Run("sprig_func", func(t *testing.T) {
		event2 := corev2.FixtureEvent("super.foo", " bar ")
		executed := time.Unix(event2.Check.Executed, 0)
		executedFormatted := executed.Format("2 Jan 2006 15:04:05")
		event2.Check.Interval = 600

		template = `<html>Entity: {{.Entity.Name | upper | trimPrefix "S"}} Check: {{trim .Check.Name}} Executed: {{(UnixTime .Check.Executed).Format "2 Jan 2006 15:04:05"}}</html>`
		templout, err = resolveTemplate(template, event2, "text/plain")
		assert.NoError(t, err)
		expected = fmt.Sprintf("<html>Entity: UPER.FOO Check: bar Executed: %s</html>", executedFormatted)
		assert.Equal(t, expected, templout)

		template = `{{ $host := split "." .Entity.Name}}<html>Entity: {{ $host._0 | upper }} Check: {{trim .Check.Name}} Executed: {{(UnixTime .Check.Executed).Format "2 Jan 2006 15:04:05"}} Interval: {{ div .Check.Interval 60 }} minutes</html>`
		templout, err = resolveTemplate(template, event2, "text/html")
		assert.NoError(t, err)
		expected = fmt.Sprintf("<html>Entity: SUPER Check: bar Executed: %s Interval: 10 minutes</html>", executedFormatted)
		assert.Equal(t, expected, templout)
	})

}
