package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"pethelp-backend/internal/core/ports"

	"github.com/mailjet/mailjet-apiv3-go"
)

type Sender struct {
	client     *mailjet.Client
	specialist ports.SpecialistService
}

func NewSender(apiKey string, specialist ports.SpecialistService) *Sender {
	ms := mailjet.NewMailjetClient(os.Getenv("EMAIL_SENDER_TOKEN"), os.Getenv("EMAIL_SECRET_KEY"))
	return &Sender{
		client:     ms,
		specialist: specialist,
	}

}

func (s *Sender) GetSpecialistConfirmationEmail(ctx context.Context, id int64) (string, error) {
	specialistData, err := s.specialist.ShowByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get specialist data: %w", err)
	}

	specialistEmail := specialistData.Email
	return specialistEmail, nil
}

func (s *Sender) SendAppointmentConfirmationEmail(ctx context.Context, id int64, clientEmail string, date, startTime, endTime time.Time) error {
	htmlBody := fmt.Sprintf(
		"<p>Your appointment on %s from %s to %s has been booked successfully!</p>",
		date.Format("2006-01-02"), startTime.Format("15:04"), endTime.Format("15:04"),
	)

	specialistEmail, err := s.GetSpecialistConfirmationEmail(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get specialist email: %w", err)
	}

	msg := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "pethelpdev@gmail.com",
				Name:  "PetHelp",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: clientEmail,
					Name:  "Client",
				},
			},
			Cc: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: specialistEmail,
					Name:  "Specialist",
				},
			},
			Subject:  "PetHelp: Appointment Confirmation",
			TextPart: "Your appointment has been booked successfully!",
			HTMLPart: htmlBody,
		},
	}

	messages := mailjet.MessagesV31{Info: msg}
	res, err := s.client.SendMailV31(&messages)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	fmt.Printf("Data: %+v\n", res)

	return nil
}
