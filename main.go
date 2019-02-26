package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/mail"
	"net/smtp"
	"os"
	"text/template"

	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
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
	stdin            *os.File

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
	cmd.Flags().BoolVarP(&insecure, "insecure", "i", false, "Use an insecure connection (unauthenticated on port 25)")
	cmd.Flags().StringVarP(&bodyTemplateFile, "bodyTemplateFile", "T", "", "A template file to use for the body")

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
