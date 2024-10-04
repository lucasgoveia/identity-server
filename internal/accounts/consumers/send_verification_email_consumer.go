package consumers

import (
	"fmt"
	"go.uber.org/zap"
	"identity-server/internal/accounts/messages/commands"
	"identity-server/internal/accounts/services"
	"identity-server/pkg/providers/mailing"
	"reflect"
)

type SendVerificationEmailConsumer struct {
	verificationManager *services.IdentityVerificationManager
	logger              *zap.Logger
	mailSender          mailing.Sender
}

func NewSendVerificationEmailConsumer(verificationManager *services.IdentityVerificationManager, logger *zap.Logger, sender mailing.Sender) *SendVerificationEmailConsumer {
	return &SendVerificationEmailConsumer{verificationManager: verificationManager, logger: logger, mailSender: sender}
}

func (c *SendVerificationEmailConsumer) Handle(message interface{}) error {
	c.logger.Info("Received message in consumer",
		zap.String("type", reflect.TypeOf(message).String()))

	sendEmailVerificationMsg := message.(commands.SendVerificationEmail)
	otp, err := c.verificationManager.GenerateEmailOTP(sendEmailVerificationMsg.UserId, sendEmailVerificationMsg.IdentityId)

	if err != nil {
		c.logger.Error("Failed to generate OTP", zap.Error(err))
		return err
	}

	err = c.mailSender.Send(sendEmailVerificationMsg.Email, "Email verification", fmt.Sprintf("Your verification code is: %s", otp))

	if err != nil {
		c.logger.Error("Failed to send email", zap.Error(err))
		return err
	}

	return nil
}
