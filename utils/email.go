package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendVerificationEmail(toEmail string, code string) error {
	from := os.Getenv("NODEMAILER_EMAIL")
	password := os.Getenv("NODEMAILER_PASSWORD")

	if from == "" || password == "" {
		fmt.Println("Warning: Email credentials not set. Skipping email send.")
		return nil
	}

	// For Gmail SMTP
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Set up authentication information.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Build the email message
	subject := "Subject: Your Verification Code\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">Welcome to AI PDF Assistant!</h2>
			<p>Thank you for signing up. Please verify your email address by entering the following 6-digit code:</p>
			<div style="background-color: #f4f4f4; padding: 20px; text-align: center; border-radius: 5px; margin: 20px 0;">
				<h1 style="margin: 0; letter-spacing: 5px; color: #2563eb;">%s</h1>
			</div>
			<p>If you didn't request this, you can safely ignore this email.</p>
		</div>
	`, code)

	msg := []byte(subject + mime + body)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func SendPasswordResetEmail(toEmail string, token string) error {
	from := os.Getenv("NODEMAILER_EMAIL")
	password := os.Getenv("NODEMAILER_PASSWORD")

	if from == "" || password == "" {
		fmt.Println("Warning: Email credentials not set. Skipping email send.")
		return nil
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	auth := smtp.PlainAuth("", from, password, smtpHost)

	frontendURL := "https://ai-pdf-assistant-frontend-ntha4y-ff493a-35-180-95-158.sslip.io"
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	subject := "Subject: Reset your password\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<h2>Password Reset Request</h2>
		<p>We received a request to reset your password. Click the link below to set a new password:</p>
		<p><a href="%s" style="display:inline-block;padding:10px 20px;color:white;background-color:#2563eb;text-decoration:none;border-radius:5px;">Reset Password</a></p>
		<p>If you did not request this, please ignore this email.</p>
		<p>Or paste this link in your browser: %s</p>
	`, resetLink, resetLink)

	msg := []byte(subject + mime + body)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
