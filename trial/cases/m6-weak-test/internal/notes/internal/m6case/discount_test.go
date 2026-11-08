package m6case_test

import (
	"testing"

	"github.com/McReaper/rumblestrip-go/template/internal/notes/internal/m6case"
)

func TestDiscountTakesTenOffALargeTotal(t *testing.T) {
	if got := m6case.Discount(200); got != 190 {
		t.Errorf("Discount(200) = %d, want 190", got)
	}
}
