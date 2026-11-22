package email

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	domainService "github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
)

type smtpConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

type provider struct {
	cfg smtpConfig
}

// NewProvider initializes SMTP email delivery when SMTP_HOST is configured, or falls back to console logging in dev.
func NewProviderFromEnv() domainService.EmailProvider {
	port := 587
	if rawPort := strings.TrimSpace(os.Getenv("SMTP_PORT")); rawPort != "" {
		if parsed, err := strconv.Atoi(rawPort); err == nil && parsed > 0 {
			port = parsed
		}
	}
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	if from == "" {
		if user != "" {
			from = fmt.Sprintf("BidKan Campus <%s>", user)
		} else {
			from = "BidKan Campus <no-reply@bidkan.local>"
		}
	}
	return &provider{
		cfg: smtpConfig{
			Host: host,
			Port: port,
			User: user,
			Pass: os.Getenv("SMTP_PASS"),
			From: from,
		},
	}
}

func (p *provider) Mode() string {
	if p.cfg.Host != "" {
		return "smtp"
	}
	return "console"
}

func (p *provider) Send(ctx context.Context, msg domainService.EmailMessage) error {
	if strings.ContainsAny(msg.To, "\r\n") || strings.ContainsAny(p.cfg.From, "\r\n") {
		return fmt.Errorf("invalid email header characters")
	}
	parsedTo, err := mail.ParseAddress(strings.TrimSpace(msg.To))
	if err != nil || parsedTo.Address == "" {
		return fmt.Errorf("invalid recipient email address")
	}
	msg.To = parsedTo.Address

	if p.cfg.Host == "" {
		log.Printf("📧 [CONSOLE EMAIL] To: %s | Subject: %s\n%s", msg.To, msg.Subject, msg.Text)
		return nil
	}
	payload := buildMIME(p.cfg.From, msg)
	addr := net.JoinHostPort(p.cfg.Host, strconv.Itoa(p.cfg.Port))
	var auth smtp.Auth
	if p.cfg.User != "" {
		auth = smtp.PlainAuth("", p.cfg.User, p.cfg.Pass, p.cfg.Host)
	}
	if p.cfg.Port == 465 {
		return sendImplicitTLS(ctx, addr, p.cfg.Host, auth, extractEmailAddr(p.cfg.From), msg.To, payload)
	}
	return sendSTARTTLS(addr, p.cfg.Host, auth, extractEmailAddr(p.cfg.From), msg.To, payload)
}

func sendSTARTTLS(addr, host string, auth smtp.Auth, from, to string, payload []byte) error {
	return smtp.SendMail(addr, auth, from, []string{to}, payload)
}

func sendImplicitTLS(ctx context.Context, addr, host string, auth smtp.Auth, from, to string, payload []byte) error {
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 10 * time.Second},
		Config:    &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12},
	}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer client.Close()
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(payload); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func extractEmailAddr(from string) string {
	start := strings.Index(from, "<")
	end := strings.Index(from, ">")
	if start >= 0 && end > start {
		return strings.TrimSpace(from[start+1 : end])
	}
	return strings.TrimSpace(from)
}

func buildMIME(from string, msg domainService.EmailMessage) []byte {
	boundary := "bidkan-email-boundary-7c3aed"
	encodedSubj := "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(msg.Subject)) + "?="
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + msg.To + "\r\n")
	b.WriteString("Subject: " + encodedSubj + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n\r\n")
	b.WriteString("--" + boundary + "\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\nContent-Transfer-Encoding: base64\r\n\r\n")
	b.WriteString(base64.StdEncoding.EncodeToString([]byte(msg.Text)) + "\r\n")
	b.WriteString("--" + boundary + "\r\n")
	b.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\nContent-Transfer-Encoding: base64\r\n\r\n")
	b.WriteString(base64.StdEncoding.EncodeToString([]byte(msg.HTML)) + "\r\n")
	b.WriteString("--" + boundary + "--\r\n")
	return []byte(b.String())
}
