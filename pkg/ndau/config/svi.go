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
