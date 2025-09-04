package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	// "github.com/mailersend/mailersend-go"

	"github.com/mailjet/mailjet-apiv3-go"
)

type Sender struct {
	client *mailjet.Client
}

func NewSender(apiKey string) *Sender {
	ms := mailjet.NewMailjetClient(os.Getenv("EMAIL_SENDER_TOKEN"), os.Getenv("EMAIL_SECRET_KEY"))
	return &Sender{
		client: ms,
	}

}



func(s *Sender) SendAppointmentConfirmationEmail(ctx context.Context, clientEmail string, date, startTime, endTime time.Time) error {
	htmlBody := fmt.Sprintf(
		"<p>Your appointment on %s from %s to %s has been booked successfully!</p>",
		date.Format("2006-01-02"), startTime.Format("15:04"), endTime.Format("15:04"),
	)


	msg := []mailjet.InfoMessagesV31{
		mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: "pethelpdev@gmail.com",
				Name: "PetHelp",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: clientEmail,
					Name: "Client",
				},
			},
			Subject: "PetHelp: Appointment Confirmation",
			TextPart: "Your appointment has been booked successfully!",
			HTMLPart: htmlBody,

		},
	}

	messages := mailjet.MessagesV31{Info: msg}
	_, err := s.client.SendMailV31(&messages)
		
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

