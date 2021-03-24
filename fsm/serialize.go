package fsm

import (
	"encoding/json"
)

// MarshalText implements the TextMarshaler interface and is used by various
// encoding packages. necessary to marshal FSM to json.
func (m *FSM) MarshalText() ([]byte, error) {
	b, err := json.Marshal(
		struct {
			Current     string            `json:"current"`
			Transitions map[eParts]string `json:"transitions"`
		}{
			Current:     m.current,
			Transitions: m.transitions,
		},
	)

	return b, err
}

// UnmarshalText implements the TextUnmarshaler interface and is used by various
// encoding packages. necessary to unmarshal FSM from JSON.
func (m *FSM) UnmarshalText(b []byte) error {
	fsm := struct {
		Current     string            `json:"current"`
		Transitions map[eParts]string `json:"transitions"`
	}{}

	err := json.Unmarshal(b, &fsm)
	if err != nil {
		return err
	}

	m.current = fsm.Current
	m.transitions = fsm.Transitions

	return nil
}

// MarshalText implements the TextMarshaler interface and is used by various
// encoding packages. Necessary to marshal eParts to JSON.
func (e eParts) MarshalText() ([]byte, error) {
	b, err := json.Marshal(
		struct {
			Event string `json:"event"`
			Src   string `json:"src"`
		}{
			Event: e.event,
			Src:   e.src,
		},
	)

	return b, err
}

// UnmarshalText implements the TextUnmarshaler interface and is used by various
// encoding packages. Necessary to unmarshal eParts from JSON.
func (e *eParts) UnmarshalText(b []byte) error {
	eParts := struct {
		Event string `json:"event"`
		Src   string `json:"src"`
	}{}

	err := json.Unmarshal(b, &eParts)
	if err != nil {
		return err
	}

	e.src = eParts.Src
	e.event = eParts.Event

	return nil
}
