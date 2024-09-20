package mailing

import "go.uber.org/zap"

type StubSender struct {
	logger *zap.Logger
}

func NewStubSender(logger *zap.Logger) *StubSender {

	return &StubSender{
		logger: logger,
	}
}

func (s *StubSender) Send(to, subject, body string) error {
	sugar := s.logger.Sugar()
	sugar.Infof("TO: %s, subject: %s", to, subject)
	sugar.Infof("BODY: %s", body)
	return nil
}
