package broker

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorKind classifies broker failures for user-facing messages and retry policy.
type ErrorKind string

const (
	ErrAuth       ErrorKind = "auth"
	ErrSymbol     ErrorKind = "symbol"
	ErrMargin     ErrorKind = "margin"
	ErrRateLimit  ErrorKind = "rate_limit"
	ErrUnavailable ErrorKind = "unavailable"
	ErrValidation ErrorKind = "validation"
	ErrUnknown    ErrorKind = "unknown"
)

// BrokerError wraps a broker failure with a stable kind.
type BrokerError struct {
	Kind    ErrorKind
	Message string
	Cause   error
}

func (e *BrokerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *BrokerError) Unwrap() error { return e.Cause }

func NewBrokerError(kind ErrorKind, msg string, cause error) *BrokerError {
	return &BrokerError{Kind: kind, Message: msg, Cause: cause}
}

// ClassifyError maps raw broker/API errors to BrokerError.
func ClassifyError(provider string, err error) error {
	if err == nil {
		return nil
	}
	var be *BrokerError
	if errors.As(err, &be) {
		return err
	}
	msg := strings.ToLower(err.Error())
	kind := ErrUnknown
	userMsg := err.Error()

	switch {
	case strings.Contains(msg, "unauthorized"), strings.Contains(msg, "401"),
		strings.Contains(msg, "invalid token"), strings.Contains(msg, "authentication"),
		strings.Contains(msg, "credentials"):
		kind = ErrAuth
		userMsg = "Broker authentication failed — check API token and account credentials"
	case strings.Contains(msg, "symbol"), strings.Contains(msg, "instrument"),
		strings.Contains(msg, "unknown market"), strings.Contains(msg, "not found") && strings.Contains(msg, "symbol"):
		kind = ErrSymbol
		userMsg = "Symbol not available on this broker — check your pair list"
	case strings.Contains(msg, "margin"), strings.Contains(msg, "insufficient"),
		strings.Contains(msg, "not enough money"), strings.Contains(msg, "funds"):
		kind = ErrMargin
		userMsg = "Insufficient margin or balance for this trade"
	case strings.Contains(msg, "rate limit"), strings.Contains(msg, "429"), strings.Contains(msg, "too many"):
		kind = ErrRateLimit
		userMsg = "Broker rate limit — try again in a moment"
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "deploy"), strings.Contains(msg, "not connected"):
		kind = ErrUnavailable
		if provider == "metaapi" {
			userMsg = "MetaAPI account not ready — first connect can take 1–3 minutes; try Test again"
		} else {
			userMsg = "Broker temporarily unavailable — try again"
		}
	}

	return NewBrokerError(kind, userMsg, err)
}

// IsRetryable returns whether the executor should retry the operation.
func IsRetryable(err error) bool {
	var be *BrokerError
	if errors.As(err, &be) {
		switch be.Kind {
		case ErrRateLimit, ErrUnavailable:
			return true
		default:
			return false
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") || strings.Contains(msg, "429")
}
