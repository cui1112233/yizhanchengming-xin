package agent

import (
	"fmt"
	"sync"
)

// State describes the public lifecycle of the Windows giant-material agent.
// The OCR worker is intentionally not represented as a per-job process state:
// it remains alive while the agent returns to idle between jobs.
type State string

type BindingState string

const (
	StateIdle             State = "idle"
	StatePairing          State = "pairing"
	StateDownloadingModel State = "downloading_model"
	StateReady            State = "ready"
	StateRunning          State = "running"
	StateCleaning         State = "cleaning"
	StateUploading        State = "uploading"
	StateFailed           State = "failed"
	StateStopping         State = "stopping"
)

const (
	BindingUnpaired     BindingState = "unpaired"
	BindingConnecting   BindingState = "connecting"
	BindingOnline       BindingState = "online"
	BindingOffline      BindingState = "offline"
	BindingNeedsPairing BindingState = "needs_pairing"
)

type Progress struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
	Percent   int `json:"percent"`
}

type Snapshot struct {
	State          State        `json:"state"`
	BindingState   BindingState `json:"bindingState"`
	JobID          string       `json:"jobId,omitempty"`
	ModelVersion   string       `json:"modelVersion,omitempty"`
	Progress       Progress     `json:"progress"`
	ErrorCode      string       `json:"errorCode,omitempty"`
	ErrorMessage   string       `json:"errorMessage,omitempty"`
	WorkerResident bool         `json:"workerResident"`
}

type StateMachine struct {
	mu       sync.RWMutex
	snapshot Snapshot
}

var allowedTransitions = map[State]map[State]bool{
	StateIdle: {
		StatePairing:          true,
		StateDownloadingModel: true,
		StateReady:            true,
		// The OCR worker remains resident after completing a task. The next
		// claimed job therefore starts directly from idle instead of requiring
		// a synthetic ready transition.
		StateRunning:  true,
		StateStopping: true,
	},
	StatePairing: {
		StateIdle:             true,
		StateDownloadingModel: true,
		StateReady:            true,
		StateFailed:           true,
		StateStopping:         true,
	},
	StateDownloadingModel: {
		StateReady:    true,
		StateFailed:   true,
		StateStopping: true,
	},
	StateReady: {
		StateRunning:          true,
		StatePairing:          true,
		StateDownloadingModel: true,
		StateFailed:           true,
		StateStopping:         true,
	},
	StateRunning: {
		StateCleaning: true,
		StateFailed:   true,
		StateStopping: true,
	},
	StateCleaning: {
		StateUploading: true,
		StateFailed:    true,
		StateStopping:  true,
	},
	StateUploading: {
		StateIdle:     true,
		StateFailed:   true,
		StateStopping: true,
	},
	StateFailed: {
		StateIdle:             true,
		StatePairing:          true,
		StateDownloadingModel: true,
		StateReady:            true,
		StateRunning:          true,
		StateStopping:         true,
	},
	StateStopping: {},
}

func NewStateMachine() *StateMachine {
	return &StateMachine{snapshot: Snapshot{State: StateIdle, BindingState: BindingUnpaired, WorkerResident: true}}
}

func (m *StateMachine) Transition(next State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !allowedTransitions[m.snapshot.State][next] {
		return fmt.Errorf("invalid agent state transition %q -> %q", m.snapshot.State, next)
	}
	m.snapshot.State = next
	if next == StateStopping {
		m.snapshot.WorkerResident = false
	}
	if next == StateIdle {
		m.snapshot.JobID = ""
		m.snapshot.Progress = Progress{}
		m.snapshot.ErrorCode = ""
		m.snapshot.ErrorMessage = ""
		m.snapshot.WorkerResident = true
	}
	return nil
}

func (m *StateMachine) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.snapshot
}

func (m *StateMachine) ResetToIdle() {
	m.mu.Lock()
	defer m.mu.Unlock()
	bindingState := m.snapshot.BindingState
	if bindingState == "" {
		bindingState = BindingUnpaired
	}
	m.snapshot = Snapshot{State: StateIdle, BindingState: bindingState, WorkerResident: true}
}

func (m *StateMachine) SetJob(jobID, modelVersion string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshot.JobID = jobID
	m.snapshot.ModelVersion = modelVersion
}

func (m *StateMachine) SetProgress(progress Progress) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshot.Progress = progress
}

func (m *StateMachine) SetBindingState(state BindingState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshot.BindingState = state
}
