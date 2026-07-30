package librepayment

import "testing"

func TestLibrePaymentStatusReturnsStatusHistory(t *testing.T) {
	payments := &LibrePayment{stor: NewMemoryStorage()}

	id, err := payments.Register("merchant-1", 150.25, nil)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err = payments.Confirm(id)
	if err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}

	payment, err := payments.Status(id)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	if len(payment.StatusHistory) != 2 {
		t.Fatalf("len(StatusHistory) = %d, want 2", len(payment.StatusHistory))
	}

	if payment.StatusHistory[0].Status != StatusNew {
		t.Fatalf("first status = %v, want %v", payment.StatusHistory[0].Status, StatusNew)
	}

	if payment.StatusHistory[1].Status != StatusConfirmed {
		t.Fatalf("second status = %v, want %v", payment.StatusHistory[1].Status, StatusConfirmed)
	}

	if payment.StatusHistory[1].EventAt.IsZero() {
		t.Fatalf("second status EventAt is zero")
	}
}

func TestLibrePaymentForceSetSameStatusDoesNotAppendHistory(t *testing.T) {
	payments := &LibrePayment{stor: NewMemoryStorage()}

	id, err := payments.Register("merchant-1", 150.25, nil)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err = payments.ForceSetStatus(id, StatusNew)
	if err != nil {
		t.Fatalf("ForceSetStatus() error = %v", err)
	}

	payment, err := payments.Status(id)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	if len(payment.StatusHistory) != 1 {
		t.Fatalf("len(StatusHistory) = %d, want 1", len(payment.StatusHistory))
	}

	if payment.StatusHistory[0].Status != StatusNew {
		t.Fatalf("first status = %v, want %v", payment.StatusHistory[0].Status, StatusNew)
	}
}
