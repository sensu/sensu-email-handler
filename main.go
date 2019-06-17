package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/mail"
	"net/smtp"
	"strings"
	"text/template"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugins-go-library/sensu"
)

type HandlerConfig struct {
	sensu.PluginConfig
	SmtpHost         string
	SmtpUsername     string
	SmtpPassword     string
	SmtpPort         uint64
	ToEmail          string
	FromEmail        string
	FromHeader       string
	Subject          string
	Hookout          bool
	Insecure         bool
	LoginAuth        bool
	BodyTemplateFile string
}

type loginAuth struct {
	username, password string
}

const (
	smtpHost         = "smtpHost"
	smtpUsername     = "smtpUsername"
	smtpPassword     = "smtpPassword"
	smtpPort         = "smtpPort"
	toEmail          = "toEmail"
	fromEmail        = "fromEmail"
	insecure         = "insecure"
	hookout          = "hookout"
	loginauth        = "loginauth"
	bodyTemplateFile = "bodyTemplateFile"
	defaultSmtpPort  = 587
)

var (
	config = HandlerConfig{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-email-handler",
			Short:    "The Sensu Go Email handler for sending an email notification",
			Keyspace: "sensu.io/plugins/email/config",
		},
	}

	emailSubjectTemplate = "Sensu Alert - {{.Entity.Name}}/{{.Check.Name}}: {{.Check.State}}"
	emailBodyTemplate    = "{{.Check.Output}}"

	emailConfigOptions = []*sensu.PluginConfigOption{
		{
			Path:      smtpHost,
			Argument:  smtpHost,
			Shorthand: "s",
			Default:   "",
			Usage:     "The SMTP host to use to send to send email",
			Value:     &config.SmtpHost,
		},
		{
			Path:      smtpUsername,
			Env:       "SMTP_USERNAME",
			Argument:  smtpUsername,
			Shorthand: "u",
			Default:   "",
			Usage:     "The SMTP username, if not in env SMTP_USERNAME",
			Value:     &config.SmtpUsername,
		},
		{
			Path:      smtpPassword,
			Env:       "SMTP_PASSWORD",
			Argument:  smtpPassword,
			Shorthand: "p",
			Default:   "",
			Usage:     "The SMTP password, if not in env SMTP_PASSWORD",
			Value:     &config.SmtpPassword,
		},
		{
			Path:      smtpPort,
			Argument:  smtpPort,
			Shorthand: "P",
			Default:   uint64(defaultSmtpPort),
			Usage:     "The SMTP server port",
			Value:     &config.SmtpPort,
		},
		{
			Path:      toEmail,
			Argument:  toEmail,
			Shorthand: "t",
			Default:   "",
			Usage:     "The 'to' email address",
			Value:     &config.ToEmail,
		},
		{
			Path:      fromEmail,
			Argument:  fromEmail,
			Shorthand: "f",
			Default:   "",
			Usage:     "The 'from' email address",
			Value:     &config.FromEmail,
		},
		{
			Path:      insecure,
			Argument:  insecure,
			Shorthand: "i",
			Default:   false,
			Usage:     "Use an insecure connection (unauthenticated on port 25)",
			Value:     &config.Insecure,
		},
		{
			Path:      hookout,
			Argument:  hookout,
			Shorthand: "H",
			Default:   false,
			Usage:     "Include output from check hook(s)",
			Value:     &config.Hookout,
		},
		{
			Path:      loginauth,
			Argument:  loginauth,
			Shorthand: "l",
			Default:   false,
			Usage:     "Use \"login auth\" mechanisim",
			Value:     &config.LoginAuth,
		},
		{
			Path:      bodyTemplateFile,
			Argument:  bodyTemplateFile,
			Shorthand: "T",
			Default:   "",
			Usage:     "A template file to use for the body",
			Value:     &config.BodyTemplateFile,
		},
	}
)

func main() {
	goHandler := sensu.NewGoHandler(&config.PluginConfig, emailConfigOptions, checkArgs, sendEmail)
	goHandler.Execute()
}

