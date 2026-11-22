package batchfactoryexecution

import (
	"errors"
	"testing"
)

func TestExecutionTransition(t *testing.T) {
	tests := []struct {
		from ExecutionStatus
		to   ExecutionStatus
		want bool
	}{
		{ExecutionPlanned, ExecutionRunning, true},
		{ExecutionRunning, ExecutionCompleted, true},
		{ExecutionRunning, ExecutionNeedsAttention, true},
		{ExecutionRunning, ExecutionCancelled, true},
		{ExecutionPlanned, ExecutionCompleted, false},
		{ExecutionCompleted, ExecutionRunning, false},
		{ExecutionCancelled, ExecutionRunning, false},
		{ExecutionRunning, ExecutionPlanned, false},
		{ExecutionRunning, ExecutionRunning, false},
		{ExecutionRunning, "unknown", false},
		{"", ExecutionRunning, false},
	}
	for _, test := range tests {
		t.Run(string(test.from)+"_to_"+string(test.to), func(t *testing.T) {
			if got := test.from.CanTransition(test.to); got != test.want {
				t.Errorf("CanTransition(%q, %q) = %t, want %t", test.from, test.to, got, test.want)
			}
		})
	}
}

func TestBookTransition(t *testing.T) {
	tests := []struct {
		from BookStatus
		to   BookStatus
		want bool
	}{
		{BookPending, BookRunning, true},
		{BookRunning, BookWaitingExternal, true},
		{BookRunning, BookCompleted, true},
		{BookRunning, BookRetryWait, true},
		{BookRunning, BookNeedsAttention, true},
		{BookRunning, BookCancelled, true},
		{BookWaitingExternal, BookCompleted, true},
		{BookWaitingExternal, BookRetryWait, true},
		{BookWaitingExternal, BookNeedsAttention, true},
		{BookWaitingExternal, BookCancelled, true},
		{BookRetryWait, BookRunning, true},
		{BookRetryWait, BookNeedsAttention, true},
		{BookRetryWait, BookCancelled, true},
		{BookPending, BookCompleted, false},
		{BookWaitingExternal, BookRunning, false},
		{BookRetryWait, BookCompleted, false},
		{BookCompleted, BookRunning, false},
		{BookCancelled, BookRunning, false},
		{BookRunning, BookPending, false},
		{BookRunning, BookRunning, false},
		{BookRunning, "unknown", false},
		{"", BookRunning, false},
	}
	for _, test := range tests {
		t.Run(string(test.from)+"_to_"+string(test.to), func(t *testing.T) {
			if got := test.from.CanTransition(test.to); got != test.want {
				t.Errorf("CanTransition(%q, %q) = %t, want %t", test.from, test.to, got, test.want)
			}
		})
	}
}

func TestStageTransition(t *testing.T) {
	tests := []struct {
		from StageStatus
		to   StageStatus
		want bool
	}{
		{StagePending, StageQueued, true},
		{StageQueued, StageRunning, true},
		{StageRunning, StageSucceeded, true},
		{StageRunning, StageReconciling, true},
		{StageRunning, StageRetryWait, true},
		{StageRunning, StageTerminalFailed, true},
		{StageRunning, StageCancelled, true},
		{StageReconciling, StageSucceeded, true},
		{StageReconciling, StageRetryWait, true},
		{StageReconciling, StageTerminalFailed, true},
		{StageReconciling, StageCancelled, true},
		{StageRetryWait, StageQueued, true},
		{StageRetryWait, StageRunning, true},
		{StageRetryWait, StageTerminalFailed, true},
		{StageRetryWait, StageCancelled, true},
		{StagePending, StageRunning, false},
		{StageQueued, StageSucceeded, false},
		{StageReconciling, StageRunning, false},
		{StageRetryWait, StageSucceeded, false},
		{StageSucceeded, StageRunning, false},
		{StageTerminalFailed, StageRunning, false},
		{StageCancelled, StageRunning, false},
		{StageRunning, StageRunning, false},
		{StageRunning, "unknown", false},
		{"", StageRunning, false},
	}
	for _, test := range tests {
		t.Run(string(test.from)+"_to_"+string(test.to), func(t *testing.T) {
			if got := test.from.CanTransition(test.to); got != test.want {
				t.Errorf("CanTransition(%q, %q) = %t, want %t", test.from, test.to, got, test.want)
			}
		})
	}
}

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		name  string
		valid error
		bad   error
	}{
		{"execution", ExecutionPlanned.ValidateTransition(ExecutionRunning), ExecutionCancelled.ValidateTransition(ExecutionRunning)},
		{"book", BookRetryWait.ValidateTransition(BookRunning), BookCompleted.ValidateTransition(BookRunning)},
		{"stage", StageReconciling.ValidateTransition(StageSucceeded), StageTerminalFailed.ValidateTransition(StageRunning)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.valid != nil {
				t.Errorf("valid transition returned %v", test.valid)
			}
			if !errors.Is(test.bad, ErrInvalidTransition) {
				t.Errorf("invalid transition returned %v, want ErrInvalidTransition", test.bad)
			}
		})
	}
}

func TestDeriveExecutionStatus(t *testing.T) {
	tests := []struct {
		name  string
		books []Book
		want  ExecutionStatus
	}{
		{"empty", nil, ExecutionPlanned},
		{"all pending", []Book{{Status: BookPending}, {Status: BookPending}}, ExecutionPlanned},
		{"running alongside a failed book", []Book{{Status: BookNeedsAttention}, {Status: BookRunning}}, ExecutionRunning},
		{"running before a failed book", []Book{{Status: BookRunning}, {Status: BookNeedsAttention}}, ExecutionRunning},
		{"pending alongside a failed book", []Book{{Status: BookNeedsAttention}, {Status: BookPending}}, ExecutionRunning},
		{"waiting on external work", []Book{{Status: BookCompleted}, {Status: BookWaitingExternal}}, ExecutionRunning},
		{"waiting to retry", []Book{{Status: BookCompleted}, {Status: BookRetryWait}}, ExecutionRunning},
		{"needs attention after other books finish", []Book{{Status: BookNeedsAttention}, {Status: BookCompleted}, {Status: BookCancelled}}, ExecutionNeedsAttention},
		{"all completed", []Book{{Status: BookCompleted}, {Status: BookCompleted}}, ExecutionCompleted},
		{"completed with an explicit cancellation", []Book{{Status: BookCompleted}, {Status: BookCancelled}}, ExecutionCompleted},
		{"all cancelled", []Book{{Status: BookCancelled}, {Status: BookCancelled}}, ExecutionCancelled},
		{"unknown book state cannot complete", []Book{{Status: BookCompleted}, {Status: "unexpected"}}, ExecutionNeedsAttention},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := DeriveExecutionStatus(test.books); got != test.want {
				t.Errorf("DeriveExecutionStatus() = %q, want %q", got, test.want)
			}
		})
	}
}
