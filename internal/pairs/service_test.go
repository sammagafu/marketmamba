package pairs

import "testing"

func TestSetSymbolsQuickValidation(t *testing.T) {
	s := &Service{}
	s.cfg = nil
	// AvailableSymbols without cfg uses default
	avail := s.AvailableSymbols()
	if len(avail) < 2 {
		t.Fatal(avail)
	}
}