func checkArgs(_ *corev2.Event) error {
	if len(config.SmtpHost) == 0 {
		return errors.New("missing smtp host")
	}
	if len(config.ToEmail) == 0 {
		return errors.New("missing destination email address")
	}
	if config.Insecure && config.LoginAuth {
		return fmt.Errorf("--insecure (-i) and --loginauth (-l) flags are mutually exclusive")
	}
	if !config.Insecure {
		if len(config.SmtpUsername) == 0 {
			return errors.New("smtp username is empty")
		}
		if len(config.SmtpPassword) == 0 {
			return errors.New("smtp password is empty")
		}
	} else {
		config.SmtpPort = 25
	}
	if config.SmtpPort > math.MaxUint16 {
		return errors.New("smtp port is out of range")
	}
	if config.Hookout && len(config.BodyTemplateFile) > 0 {
		return errors.New("--hookout (-H) and --bodyTemplateFile (-T) are mutually exclusive")
	}
	if config.Hookout {
		emailBodyTemplate = "{{.Check.Output}}\n{{range .Check.Hooks}}Hook Name:  {{.Name}}\nHook Command:  {{.Command}}\n\n{{.Output}}\n\n{{end}}"
	} else if len(config.BodyTemplateFile) > 0 {
		templateBytes, fileErr := ioutil.ReadFile(config.BodyTemplateFile)
		if fileErr != nil {
			return fmt.Errorf("failed to read specified template file %s", config.BodyTemplateFile)
		}
		emailBodyTemplate = string(templateBytes)
	}
	if len(config.FromEmail) == 0 {
		return errors.New("from email is empty")
	}
	fromAddr, addrErr := mail.ParseAddress(config.FromEmail)
	if addrErr != nil {
		return addrErr
	}
	config.FromEmail = fromAddr.Address
	config.FromHeader = fromAddr.String()
	return nil
}

func sendEmail(event *corev2.Event) error {
	var contentType string

	smtpAddress := fmt.Sprintf("%s:%d", config.SmtpHost, config.SmtpPort)
	subject, subjectErr := resolveTemplate(emailSubjectTemplate, event)
	if subjectErr != nil {
		return subjectErr
	}
	body, bodyErr := resolveTemplate(emailBodyTemplate, event)
	if bodyErr != nil {
		return bodyErr
	}

	if strings.Contains(body, "<html>") {
		contentType = "text/html"
	} else {
		contentType = "text/plain"
	}

	msg := []byte("From: " + config.FromHeader + "\r\n" +
		"To: " + config.ToEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: " + contentType + "\r\n" +
		"\r\n" +
		body + "\r\n")

	if config.Insecure {
		smtpconn, connErr := smtp.Dial(smtpAddress)
		if connErr != nil {
			return connErr
		}
		defer smtpconn.Close()
		smtpconn.Mail(config.FromEmail)
		smtpconn.Rcpt(config.ToEmail)
		smtpdata, dataErr := smtpconn.Data()
		if dataErr != nil {
			return dataErr
		}
		defer smtpdata.Close()
		buf := bytes.NewBuffer(msg)
		if _, dataErr := buf.WriteTo(smtpdata); dataErr != nil {
			return dataErr
		}

		return nil
	} else if config.LoginAuth {
		return smtp.SendMail(smtpAddress, LoginAuth(config.SmtpUsername, config.SmtpPassword), config.FromEmail, []string{config.ToEmail}, msg)
	}
	return smtp.SendMail(smtpAddress, smtp.PlainAuth("", config.SmtpUsername, config.SmtpPassword, config.SmtpHost), config.FromEmail, []string{config.ToEmail}, msg)

}

func resolveTemplate(templateValue string, event *corev2.Event) (string, error) {
	var resolved bytes.Buffer
	tmpl, err := template.New("test").Parse(templateValue)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(&resolved, *event)
	if err != nil {
		panic(err)
	}

	return resolved.String(), nil
}

// https://gist.github.com/homme/22b457eb054a07e7b2fb
// https://gist.github.com/andelf/5118732

// MIT license (c) andelf 2013

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unkown fromServer")
		}
	}
	return nil, nil
}
