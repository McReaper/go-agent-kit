package ta

import (
	"errors"
	"testing"
)

func TestOnlyChecksTheError(t *testing.T) { // want `M6: TestOnlyChecksTheError only fails on an error or a nil`
	_, err := Do()
	if err != nil {
		t.Fatal(err)
	}
}

func TestOnlyChecksNil(t *testing.T) { // want `M6: TestOnlyChecksNil only fails`
	p := new(int)
	if p == nil {
		t.Error("nil")
	}
}

func TestNeverFails(t *testing.T) { // want `M6: TestNeverFails only fails`
	_, _ = Do()
}

func TestChecksTheValue(t *testing.T) {
	n, err := Do()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("n = %d", n)
	}
}

func TestExpectsARefusal(t *testing.T) {
	if _, err := Do(); err == nil {
		t.Fatal("accepted")
	}
}

func TestChecksTheKindOfError(t *testing.T) {
	if _, err := Do(); !errors.Is(err, errors.ErrUnsupported) {
		t.Fatal(err)
	}
}

func TestInASubtest(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		if n, _ := Do(); n != 1 {
			t.Error(n)
		}
	})
}

func helper(t *testing.T) {}

func TestDelegates(t *testing.T) {
	helper(t)
}

func TestMain(m *testing.M) {
	m.Run()
}
