package agent

import "testing"

func TestSuccessfulJobReturnsToIdleAndKeepsWorkerResident(t *testing.T) {
	machine := NewStateMachine()
	for _, state := range []State{StateReady, StateRunning, StateCleaning, StateUploading} {
		if err := machine.Transition(state); err != nil {
			t.Fatal(err)
		}
	}
	if err := machine.Transition(StateIdle); err != nil {
		t.Fatal(err)
	}
	if got := machine.Snapshot().State; got != StateIdle {
		t.Fatalf("state=%s", got)
	}
	if !machine.Snapshot().WorkerResident {
		t.Fatal("expected OCR worker to remain resident")
	}
}

func TestIdleWorkerCanStartTheNextClaimedJob(t *testing.T) {
	machine := NewStateMachine()
	for _, state := range []State{StateReady, StateRunning, StateCleaning, StateUploading, StateIdle, StateRunning} {
		if err := machine.Transition(state); err != nil {
			t.Fatalf("transition to %s: %v", state, err)
		}
	}
	if got := machine.Snapshot().State; got != StateRunning {
		t.Fatalf("state=%s, want %s", got, StateRunning)
	}
}

func TestInvalidTransitionDoesNotSilentlyStopTheAgent(t *testing.T) {
	machine := NewStateMachine()
	if err := machine.Transition(StateUploading); err == nil {
		t.Fatal("expected invalid transition")
	}
	if got := machine.Snapshot().State; got != StateIdle {
		t.Fatalf("state=%s", got)
	}
}

func TestFailedJobCanRecoverToIdleAfterWorkerRestart(t *testing.T) {
	machine := NewStateMachine()
	for _, state := range []State{StateReady, StateRunning, StateFailed, StateIdle} {
		if err := machine.Transition(state); err != nil {
			t.Fatal(err)
		}
	}
	if got := machine.Snapshot().State; got != StateIdle {
		t.Fatalf("state=%s", got)
	}
}

func TestStateMachineTracksPersistentBindingState(t *testing.T) {
	machine := NewStateMachine()
	if got := machine.Snapshot().BindingState; got != BindingUnpaired {
		t.Fatalf("initial binding state=%s, want %s", got, BindingUnpaired)
	}
	machine.SetBindingState(BindingOnline)
	if got := machine.Snapshot().BindingState; got != BindingOnline {
		t.Fatalf("binding state=%s, want %s", got, BindingOnline)
	}
}

func TestResetToIdlePreservesPersistentBindingState(t *testing.T) {
	machine := NewStateMachine()
	machine.SetBindingState(BindingOnline)
	machine.ResetToIdle()
	if got := machine.Snapshot().BindingState; got != BindingOnline {
		t.Fatalf("binding state after reset=%s, want %s", got, BindingOnline)
	}
}
