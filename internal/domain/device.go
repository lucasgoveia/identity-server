package domain

type Device struct {
	IpAddress string
	UserAgent string
}

func NewDevice(ipAddress string, userAgent string) *Device {
	return &Device{
		IpAddress: ipAddress,
		UserAgent: userAgent,
	}
}
