package email

import (
	"fmt"
	"net/smtp"

	"github.com/time_capsule/Auth-Servic-Timecapsule/config"
)

// SendOTP sends an OTP (One-Time Password) email to the specified recipient.
func SendOTP(cfg *config.Config, recipient string, otp string) error {
	// Construct the email message
	subject := "Your OTP Code"
	body := fmt.Sprintf("Your OTP code is: %s", otp)
	message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)

	// Set up authentication
	auth := smtp.PlainAuth("", cfg.EmailSender, cfg.EmailPassword, cfg.EmailHost)

	// Send the email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", cfg.EmailHost, cfg.EmailPort),
		auth,
		cfg.EmailFromAddress,
		[]string{recipient},
		[]byte(message),
	)
	if err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	return nil
}
