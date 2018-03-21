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
	ErrVerificationFailed          = errors.New("unable to verify code")
	ErrSatisfyFailed               = errors.New("unable to satisfy code")
	ErrVerificationRequestTimedOut = errors.New("verification request timed out")
)

// VerificationService interface
type VerificationService interface {
	VerifyCode(code string) error
	SatisfyCode(code string) error
}

// Verifier verifies code that comes with a kitty reservation request with
// the verification service.
type Verifier struct {
	client *http.Client
	log    logrus.FieldLogger
}

type VerificationServiceRequest struct {
	Code string `json:"code"`
}

// NewVerifier creates a new verifier instance
func NewVerifier(log logrus.FieldLogger) *Verifier {
	return &Verifier{
		client: &http.Client{
			Timeout: VerifierTimeout,
		},
		log: log,
	}
}

// VerifyCode verifies the reservation code with verfication service
func (v *Verifier) VerifyCode(code string) error {
	verificationRequest, err := json.Marshal(VerificationServiceRequest{
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
		return ErrVerificationRequestTimedOut
	default:
		return ErrVerificationFailed
	}

	return nil
}

// SatisfyCode is called after reservation is completed
func (v *Verifier) SatisfyCode(code string) error {
	satisfyRequest, err := json.Marshal(VerificationServiceRequest{
		code,
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
		return ErrVerificationRequestTimedOut
	default:
		return ErrSatisfyFailed
	}

	return nil
}
