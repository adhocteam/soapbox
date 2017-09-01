package soapboxd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type state string

type event string

type handler func(state, event) error

type transition struct {
	src     state
	event   event
	dest    state
	handler handler
}

// fsm is a finite state machine.
type fsm struct {
	state state

	// transitions is a lookup table mapping states to new states
	// via events.
	transitions []*transition
}

func newFsm(initial string) *fsm {
	s := &fsm{
		state: state(initial),
	}
	return s
}

func (m *fsm) lookup(s state, e event) *transition {
	for _, t := range m.transitions {
		if t.src == s && t.event == e {
			return t
		}
	}
	return nil
}

func (m *fsm) step(e event) error {
	t := m.lookup(m.state, e)
	// an invalid transition, meaning user never defined a
	// transition from state state via this event.
	if t == nil {
		return fmt.Errorf("transition from state %s via event %s not found", m.state, e)
	}
	if t.handler != nil {
		if err := t.handler(m.state, e); err != nil {
			m.state = "error"
			return err
		}
	}
	m.state = t.dest
	return nil
}

func (m *fsm) addTransition(src state, e event, dest state, handler handler) {
	t := m.lookup(src, e)
	if t == nil {
		m.transitions = append(m.transitions, &transition{
			src:     src,
			event:   e,
			dest:    dest,
			handler: handler,
		})
		return
	}
	panic("transition already defined")
}

// dot generates Dot language description of the machine for
// generating graph images from Graphviz.
func (m *fsm) dot() string {
	var buf bytes.Buffer
	name := "deployment"
	fmt.Fprintf(&buf, "digraph %s {\n", name)
	for _, t := range m.transitions {
		fmt.Fprintf(&buf, "\t%s -> %s [label=\"%s\"];\n", t.src, t.dest, t.event)
	}
	buf.WriteString("}")
	return buf.String()
}

// graph generates a PNG representation of the state machine graph
// using dot/Graphviz.
func (m *fsm) graph() error {
	cmd := exec.Command("dot", "-Tpng", "-o", "deployment_fsm.png")
	cmd.Stdin = strings.NewReader(m.dot())
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
