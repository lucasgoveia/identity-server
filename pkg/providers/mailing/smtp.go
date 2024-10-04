package mailing

import (
	"crypto/tls"
	"fmt"
	"go.uber.org/zap"
	"identity-server/config"
	"net"
	"net/smtp"
)

type SmtpSender struct {
	config *config.SmtpConfig
	logger *zap.Logger
}

func NewSmtpSender(config *config.SmtpConfig, logger *zap.Logger) *SmtpSender {
	return &SmtpSender{config: config, logger: logger}
}

func (s *SmtpSender) Send(to, subject, body string) error {
	// Create the email headers
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.From)
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	// Connect to the SMTP server over TLS if configured
	serverAddr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Set up TLS configuration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false, // Use true only if testing with invalid certs, not in production
		ServerName:         s.config.Host,
	}

	// Establish a TLS connection if TLS is enabled
	var conn net.Conn
	var err error
	if s.config.TLS {
		conn, err = tls.Dial("tcp", serverAddr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to dial TLS: %w", err)
		}
	} else {
		conn, err = net.Dial("tcp", serverAddr)
		if err != nil {
			return fmt.Errorf("failed to dial server: %w", err)
		}
	}
	defer conn.Close()

	// Create SMTP client with the connection
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Start TLS if it's enabled but was not already established
	if s.config.TLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("failed to start TLS: %w", err)
			}
		}
	}

	// Authenticate if DefaultCredentials is false (use authentication)
	if !s.config.DefaultCredentials {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	// Set the sender and recipient
	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send the email body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get writer: %w", err)
	}
	_, err = writer.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}
