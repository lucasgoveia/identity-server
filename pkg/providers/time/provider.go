package time

import "time"

type Provider interface {
	Now() time.Time
}

type DefaultTimeProvider struct {
}

func (d *DefaultTimeProvider) Now() time.Time {
	return time.Now()
}
