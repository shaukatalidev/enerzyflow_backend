package auth

import (
	"fmt"
	"log"
	"math/rand"
	"net/smtp"
	"os"
	"sync"
	"time"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type OTPEntry struct {
	Code      string
	ExpiresAt time.Time
}

var otpStore = struct {
	sync.RWMutex
	m map[string]OTPEntry
}{m: make(map[string]OTPEntry)}

type smtpCred struct {
	From     string
	Password string
	Host     string
	Port     string
}

func getSMTPCred() smtpCred {
    return smtpCred{
        From:     os.Getenv("SMTP_FROM"),
        Password: os.Getenv("SMTP_PASSWORD"),
        Host:     os.Getenv("SMTP_HOST"),
        Port:     os.Getenv("SMTP_PORT"),
    }
}



const emailOtpTTL = 5 * time.Minute

func generateEmailOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func sendEmailWithCustomSMTP(to string, otp string) error {
	cred := getSMTPCred()
	if cred.From == "" || cred.Password == "" {
        log.Fatal("SMTP credentials not set in environment")
    }

	subject := "Verify Your OTP"
	body := fmt.Sprintf("Your OTP is: %s", otp)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
		cred.From, to, subject, body,
	))

	auth := smtp.PlainAuth("", cred.From, cred.Password, cred.Host)
	return smtp.SendMail(cred.Host+":"+cred.Port, auth, cred.From, []string{to}, msg)
}

func sendEmailWithSendGrid(toEmail, otp string) error {
    from := mail.NewEmail("EnerzyFlow", os.Getenv("SENDGRID_FROM"))
    to := mail.NewEmail("", toEmail)
    subject := "Verify Your OTP"
    plainTextContent := fmt.Sprintf("Your OTP is: %s", otp)
    htmlContent := fmt.Sprintf("<p>Your OTP is: <b>%s</b></p>", otp)

    message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
    client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
    response, err := client.Send(message)
    if err != nil {
        return err
    }

    if response.StatusCode >= 400 {
        return fmt.Errorf("sendgrid error: status %d, body: %s", response.StatusCode, response.Body)
    }

    return nil
}

func keyForOTP(email, role string) string {
	return email + "|" + role
}


func SendOTP(email, role string) (string, error) {
	otp := generateEmailOTP()
	if err := sendEmailWithSendGrid(email, otp); err != nil {
		log.Printf("Failed to send email to %s: %v", email, err)
		return "", err
	}

	otpStore.Lock()
	otpStore.m[keyForOTP(email, role)] = OTPEntry{Code: otp, ExpiresAt: time.Now().Add(emailOtpTTL)}
	otpStore.Unlock()

	return otp, nil
}

// VerifyOTP checks OTP for given email + role and TTL
func VerifyOTP(email, role, otp string) (valid bool, expired bool, err error) {
	otpKey := keyForOTP(email, role)

	otpStore.RLock()
	entry, exists := otpStore.m[otpKey]
	otpStore.RUnlock()

	if !exists {
		return false, false, nil
	}

	if time.Now().After(entry.ExpiresAt) {
		otpStore.Lock()
		delete(otpStore.m, otpKey)
		otpStore.Unlock()
		return false, true, nil
	}

	if entry.Code != otp {
		return false, false, nil
	}

	// OTP is valid; delete after verification
	otpStore.Lock()
	delete(otpStore.m, otpKey)
	otpStore.Unlock()

	return true, false, nil
}

