package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

type HandlerConfigOption struct {
	Value string
	Path  string
	Env   string
}

type HandlerConfig struct {
	EmailSubjectTemplate HandlerConfigOption
	EmailBodyTemplate    HandlerConfigOption
	FromEmail            HandlerConfigOption
	ToEmail              HandlerConfigOption
	Keyspace             string
}

var (
	smtpHost         string
	smtpUsername     string
	smtpPassword     string
	smtpPort         uint16
	fromHeader       string
	subject          string
	hookout          bool
	insecure         bool
	bodyTemplateFile string
	stdin            *os.File

	config = HandlerConfig{
		EmailSubjectTemplate: HandlerConfigOption{Value: "Sensu Alert - {{.Entity.Name}}/{{.Check.Name}}: {{.Check.State}}", Path: "subject-template", Env: "SENSU_EMAIL_SUBJECT_TEMPLATE"},
		EmailBodyTemplate:    HandlerConfigOption{Value: "{{.Check.Output}}", Path: "body-template", Env: "SENSU_EMAIL_BODY_TEMPLATE"},
		FromEmail:            HandlerConfigOption{Path: "from", Env: "SENSU_EMAIL_FROM"},
		ToEmail:              HandlerConfigOption{Path: "to", Env: "SENSU_EMAIL_TO"},
		Keyspace:             "sensu.io/plugins/email/config",
	}
	options = []*HandlerConfigOption{
		&config.EmailSubjectTemplate,
		&config.EmailBodyTemplate,
		&config.FromEmail,
		&config.ToEmail,
	}
)

func main() {
	cmd := &cobra.Command{
		Use:   "sensu-email-handler",
		Short: "The Sensu Go Email handler for sending an email notification",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&smtpHost, "smtpHost", "s", "", "The SMTP host to use to send to send email")
	cmd.Flags().StringVarP(&smtpUsername, "smtpUsername", "u", os.Getenv("SENSU_EMAIL_SMTP_USERNAME"), "The SMTP username, if not in env SENSU_EMAIL_SMTP_USERNAME")
	cmd.Flags().StringVarP(&smtpPassword, "smtpPassword", "p", os.Getenv("SENSU_EMAIL_SMTP_PASSWORD"), "The SMTP password, if not in env SENSU_EMAIL_SMTP_PASSWORD")
	cmd.Flags().Uint16VarP(&smtpPort, "smtpPort", "P", 587, "The SMTP server port")
	cmd.Flags().StringVarP(&config.ToEmail.Value, "toEmail", "t", os.Getenv(config.ToEmail.Env), "The 'to' email address, if not in env "+config.ToEmail.Env)
	cmd.Flags().StringVarP(&config.FromEmail.Value, "fromEmail", "f", os.Getenv(config.FromEmail.Env), "The 'from' email address, if not in env "+config.FromEmail.Env)
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

	configurationOverrides(&config, options, event)

	if len(config.ToEmail.Value) == 0 {
		return fmt.Errorf("missing destination email address")
	}
	if len(config.FromEmail.Value) == 0 {
		return fmt.Errorf("from email is empty")
	}

	parseError := parseFrom()
	if parseError != nil {
		return fmt.Errorf("failed to parse from address:  %s", parseError)
	}

	sendMailError := sendEmail(event)
	if sendMailError != nil {
		return fmt.Errorf("failed to send email: %s", sendMailError)
	}

	return nil
}

func checkArgs() error {
	if len(smtpHost) == 0 {
		return fmt.Errorf("missing smtp host")
	}
	if !insecure {
		if len(smtpUsername) == 0 {
			return fmt.Errorf("smtp username is empty")
		}
		if len(smtpPassword) == 0 {
			return fmt.Errorf("smtp password is empty")
		}
	} else {
		smtpPort = 25
	}
	if hookout && len(bodyTemplateFile) > 0 {
		return fmt.Errorf("--hookout (-H) and --bodyTemplateFile (-T) are mutually exclusive")
	}
	if hookout {
		config.EmailBodyTemplate.Value = "{{.Check.Output}}\n{{range .Check.Hooks}}Hook Name:  {{.Name}}\nHook Command:  {{.Command}}\n\n{{.Output}}\n\n{{end}}"
	} else if len(bodyTemplateFile) > 0 {
		templateBytes, fileErr := ioutil.ReadFile(bodyTemplateFile)
		if fileErr != nil {
			return fmt.Errorf("failed to read specified template file %s", bodyTemplateFile)
		}
		config.EmailBodyTemplate.Value = string(templateBytes)
	}

	return nil
}

func parseFrom() error {
	fromAddr, addrErr := mail.ParseAddress(config.FromEmail.Value)
	if addrErr != nil {
		return addrErr
	}
	config.FromEmail.Value = fromAddr.Address
	fromHeader = fromAddr.String()
	return nil
}

func sendEmail(event *types.Event) error {
	var contentType string
	smtpAddress := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	subject, subjectErr := resolveTemplate(config.EmailSubjectTemplate.Value, event)
	if subjectErr != nil {
		return subjectErr
	}
	body, bodyErr := resolveTemplate(config.EmailBodyTemplate.Value, event)
	if bodyErr != nil {
		return bodyErr
	}
	if strings.Contains(body, "<html>") {
		contentType = "text/html"
	} else {
		contentType = "text/plain"
	}

	msg := []byte("From: " + fromHeader + "\r\n" +
		"To: " + config.ToEmail.Value + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: " + contentType + "\r\n" +
		"\r\n" +
		body + "\r\n")

	if insecure {
		smtpconn, connErr := smtp.Dial(smtpAddress)
		if connErr != nil {
			return connErr
		}
		defer smtpconn.Close()
		smtpconn.Mail(config.FromEmail.Value)
		smtpconn.Rcpt(config.ToEmail.Value)
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
	return smtp.SendMail(smtpAddress, smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost), config.FromEmail.Value, []string{config.ToEmail.Value}, msg)

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

func configurationOverrides(config *HandlerConfig, options []*HandlerConfigOption, event *types.Event) {
	if config.Keyspace == "" {
		return
	}
	for _, opt := range options {
		if opt.Path != "" {
			// compile the Annotation keyspace to look for configuration overrides
			k := path.Join(config.Keyspace, opt.Path)
			switch {
			case event.Check.Annotations[k] != "":
				opt.Value = event.Check.Annotations[k]
				log.Printf("Overriding default handler configuration with value of \"Check.Annotations.%s\" (\"%s\")\n", k, event.Check.Annotations[k])
			case event.Entity.Annotations[k] != "":
				opt.Value = event.Entity.Annotations[k]
				log.Printf("Overriding default handler configuration with value of \"Entity.Annotations.%s\" (\"%s\")\n", k, event.Entity.Annotations[k])
			}
		}
	}
}
