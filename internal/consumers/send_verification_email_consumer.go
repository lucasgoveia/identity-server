package consumers

import (
	"fmt"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
	"identity-server/internal/cache"
	"identity-server/internal/mailing"
	"identity-server/internal/messages/commands"
	"identity-server/internal/security"
	"reflect"
)

type SendVerificationEmailConsumer struct {
	otpGen     *security.OTPGenerator
	cache      cache.Cache
	logger     *zap.Logger
	mailSender mailing.Sender
}

func NewSendVerificationEmailConsumer(otpGen *security.OTPGenerator, cache cache.Cache, logger *zap.Logger, sender mailing.Sender) *SendVerificationEmailConsumer {
	return &SendVerificationEmailConsumer{otpGen: otpGen, cache: cache, logger: logger, mailSender: sender}
}

func buildOtpCacheKey(userId ulid.ULID, identityId ulid.ULID, email string) string {
	return fmt.Sprintf("users:%s:identities:%s:email-verification:%s", userId.String(), identityId.String(), email)
}

func (c *SendVerificationEmailConsumer) Handle(message interface{}) {
	c.logger.Info("Received message in consumer",
		zap.String("type", reflect.TypeOf(message).String()))

	otp, err := c.otpGen.GenerateOTP()

	if err != nil {
		c.logger.Error("Failed to generate OTP", zap.Error(err))
		return
	}

	sendEmailVerificationMsg := message.(commands.SendVerificationEmail)

	c.cache.Set(buildOtpCacheKey(sendEmailVerificationMsg.UserId, sendEmailVerificationMsg.IdentityId, sendEmailVerificationMsg.Email), otp)

	// TODO: Use mjml for email templates
	err = c.mailSender.Send(sendEmailVerificationMsg.Email, "Email verification", fmt.Sprintf("Your verification code is: %s", otp))

	if err != nil {
		c.logger.Error("Failed to send email", zap.Error(err))
	}
}
