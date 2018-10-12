package xep0004

func (f *DataForm) AddField(field Field) {
	if len(field.Var) == 0 && len(field.Values) == 0 {
		return
	}

	targetIndex := -1
	for idx, itemField := range f.Fields {
		if len(itemField.Var) > 0 && itemField.Var == field.Var {
			targetIndex = idx
		}
	}

	// replace or append
	if targetIndex >= 0 {
		filtered := f.Fields[:0]
		filtered = append(filtered, f.Fields[:targetIndex]...)
		filtered = append(filtered, field)
		filtered = append(filtered, f.Fields[targetIndex+1:]...)
		f.Fields = filtered
	} else {
		f.Fields = append(f.Fields, field)
	}
}

func (f *DataForm) RemoveField(variable string) (pos int, retfield Field) {
	targetIndex := -1
	targetField := Field{}
	if len(variable) == 0 {
		return targetIndex, Field{}
	}

	for idx, itemField := range f.Fields {
		if len(itemField.Var) > 0 && itemField.Var ==variable {
			targetIndex = idx
		}
	}

	// remove
	if targetIndex >= 0 {
		targetField = f.Fields[targetIndex]
		filtered := f.Fields[:0]
		filtered = append(filtered, f.Fields[:targetIndex]...)
		filtered = append(filtered, f.Fields[targetIndex+1:]...)
		f.Fields = filtered
	}
	return targetIndex, targetField
}

// return pos >= 0 if found
func (f *DataForm) Field(variable string) (pos int, field Field) {
	targetIndex := -1
	if len(variable) == 0 {
		return targetIndex, Field{}
	}

	for idx, itemField := range f.Fields {
		if len(itemField.Var) > 0 && itemField.Var ==variable {
			targetIndex = idx
		}
	}
	return targetIndex, f.Fields[targetIndex]
}

func (f *DataForm) Contains(variable string) bool {
	targetIndex := -1
	if len(variable) == 0 {
		return false
	}

	for idx, itemField := range f.Fields {
		if len(itemField.Var) > 0 && itemField.Var == variable {
			targetIndex = idx
		}
	}
	return targetIndex >= 0
}