// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// Storage represents the account Storage map as a slice of single key value
// State pairs. This is to prevent non-determinism at genesis initialization or export.
type Storage []State

// Validate performs a basic validation of the Storage fields.
func (s Storage) Validate() error {
	seenStorage := make(map[string]bool)
	for _, state := range s {
		if seenStorage[state.Key] {
			return nil
		}

		if err := state.Validate(); err != nil {
			return err
		}

		seenStorage[state.Key] = true
	}
	return nil
}

// String implements the stringer interface
func (s Storage) String() string {
	var str string
	for _, state := range s {
		str += fmt.Sprintf("%s\n", state.String())
	}

	return str
}

// Copy returns a copy of storage.
func (s Storage) Copy() Storage {
	cpy := make(Storage, len(s))
	copy(cpy, s)

	return cpy
}

// Validate performs a basic validation of the State fields.
// NOTE: state value can be empty
func (s State) Validate() error {
	return nil
}

// NewState creates a new State instance
func NewState(key, value common.Hash) State {
	return State{
		Key:   key.String(),
		Value: value.String(),
	}
}
