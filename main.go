package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"os"
	"text/template"

	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

var (
	smtpHost     string
	smtpUsername string
	smtpPassword string
	smtpPort     uint16
	toEmail      string
	fromEmail    string
	subject      string
	hookout      bool
	stdin        *os.File

	emailSubjectTemplate = "Sensu Alert - {{.Entity.Name}}/{{.Check.Name}}: {{.Check.State}}"
	emailBodyTemplate    = "{{.Check.Output}}"
)

func main() {
	cmd := &cobra.Command{
		Use:   "sensu-email-handler",
		Short: "The Sensu Go Email handler for sending an email notification",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&smtpHost, "smtpHost", "s", "", "The SMTP host to use to send to send email")
	cmd.Flags().StringVarP(&smtpUsername, "smtpUsername", "u", "", "The SMTP username")
	cmd.Flags().StringVarP(&smtpPassword, "smtpPassword", "p", "", "The SMTP password")
	cmd.Flags().Uint16VarP(&smtpPort, "smtpPort", "P", 587, "The SMTP server port")
	cmd.Flags().StringVarP(&toEmail, "toEmail", "t", "", "The 'to' email address")
	cmd.Flags().StringVarP(&fromEmail, "fromEmail", "f", "", "The 'from' email address")
	cmd.Flags().BoolVarP(&hookout, "hookout", "H", false, "Include output from check hook(s)")

	cmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	validationError := checkArgs()
	if validationError != nil {
		return validationError
	}

	if stdin == nil {
		stdin = os.Stdin
	}

	eventJSON, err := ioutil.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %s", err)
	}

	event := &types.Event{}
	err = json.Unmarshal(eventJSON, event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin data: %s", err)
	}

	if err = event.Validate(); err != nil {
		return fmt.Errorf("failed to validate event: %s", err)
	}

	if !event.HasCheck() {
		return fmt.Errorf("event does not contain check")
	}

	sendMailError := sendEmail(event)
	if sendMailError != nil {
		return fmt.Errorf("failed to send email: %s", sendMailError)
	}

	return nil
}

func checkArgs() error {
	if len(smtpHost) == 0 {
		return errors.New("missing smtp host")
	}
	if len(toEmail) == 0 {
		return errors.New("missing destination email address")
	}
	if len(smtpUsername) == 0 {
		return errors.New("smtp username is empty")
	}
	if len(smtpPassword) == 0 {
		return errors.New("smtp password is empty")
	}
	if len(smtpPassword) == 0 {
		return errors.New("smtp password is empty")
	}
	if hookout {
		emailBodyTemplate = "{{.Check.Output}}\n{{range .Check.Hooks}}Hook Name:  {{.Name}}\nHook Command:  {{.Command}}\n\n{{.Output}}\n\n{{end}}"
	}
	if len(fromEmail) == 0 {
		return errors.New("from email is empty")
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

	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

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
