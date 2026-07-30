package librepayment

import (
	"testing"
	"time"
)

func TestMemoryStorageAddsInitialStatusHistory(t *testing.T) {
	createdAt := time.Date(2026, 7, 30, 10, 11, 12, 0, time.UTC)
	stor := NewMemoryStorage()

	err := stor.Add(createdAt, "payment-1", 150.25, "merchant-1", map[string]string{"order": "42"})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	payment, err := stor.Get("payment-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if len(payment.StatusHistory) != 1 {
		t.Fatalf("len(StatusHistory) = %d, want 1", len(payment.StatusHistory))
	}

	record := payment.StatusHistory[0]
	if !record.EventAt.Equal(createdAt) {
		t.Fatalf("record.EventAt = %v, want %v", record.EventAt, createdAt)
	}

	if record.Status != StatusNew {
		t.Fatalf("record.Status = %v, want %v", record.Status, StatusNew)
	}
}

func TestMemoryStorageCloneCopiesStatusHistory(t *testing.T) {
	createdAt := time.Date(2026, 7, 30, 10, 11, 12, 0, time.UTC)
	stor := NewMemoryStorage()

	err := stor.Add(createdAt, "payment-1", 150.25, "merchant-1", nil)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	firstCopy, err := stor.Get("payment-1")
	if err != nil {
		t.Fatalf("Get() first copy error = %v", err)
	}

	firstCopy.StatusHistory[0].Status = StatusConfirmed

	secondCopy, err := stor.Get("payment-1")
	if err != nil {
		t.Fatalf("Get() second copy error = %v", err)
	}

	if secondCopy.StatusHistory[0].Status != StatusNew {
		t.Fatalf("stored history was mutated through clone: got %v, want %v", secondCopy.StatusHistory[0].Status, StatusNew)
	}
}
