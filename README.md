# Sensu Go Email Handler Plugin
TravisCI: [![TravisCI Build Status](https://travis-ci.org/sensu/sensu-email-handler.svg?branch=master)](https://travis-ci.org/sensu/sensu-email-handler)

The Sensu Go Email Handler is a [Sensu Event Handler][2] for sending
incident notification emails.

## Installation

Download the latest version of the sensu-email-handler from [releases][1],
or create an executable script from this source.

From the local path of the sensu-email-handler repository:

```
go build -o /usr/local/bin/sensu-email-handler main.go
```
## Usage Examples

### Help

```
The Sensu Go Email handler for sending an email notification

Usage:
  sensu-email-handler [flags]

Flags:
  -T, --bodyTemplateFile string   A template file to use for the body
  -t, --toEmail string            The 'to' email address
  -f, --fromEmail string          The 'from' email address
  -h, --help                      help for sensu-email-handler
  -s, --smtpHost string           The SMTP host to use to send to send email
  -p, --smtpPassword string       The SMTP password, if not in env SMTP_PASSWORD
  -P, --smtpPort uint16           The SMTP server port (default 587)
  -u, --smtpUsername string       The SMTP username, if not in env SMTP_USERNAME
  -H, --hookout                   Include output from check hook(s)
  -i, --insecure                  Use an insecure connection (unauthenticated on port 25)
```
## Configuration

### Environment Variables and Annotations
|Environment Variable|Setting|Annotation|
|--------------------|-------|----------|
|SENSU_EMAIL_TO|-t / -- toEmail|sensu.io/plugins/email/config/to|
|SENSU_EMAIL_FROM|-f / --fromEmail|sensu.io/plugins/email/config/from|
|SENSU_EMAIL_SUBJECT_TEMPLATE|N/A|sensu.io/plugins/email/config/subject-template|
|SENSU_EMAIL_BODY_TEMPLATE|N/A|sensu.io/plugins/email/config/body-template|
|SENSU_EMAIL_SMTP_USERNAME|-u / --smtpUsername|N/A|
|SENSU_EMAIL_SMTP_PASSWORD|-p / --smtpPassword|N/A|

#### Precedence
environment variable < command-line argument < annotation

#### Definition Examples
Simple:
```json
{
    "api_version": "core/v2",
    "type": "Handler",
    "metadata": {
        "namespace": "default",
        "name": "email"
    },
    "spec": {
        "type": "pipe",
        "command": "sensu-email-handler -f from@example.com -t to@example.com -s smtp.example.com -u emailuser -p sup3rs3cr3t",
        "timeout": 10,
        "filters": [
            "is_incident",
            "not_silenced"
        ]
    }
}
```
Using Environment Variables and Annotations:

Handler:
```json
{
    "type": "Handler",
    "api_version": "core/v2",
    "metadata": {
        "name": "mail",
        "namespace": "default"
    },
    "spec": {
        "command": "sensu-email-handler -f sensu@example.com -s smtp.example.com",
        "env_vars": [
            "SENSU_EMAIL_SMTP_USERNAME=emailuser",
            "SENSU_EMAIL_SMTP_PASSWORD=sup3rs3cr3t"
        ],
        "filters": [
            "is_incident",
            "not_silenced"
        ],
        "handlers": null,
        "runtime_assets": [
            "sensu-email-handler"
        ],
        "timeout": 10,
        "type": "pipe"
    }
}
```
Check:
```json
{
  "type": "CheckConfig",
  "api_version": "core/v2",
  "metadata": {
    "name": "linux-cpu-check",
    "namespace": "AWS"
    "annotations": {
      "sensu.io/plugins/email/config/body-template": "Check: {{ .Check.Name }}\nEntity: {{ .Entity.Name }}\n\nOutput: {{ .Check.Output }}\n\nSensu URL: https://sensu.example.com:3000/{{ .Check.Namespace }}/events/{{ .Entity.Name }}/{{ .Check.Name }}\n",
      "sensu.io/plugins/email/config/to": "ops@example.com"
    },
  },
  "spec": {
    "check_hooks": [
      {
        "non-zero": [
          "linux-process-list-cpu-hook"
        ]
      }
    ],
    "command": "/opt/sensu-plugins-ruby/embedded/bin/check-cpu.rb -w {{ .labels.cpu_warning | default 90 }} -c {{ .labels.cpu_critical | default 95 }}",
    "env_vars": null,
    "handlers": [
      "mail",
      "flowdock"
    ],
    "high_flap_threshold": 0,
    "interval": 60,
    "low_flap_threshold": 0,
    "output_metric_format": "",
    "output_metric_handlers": null,
    "proxy_entity_name": "",
    "publish": true,
    "round_robin": false,
    "runtime_assets": null,
    "stdin": false,
    "subdue": null,
    "subscriptions": [
      "linux"
    ],
    "timeout": 0,
    "ttl": 0
  }
}

```
#### Template Defaults
The defaults for the two available templates are:

Subject:
```
Sensu Alert - {{.Entity.Name}}/{{.Check.Name}}: {{.Check.State}}
```
Body:
```
{{.Check.Output}}
```
Body when including hook output:
```
{{.Check.Output}}\n{{range .Check.Hooks}}Hook Name:  {{.Name}}\nHook Command:  {{.Command}}\n\n{{.Output}}\n\n{{end}}
```

## Contributing

See https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/sensu/sensu-email-handler/releases
[2]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
