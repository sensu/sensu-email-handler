package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
)

var (
	smtpHost  string
	destEmail string
	stdin     *os.File
)

func main() {
	cmd := &cobra.Command{
		Use:   "sensu-email-handler",
		Short: "The Sensu Go Email handler for sending an email notification",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&smtpHost, "smtpHost", "s", "", "The SMTP host to use to send to send email")
	cmd.Flags().StringVarP(&destEmail, "destEmail", "d", "", "The destination email address")

	cmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	if len(smtpHost) == 0 {
		return errors.New("missing smtp host")
	}
	if len(destEmail) == 0 {
		return errors.New("missing destination email address")
	}
	log.Println("Executing with arguments: smtpHost", smtpHost, "destEmail", destEmail)

	if stdin == nil {
		stdin = os.Stdin
	}

	eventJSON, err := ioutil.ReadAll(stdin)
	log.Println("Event JSON:", eventJSON)

	event := &types.Event{}
	err = json.Unmarshal(eventJSON, event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin data: %s", eventJSON)
	}

	log.Println("Event", event)
	return nil
}
