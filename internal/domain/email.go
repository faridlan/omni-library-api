package domain

type EmailSender interface {
	SendVerificationEmail(toEmail string, token string) error
	SendPasswordResetEmail(toEmail string, token string) error
}
