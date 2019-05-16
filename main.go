package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sensu/sensu-go/types"
	"io/ioutil"
	"net/mail"
	"net/smtp"
	"text/template"

	"github.com/sensu/sensu-plugins-go-library/sensu"
)

var (
	smtpHost         string
	smtpUsername     string
	smtpPassword     string
	smtpPort         uint16
	toEmail          string
	fromEmail        string
	fromHeader       string
	subject          string
	hookout          bool
	insecure         bool
	bodyTemplateFile string

	emailSubjectTemplate = "Sensu Alert - {{.Entity.Name}}/{{.Check.Name}}: {{.Check.State}}"
	emailBodyTemplate    = "{{.Check.Output}}"

	emailPluginConfig = sensu.PluginConfig{
		Name:  "sensu-email-handler",
		Short: "The Sensu Go Email handler for sending an email notification",
	}

	emailConfigOptions = []*sensu.PluginConfigOption{
		{
			Path:      "smtpHost",
			Argument:  "smtpHost",
			Shorthand: "s",
			Default:   "",
			Usage:     "The SMTP host to use to send to send email",
			Value:     &smtpHost,
		},
		{
			Path:      "smtpUsername",
			Env:       "SMTP_USERNAME",
			Argument:  "smtpUsername",
			Shorthand: "u",
			Default:   "",
			Usage:     "The SMTP username, if not in env SMTP_USERNAME",
			Value:     &smtpUsername,
		},
		{
			Path:      "smtpPassword",
			Env:       "SMTP_PASSWORD",
			Argument:  "smtpPassword",
			Shorthand: "p",
			Default:   "",
			Usage:     "The SMTP password, if not in env SMTP_PASSWORD",
			Value:     &smtpPassword,
		},
		{
			Path:      "smtpPort",
			Argument:  "smtpPort",
			Shorthand: "P",
			Default:   587,
			Usage:     "The SMTP server port",
			Value:     &smtpPort,
		},
		{
			Path:      "toEmail",
			Argument:  "toEmail",
			Shorthand: "t",
			Default:   "",
			Usage:     "The 'to' email address",
			Value:     &toEmail,
		},
		{
			Path:      "fromEmail",
			Argument:  "fromEmail",
			Shorthand: "f",
			Default:   "",
			Usage:     "The 'from' email address",
			Value:     &fromEmail,
		},
		{
			Path:      "insecure",
			Argument:  "insecure",
			Shorthand: "i",
			Default:   false,
			Usage:     "Use an insecure connection (unauthenticated on port 25)",
			Value:     &insecure,
		},
		{
			Path:      "hookout",
			Argument:  "hookout",
			Shorthand: "H",
			Default:   false,
			Usage:     "Include output from check hook(s)",
			Value:     &hookout,
		},
		{
			Path:      "bodyTemplateFile",
			Argument:  "bodyTemplateFile",
			Shorthand: "T",
			Default:   "",
			Usage:     "A template file to use for the body",
			Value:     &bodyTemplateFile,
		},
	}
)

func main() {
	goHandler, _ := sensu.NewGoHandler(&emailPluginConfig, emailConfigOptions, checkArgs, executeHandler)
	err := goHandler.Execute()
	if err != nil {
		fmt.Printf("Error executing plugin: %s", err)
	}
}

func checkArgs(_ *types.Event) error {
	if len(smtpHost) == 0 {
		return errors.New("missing smtp host")
	}
	if len(toEmail) == 0 {
		return errors.New("missing destination email address")
	}
	if !insecure {
		if len(smtpUsername) == 0 {
			return errors.New("smtp username is empty")
		}
		if len(smtpPassword) == 0 {
			return errors.New("smtp password is empty")
		}
	} else {
		smtpPort = 25
	}
	if hookout && len(bodyTemplateFile) > 0 {
		return errors.New("--hookout (-H) and --bodyTemplateFile (-T) are mutually exclusive")
	}
	if hookout {
		emailBodyTemplate = "{{.Check.Output}}\n{{range .Check.Hooks}}Hook Name:  {{.Name}}\nHook Command:  {{.Command}}\n\n{{.Output}}\n\n{{end}}"
	} else if len(bodyTemplateFile) > 0 {
		templateBytes, fileErr := ioutil.ReadFile(bodyTemplateFile)
		if fileErr != nil {
			return fmt.Errorf("failed to read specified template file %s", bodyTemplateFile)
		}
		emailBodyTemplate = string(templateBytes)
	}
	if len(fromEmail) == 0 {
		return errors.New("from email is empty")
	}
	fromAddr, addrErr := mail.ParseAddress(fromEmail)
	if addrErr != nil {
		return addrErr
	}
	fromEmail = fromAddr.Address
	fromHeader = fromAddr.String()
	return nil
}

func executeHandler(event *types.Event) error {
	sendMailError := sendEmail(event)
	if sendMailError != nil {
		return fmt.Errorf("failed to send email: %s", sendMailError)
	}

	return nil
}

func sendEmail(event *types.Event) error {
	smtpAddress := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	subject, subjectErr := resolveTemplate(emailSubjectTemplate, event)
	if subjectErr != nil {
		return subjectErr
	}
	body, bodyErr := resolveTemplate(emailBodyTemplate, event)
	if bodyErr != nil {
		return bodyErr
	}

	msg := []byte("From: " + fromHeader + "\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	if insecure {
		smtpconn, connErr := smtp.Dial(smtpAddress)
		if connErr != nil {
			return connErr
		}
		defer smtpconn.Close()
		smtpconn.Mail(fromEmail)
		smtpconn.Rcpt(toEmail)
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
	}
	return smtp.SendMail(smtpAddress, smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost), fromEmail, []string{toEmail}, msg)

}

func resolveTemplate(templateValue string, event *types.Event) (string, error) {
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
