package mailing

type Sender interface {
	Send(to, subject, body string) error
}
