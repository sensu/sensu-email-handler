# Sensu Go Slack Handler

The Sensu slack handler is a [Sensu Event Handler][1] that sends an email using a SMTP server.

## Installation

Download the latest code from github.com:
```
go get github.com/fguimond/sensu-email-handler
```

Build the plugin:
```
go build -o /usr/local/bin/sensu-email-handler main.go
```

## Configuration

Example Sensu Go handler definition:

*TDB*

## Usage examples

Help:

```
The Sensu Go Email handler for sending an email notification

Usage:
  sensu-email-handler [flags]

Flags:
  -d, --destEmail string   The destination email address
  -h, --help               help for sensu-email-handler
  -s, --smtpHost string    The SMTP host to use to send to send email
```

[1]: https://docs.sensu.io/sensu-core/2.0/reference/handlers/#how-do-sensu-handlers-work
[2]: https://github.com/fguimond/sensu-email-handler
