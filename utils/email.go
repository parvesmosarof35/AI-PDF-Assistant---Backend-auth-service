package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendVerificationEmail(toEmail string, token string) error {
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

	// Verification Link
	// Assuming frontend will handle /verify?token=... or point to backend directly
	// We'll point directly to backend for simplicity: GET /api/auth/verify?token=
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}
	verificationLink := fmt.Sprintf("http://localhost:%s/api/auth/verify?token=%s", port, token)

	// Build the email message
	subject := "Subject: Verify your AI Assistant Account\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<h2>Welcome to AI PDF Assistant!</h2>
		<p>Thank you for signing up. Please verify your email address by clicking the link below:</p>
		<p><a href="%s" style="display:inline-block;padding:10px 20px;color:white;background-color:#2563eb;text-decoration:none;border-radius:5px;">Verify Email</a></p>
		<p>Or paste this link in your browser: %s</p>
	`, verificationLink, verificationLink)

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

	// In a real app, this should point to your Next.js frontend route, e.g., http://localhost:3000/reset-password?token=
	// We'll assume frontend runs on 3000
	frontendURL := "http://localhost:3000"
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
