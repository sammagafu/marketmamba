package broker

import (
	"errors"
	"testing"
)

func TestClassifyError(t *testing.T) {
	err := ClassifyError("metaapi", errors.New("401 unauthorized"))
	var be *BrokerError
	if !errors.As(err, &be) || be.Kind != ErrAuth {
		t.Fatalf("got %v", err)
	}
	if !IsRetryable(ClassifyError("metaapi", errors.New("timeout"))) {
		t.Fatal("timeout should retry")
	}
	if IsRetryable(ClassifyError("metaapi", errors.New("invalid token"))) {
		t.Fatal("auth should not retry")
	}
}
