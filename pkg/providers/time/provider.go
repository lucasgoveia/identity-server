package time

import "time"

type Provider interface {
	Now() time.Time
	UtcNow() time.Time
}

type DefaultTimeProvider struct {
}

func (d *DefaultTimeProvider) Now() time.Time {
	return time.Now()
}

func (d *DefaultTimeProvider) UtcNow() time.Time {
	return time.Now().UTC()
}
