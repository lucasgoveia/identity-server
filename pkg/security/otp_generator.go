package security

type OTPGenerator struct {
	keyGenerator *SecureKeyGenerator
}

func NewOTPGenerator(keyGenerator *SecureKeyGenerator) *OTPGenerator {
	return &OTPGenerator{keyGenerator: keyGenerator}
}

func (g *OTPGenerator) GenerateOTP() (string, error) {
	return g.keyGenerator.Generate([]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"), 6)
}
