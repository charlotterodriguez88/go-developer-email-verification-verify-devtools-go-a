package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type Signup struct {
	Email  string
	UserID string `json:"user_id"`
}

func VerificationLink(base string, user Signup) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/verify-email?user=%s&token=%s", base, user.UserID, hex.EncodeToString(b)), nil
}

func SendVerification(c *InfraiClient, base string, user Signup) (EmailResult, error) {
	link, err := VerificationLink(base, user)
	if err != nil {
		return EmailResult{}, err
	}
	html := fmt.Sprintf("<p>Confirm your developer account:</p><p><a href=\"%s\">Verify email</a></p>", link)
	return c.SendEmail(user.Email, "Verify your developer account", html)
}
