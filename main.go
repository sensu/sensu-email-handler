package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"text/template"
)

var (
	smtpHost      string
	smtpUsername  string
	smtpPassword  string
	smtpPort      uint16
	destEmail     string
	fromEmail     string
	subject       string
	eventJsonFile string
	stdin         *os.File

	emailSubjectTemplate = "Sensu Alert for entity {{.Entity.System.Hostname}} - {{.Check.Name}} - {{.Check.State}}"
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
	cmd.Flags().StringVarP(&destEmail, "destEmail", "d", "", "The destination email address")
	cmd.Flags().StringVarP(&fromEmail, "fromEmail", "f", "", "The from email address")
	cmd.Flags().StringVarP(&subject, "subject", "S", "", "The email subjetc")
	cmd.Flags().StringVarP(&eventJsonFile, "event", "e", "", "The JSON event file to process")

	cmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	validationError := checkArgs()
	if validationError != nil {
		return validationError
	}

	log.Println("Executing with arguments:", args)

	if stdin == nil {
		stdin = os.Stdin
	}

	event := &types.Event{}
	var eventJsonBytes []byte
	var err error
	if len(eventJsonFile) == 0 {
		eventJsonBytes, err = ioutil.ReadAll(stdin)
		log.Println("Event JSON:", eventJsonBytes)
	} else {
		//absoluteFilePath, _ := filepath.Abs(eventJsonFile)
		eventJsonBytes, err = ioutil.ReadFile(eventJsonFile)
	}
	if err != nil {
		return fmt.Errorf("Unexpected error: %s", err)
	}
	err = json.Unmarshal(eventJsonBytes, event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin data: %s", eventJsonBytes)
	}

	log.Println("Event", event)

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
	if len(destEmail) == 0 {
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

	msg := []byte("To: " + destEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	return smtp.SendMail(smtpAddress, smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost), fromEmail, []string{destEmail}, msg)
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
