// Package email abstracts transactional e-mail sending behind an interface so
// the domain never depends on a concrete provider (Princípio I).
package email

import (
	"context"
	"fmt"
	"net/smtp"
	"sync"
)

// Message is a transactional e-mail to be delivered.
type Message struct {
	To       string
	Subject  string
	HTMLBody string
	TextBody string
}

// Sender delivers transactional e-mails. Implementations are injected into
// services as a dependency.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// FakeSender captures messages in memory for assertions in tests and local dev.
type FakeSender struct {
	mu   sync.Mutex
	Sent []Message
}

// Send records the message instead of delivering it.
func (f *FakeSender) Send(_ context.Context, msg Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Sent = append(f.Sent, msg)
	return nil
}

// Last returns the most recently captured message, if any.
func (f *FakeSender) Last() (Message, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Sent) == 0 {
		return Message{}, false
	}
	return f.Sent[len(f.Sent)-1], true
}

// SMTPSender delivers messages via an SMTP server.
type SMTPSender struct {
	addr string
	auth smtp.Auth
	from string
}

// NewSMTPSender builds an SMTP-backed Sender.
func NewSMTPSender(host, port, user, password, from string) *SMTPSender {
	var auth smtp.Auth
	if user != "" {
		auth = smtp.PlainAuth("", user, password, host)
	}
	return &SMTPSender{addr: fmt.Sprintf("%s:%s", host, port), auth: auth, from: from}
}

// Send delivers the message over SMTP.
func (s *SMTPSender) Send(_ context.Context, msg Message) error {
	body := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, msg.To, msg.Subject, msg.HTMLBody)
	return smtp.SendMail(s.addr, s.auth, s.from, []string{msg.To}, []byte(body))
}
