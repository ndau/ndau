package backing

// A managed vars map (map[string]struct{}) is a hash map whose keys are managed variable names.
// If "Something" exists in the map, then the owning nomsifyable struct has a variable named
// ManagedVarSomething that has been set as a result of a transaction being applied.

// Ensure the given map exists and has the given managed variable name set as one of
// its keys.  We use a pointer to a map so that we can assign to it if it's nil on entry.
func ensureManagedVar(m *map[string]struct{}, name string) {
	if *m == nil {
		*m = make(map[string]struct{})
	}
	if _, ok := (*m)[name]; !ok {
		(*m)[name] = struct{}{}
	}
	// To make our nomsified code simpler, include the map itself.
	name = "ManagedVars"
	if _, ok := (*m)[name]; !ok {
		(*m)[name] = struct{}{}
	}
}
