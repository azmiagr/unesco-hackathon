package mail

import "fmt"

func SendVerificationEmail(to, email, otpCode, verificationLink string) error {
	htmlBody, err := RenderVerificationEmail(VerificationEmailData{
		Name:             email,
		Code:             otpCode,
		VerificationLink: verificationLink,
	})
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	subject := "Verifikasi Akun - Kode Kabi"
	return SendEmail(to, subject, htmlBody)

}

func SendAdminLoginOtpEmail(to, email, otpCode string) error {
	htmlBody, err := RenderVerificationEmail(VerificationEmailData{
		Name:             email,
		Code:             otpCode,
		VerificationLink: "",
	})
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	subject := "Kode OTP Login Admin - Kode Kabi"
	return SendEmail(to, subject, htmlBody)
}
