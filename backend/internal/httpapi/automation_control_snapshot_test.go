package httpapi

import "testing"

func TestAutomationControlCallbackUsesPrivateSnapshot(t *testing.T) {
	c, b, _ := priorityFixture(t)
	_, err := c.control("alice", b.ID, "paused", func(job map[string]any) error {
		job["callbackProbe"] = "private"
		c.mu.Lock()
		state, loadErr := c.loadLocked()
		if loadErr == nil && state.Jobs["job-active"]["callbackProbe"] != nil {
			t.Error("callback mutated the shared job before persistence")
		}
		c.mu.Unlock()
		return loadErr
	})
	if err != nil {
		t.Fatal(err)
	}
}
