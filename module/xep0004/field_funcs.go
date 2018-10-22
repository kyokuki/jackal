package xep0004

import "github.com/pkg/errors"

func NewFieldBool(fvar string, value bool, label string) Field {
	field := Field{
		Type:  Boolean,
		Label: label,
		Var:   fvar,
	}

	if value {
		field.Values = []string{"1"}
	} else {
		field.Values = []string{"0"}
	}

	return field
}

func NewFieldFixed(value string) Field {
	field := Field{
		Type:   Fixed,
		Values: []string{value},
	}

	return field
}

func NewFieldHidden(fvar string, value string) Field {
	field := Field{
		Type:   Hidden,
		Var:    fvar,
		Values: []string{value},
	}

	return field
}

func NewFieldJidMulti(fvar string, values []string, label string) Field {
	field := Field{
		Type:   JidMulti,
		Var:    fvar,
		Label:  label,
		Values: values,
	}

	return field
}

func NewFieldJidSingle(fvar string, value string, label string) Field {
	field := Field{
		Type:   JidSingle,
		Var:    fvar,
		Label:  label,
		Values: []string{value},
	}

	return field
}

func NewFieldListMulti(fvar string, values []string, label string, optionsLabel []string, optionsValue []string) (Field, error) {
	if len(optionsLabel) != 0 && len(optionsLabel) != len(optionsValue) {
		return Field{}, errors.New("Invalid optionsLabel and optionsValue length")
	}

	var options []Option
	for i := 0; i < len(optionsLabel); i++ {
		options = append(options, Option{optionsLabel[i], optionsValue[i]})
	}

	field := Field{
		Type:    ListMulti,
		Var:     fvar,
		Label:   label,
		Values:  values,
		Options: options,
	}

	return field, nil
}

func NewFieldListSingle(fvar string, value string, label string, optionsLabel []string, optionsValue []string) (Field, error) {
	if len(optionsLabel) != 0 && len(optionsLabel) != len(optionsValue) {
		return Field{}, errors.New("Invalid optionsLabel and optionsValue length")
	}

	var options []Option
	for i := 0; i < len(optionsLabel); i++ {
		options = append(options, Option{optionsLabel[i], optionsValue[i]})
	}

	field := Field{
		Type:    ListSingle,
		Var:     fvar,
		Label:   label,
		Values:  []string{value},
		Options: options,
	}

	return field, nil
}

func NewFieldTextMulti(fvar string, values []string, label string) Field {
	field := Field{
		Type:   TextMulti,
		Var:    fvar,
		Label:  label,
		Values: values,
	}

	return field
}

func NewFieldTextSingle(fvar string, value string, label string) Field {
	field := Field{
		Type:   TextSingle,
		Var:    fvar,
		Label:  label,
		Values: []string{value},
	}

	return field
}

func NewFieldTextPrivate(fvar string, value string, label string) Field {
	field := Field{
		Type:   TextPrivate,
		Var:    fvar,
		Label:  label,
		Values: []string{value},
	}

	return field
}