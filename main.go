package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net/smtp"
	"os"
)

var (
	smtpHost     string
	smtpUsername string
	smtpPassword string
	smtpPort     uint16
	destEmail    string
	fromEmail    string
	subject      string
	stdin        *os.File
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

	//eventJSON, err := ioutil.ReadAll(stdin)
	//log.Println("Event JSON:", eventJSON)

	//event := &types.Event{}
	//err = json.Unmarshal(eventJSON, event)
	//if err != nil {
	//	return fmt.Errorf("failed to unmarshal stdin data: %s", eventJSON)
	//}
	//
	//log.Println("Event", event)

	sendMailError := sendEmail()
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

func sendEmail() error {
	body := "Test Email"
	smtpAddress := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	return smtp.SendMail(smtpAddress, smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost), fromEmail, []string{destEmail}, []byte(body))
}
