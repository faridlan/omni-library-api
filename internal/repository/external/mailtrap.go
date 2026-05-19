package external

import (
	"fmt"
	"net/smtp"

	"github.com/faridlan/omni-library-api/internal/domain"
)

type mailtrapSender struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func NewMailtrapSender(host, port, username, password, from string) domain.EmailSender {
	return &mailtrapSender{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func (m *mailtrapSender) SendVerificationEmail(toEmail string, token string) error {
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)

	// URL endpoint verifikasi kita
	verifyLink := fmt.Sprintf("http://localhost:8080/api/auth/verify-email?token=%s", token)

	subject := "Subject: Verifikasi Email Omni Library Anda\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<h2>Selamat datang di Omni Library!</h2>
		<p>Klik link di bawah ini untuk memverifikasi email Anda:</p>
		<a href="%s" style="padding: 10px 20px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px;">Verifikasi Email</a>
		<br><br>
		<p>Link ini akan kadaluarsa dalam 24 jam.</p>
	`, verifyLink)

	msg := []byte(subject + mime + body)

	addr := fmt.Sprintf("%s:%s", m.Host, m.Port)
	err := smtp.SendMail(addr, auth, m.From, []string{toEmail}, msg)
	if err != nil {
		return err
	}

	return nil
}

func (m *mailtrapSender) SendPasswordResetEmail(toEmail string, token string) error {
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)

	// URL ini idealnya mengarah ke halaman Frontend kamu (misal: React/NextJS)
	// yang memiliki form untuk memasukkan password baru.
	resetLink := fmt.Sprintf("http://localhost:8080/api/auth/reset-password?token=%s", token)

	subject := "Subject: Reset Password Omni Library Anda\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<h2>Permintaan Reset Password</h2>
		<p>Seseorang telah meminta untuk melakukan reset password pada akun Omni Library Anda.</p>
		<p>Jika ini bukan Anda, abaikan email ini. Jika ini Anda, klik link di bawah ini untuk membuat password baru:</p>
		<a href="%s" style="padding: 10px 20px; background-color: #f44336; color: white; text-decoration: none; border-radius: 5px;">Reset Password</a>
		<br><br>
		<p>Link ini hanya berlaku selama 15 menit.</p>
	`, resetLink)

	msg := []byte(subject + mime + body)

	addr := fmt.Sprintf("%s:%s", m.Host, m.Port)
	err := smtp.SendMail(addr, auth, m.From, []string{toEmail}, msg)
	if err != nil {
		return err
	}

	return nil
}
