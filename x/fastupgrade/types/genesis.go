package types

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{Params: DefaultParams()}
}

// Validate performs stateless genesis validation. Gov compatibility is checked
// at runtime instead, so an export still imports after gov params change.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
