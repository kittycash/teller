package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-errors/errors"
	"github.com/sirupsen/logrus"
)

const (
	// VerifierTimeout is the timeout for requests to verification service
	VerifierTimeout = 10 * time.Second
)

var (
	ErrVerificationFailed    = errors.New("unable to verify code")
	ErrSatisfyFailed         = errors.New("unable to satisfy code")
	ErrVerifyRequestTimedOut = errors.New("verify request timed out")
	ErrSatisfyRequestTimeOut = errors.New("satisfy request timed out")
)

// VerificationService interface
type VerificationService interface {
	VerifyCode(code string) error
	SatisfyCode(code string) error
}

// Verifier verifies code that comes with a kitty reservation request with
// the verification service.
type Verifier struct {
	client  *http.Client
	log     logrus.FieldLogger
	Enabled bool
}

// VerifyRequest is the structure of a verification request
type VerifyRequest struct {
	Code string `json:"code"`
}

// SatisfyRequest is the structure of a satisfy request
type SatisfyRequest struct {
	Code   string `json:"code"`
	KittID string `json:"kitt_id"`
}

// NewVerifier creates a new verifier instance
func NewVerifier(log logrus.FieldLogger, enabled bool) *Verifier {
	return &Verifier{
		client: &http.Client{
			Timeout: VerifierTimeout,
		},
		log:     log,
		Enabled: enabled,
	}
}

// VerifyCode verifies the reservation code with verification service
func (v *Verifier) VerifyCode(code string) error {
	if !v.Enabled {
		return nil
	}

	verificationRequest, err := json.Marshal(VerifyRequest{
		code,
	})
	if err != nil {
		v.log.WithError(err).Error("unable to marshal verification request")
		return err
	}

	// make a verification request to verification service
	resp, err := v.client.Post("/api/verify_code",
		"application/json", bytes.NewBuffer(verificationRequest))
	if err != nil {
		v.log.WithError(err).Error("failed to verify reservation code")
		return err
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		break
	case http.StatusRequestTimeout:
		return ErrVerifyRequestTimedOut
	default:
		return ErrVerificationFailed
	}

	return nil
}

// SatisfyCode is called after reservation is completed
func (v *Verifier) SatisfyCode(code string, kittyID string) error {
	if !v.Enabled {
		return nil
	}

	satisfyRequest, err := json.Marshal(SatisfyRequest{
		code,
		kittyID,
	})
	if err != nil {
		v.log.WithError(err).Error("unable to marshal satisfy request")
		return err
	}

	// make a satisfy request to verification service
	resp, err := v.client.Post("/api/satisfy_code",
		"application/json", bytes.NewBuffer(satisfyRequest))
	if err != nil {
		v.log.WithError(err).Error("failed to verify reservation code")
		return err
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		break
	case http.StatusRequestTimeout:
		return ErrSatisfyRequestTimeOut
	default:
		return ErrSatisfyFailed
	}

	return nil
}
