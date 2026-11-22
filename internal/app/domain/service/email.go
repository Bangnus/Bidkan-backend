package service

import "context"

// EmailMessage represents an outbound transactional email.
type EmailMessage struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

// EmailProvider abstracts transactional email delivery (SMTP or dev console).
type EmailProvider interface {
	Send(ctx context.Context, msg EmailMessage) error
	Mode() string
}
