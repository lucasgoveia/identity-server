package services

import (
	"identity-server/internal/domain"
	"net/http"
)

func IdentifyDevice(req *http.Request) *domain.Device {
	userAgent := req.Header.Get("User-Agent")
	ip := req.RemoteAddr

	return &domain.Device{
		UserAgent: userAgent,
		IpAddress: ip,
	}
}
