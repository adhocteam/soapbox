package soapboxd

import "testing"

func TestGraph(t *testing.T) {
	m := newFsm("start")
	m.addTransition("start", "button_press", "warming", nil)
	m.addTransition("warming", "temperature_too_cool", "warming", nil)
	m.addTransition("warming", "water_boiling", "percolating", nil)
	m.addTransition("percolating", "water_not_empty", "percolating", nil)
	m.addTransition("percolating", "water_tank_empty", "done", nil)

	if err := m.graph(); err != nil {
		t.Fatalf("calling graph: %v", err)
	}
}

func TestHandlers(t *testing.T) {
	m := newFsm("start")
	var log []string

	m.addTransition("start", "a", "state_a", func(s state, e event) error {
		log = append(log, "start <- a -> state_a")
		return nil
	})
	m.addTransition("start", "b", "state_b", func(s state, e event) error {
		log = append(log, "start <- b -> state_b")
		return nil
	})
	m.addTransition("state_a", "b", "state_b", func(s state, e event) error {
		log = append(log, "state_a <- b -> state_b")
		return nil
	})
	m.addTransition("state_b", "c", "state_c", func(s state, e event) error {
		log = append(log, "state_b <- c -> state_c")
		return nil
	})

	if err := m.step("b"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := m.step("c"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	want := []string{"start <- b -> state_b", "state_b <- c -> state_c"}
	if !sliceEqual(want, log) {
		t.Errorf("want %v, got %v", want, log)
	}
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
