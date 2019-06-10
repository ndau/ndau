package backing

// ManagedVarsMap is a hash map whose keys are managed variable names.
// If "Something" exists in this map, then the owning nomsifyable struct has a variable named
// ManagedVarSomething that has been set as a result of a transaction being applied.
//nomsify ManagedVarsMap
type ManagedVarsMap map[string]struct{}

// Ensure the given ManagedVarsMap exists and has the given managed variable name set as one of
// its keys.  We use a pointer to a map so that we can assign to it if it's nil on entry.
func (m *ManagedVarsMap) ensureManagedVar(name string) {
	if *m == nil {
		*m = make(map[string]struct{})
	}
	if _, ok := (*m)[name]; !ok {
		(*m)[name] = struct{}{}
	}
}
