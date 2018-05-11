package config

// AsNamespacedKey transforms a SVINamespacedKey into a NamespacedKey
func (snk *SVINamespacedKey) AsNamespacedKey() NamespacedKey {
	return NamespacedKey{
		Namespace: NewB64Data(snk.GetNamespace()),
		Key:       NewB64Data(snk.GetKey()),
	}
}

// GetNamespacedKey gets the appropriate namespaced key given a deferred change
func (dc *SVIDeferredChange) GetNamespacedKey(currentBlock uint64) *SVINamespacedKey {
	if currentBlock >= dc.GetUpdateBlock() {
		return dc.GetFuture()
	}
	return dc.GetCurrent()
}

// GetNamespacedKey gets the appropriate namespaced key from an indirect value
func (iv *SVIIndirectValue) GetNamespacedKey(currentBlock uint64) *SVINamespacedKey {
	switch v := iv.GetValue().(type) {
	case *SVIIndirectValue_Simple:
		return v.Simple
	case *SVIIndirectValue_Deferred:
		return v.Deferred.GetNamespacedKey(currentBlock)
	default:
		return nil
	}
}

// GetNamespacedKey gets the requested namespaced key from a SVI map
func (m *SVIMap) GetNamespacedKey(key string, currentBlock uint64) *SVINamespacedKey {
	innerMap := m.GetMap()
	if innerMap == nil {
		return nil
	}
	return innerMap[key].GetNamespacedKey(currentBlock)
}

func (m *SVIMap) init() {
	if m == nil {
		*m = SVIMap{}
	}
	if m.Map == nil {
		m.Map = make(map[string]*SVIIndirectValue)
	}
}

// Set sets the given name to the specified NamedpacedKey
func (m *SVIMap) Set(name string, nsk NamespacedKey) {
	m.init()
	m.Map[name] = &SVIIndirectValue{
		Value: &SVIIndirectValue_Simple{
			Simple: &SVINamespacedKey{
				Namespace: nsk.Namespace.Bytes(),
				Key:       nsk.Key.Bytes(),
			},
		},
	}
}

// SetOn sets the given name to the specified NamespacedKey on the given block.
//
// If name already exists, a Deferred value is created updating on the
// specified block. Otherwise, a simple value is created.
func (m *SVIMap) SetOn(name string, nsk NamespacedKey, block uint64) {
	m.init()
	if currentIndirect, hasCurrent := m.Map[name]; hasCurrent {
		// the current block is almost certainly not actually 0, but
		// choosing that value should unconditionally get the current value
		current := currentIndirect.GetNamespacedKey(0)
		m.Map[name] = &SVIIndirectValue{
			Value: &SVIIndirectValue_Deferred{
				Deferred: &SVIDeferredChange{
					Current: current,
					Future: &SVINamespacedKey{
						Namespace: nsk.Namespace.Bytes(),
						Key:       nsk.Key.Bytes(),
					},
					UpdateBlock: block,
				},
			},
		}
	} else {
		m.Set(name, nsk)
	}
}

// SystemStore types are stores of system variables.
//
// No restriction is placed on their implementation, so long as they
// can get values from namespaced keys.
type SystemStore interface {
	Get(namespace, key []byte) ([]byte, error)
}

// GetNSK gets the requested namespaced key from any SystemStore
func GetNSK(ss SystemStore, nsk NamespacedKey) (out []byte, err error) {
	out, err = ss.Get(nsk.Namespace.Bytes(), nsk.Key.Bytes())
	return
}

// GetSVI returns the System Variable Indirection map from any SystemStore
func GetSVI(ss SystemStore, nsk NamespacedKey) (*SVIMap, error) {
	svib, err := GetNSK(ss, nsk)
	if err != nil {
		return nil, err
	}
	svi := new(SVIMap)
	err = svi.Unmarshal(svib)
	return svi, nil
}
