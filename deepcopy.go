package sessions

type CopyableMap map[string]interface{}
type CopyableSlice []interface{}

// DeepCopy will create a deep copy of this map.
// Both maps and slices will be considered when making the copy.
func (m CopyableMap) DeepCopy() map[string]interface{} {
	result := map[string]interface{}{}

	for k, v := range m {
		// Handle maps
		mapValue, isMap := v.(map[string]interface{})
		if isMap {
			result[k] = CopyableMap(mapValue).DeepCopy()
			continue
		}

		// Handle slices
		sliceValue, isSlice := v.([]interface{})
		if isSlice {
			result[k] = CopyableSlice(sliceValue).DeepCopy()
			continue
		}

		result[k] = v
	}

	return result
}

// DeepCopy will create a deep copy of this slice.
// Both maps and slices will be considered when making the copy.
func (s CopyableSlice) DeepCopy() []interface{} {
	var result []interface{}

	for _, v := range s {
		// Handle maps
		mapValue, isMap := v.(map[string]interface{})
		if isMap {
			result = append(result, CopyableMap(mapValue).DeepCopy())
			continue
		}

		// Handle slices
		sliceValue, isSlice := v.([]interface{})
		if isSlice {
			result = append(result, CopyableSlice(sliceValue).DeepCopy())
			continue
		}

		result = append(result, v)
	}

	return result
}
