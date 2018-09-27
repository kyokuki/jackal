package form

import (
	"errors"
	"github.com/ortuman/jackal/xmpp"
	"bytes"
	"fmt"
)

type Field struct {
	description  string
	label        string
	optionLabels []string
	optionValues []string
	required     bool
	ftype        FieldType
	values       []string
	fvar         string
}

func NewFieldBool(fvar string, value bool, label string) *Field {
	field := &Field{
		ftype: FieldType_Bool,
		label: label,
		fvar:  fvar,
	}

	if value {
		field.values = []string{"1"}
	} else {
		field.values = []string{"0"}
	}

	return field
}

func NewFieldFixed(value string) *Field {
	field := &Field{
		ftype:  FieldType_Fixed,
		values: []string{value},
	}

	return field
}

func NewFieldHidden(fvar string, value string) *Field {
	field := &Field{
		ftype:  FieldType_Hidden,
		fvar:   fvar,
		values: []string{value},
	}

	return field
}

func NewFieldJidMulti(fvar string, values []string, label string) *Field {
	field := &Field{
		ftype:  FieldType_JidMulti,
		fvar:   fvar,
		label:  label,
		values: values,
	}

	return field
}

func NewFieldJidSingle(fvar string, value string, label string) *Field {
	field := &Field{
		ftype:  FieldType_JidSingle,
		fvar:   fvar,
		label:  label,
		values: []string{value},
	}

	return field
}

func NewFieldListMulti(fvar string, values []string, label string, optionsLabel []string, optionsValue []string) (*Field, error) {
	if len(optionsLabel) != 0 && len(optionsLabel) != len(optionsValue) {
		return nil, errors.New("Invalid optionsLabel and optinsValue length")
	}
	field := &Field{
		ftype:        FieldType_ListMulti,
		fvar:         fvar,
		label:        label,
		values:       values,
		optionLabels: optionsLabel,
		optionValues: optionsValue,
	}

	return field, nil
}

func NewFieldListSingle(fvar string, value string, label string, optionsLabel []string, optionsValue []string) (*Field, error) {
	if len(optionsLabel) != 0 && len(optionsLabel) != len(optionsValue) {
		return nil, errors.New("Invalid optionsLabel and optinsValue length")
	}
	field := &Field{
		ftype:        FieldType_ListSingle,
		fvar:         fvar,
		label:        label,
		values:       []string{value},
		optionLabels: optionsLabel,
		optionValues: optionsValue,
	}

	return field, nil
}

func NewFieldTextMulti(fvar string, values []string, label string) *Field {
	field := &Field{
		ftype:  FieldType_TextMulti,
		fvar:   fvar,
		label:  label,
		values: values,
	}

	return field
}

func NewFieldTextSingle(fvar string, value string, label string) *Field {
	field := &Field{
		ftype:  FieldType_TextSingle,
		fvar:   fvar,
		label:  label,
		values: []string{value},
	}

	return field
}

func NewFieldTextPrivate(fvar string, value string, label string) *Field {
	field := &Field{
		ftype:  FieldType_TextPrivate,
		fvar:   fvar,
		label:  label,
		values: []string{value},
	}

	return field
}

func NewFieldFromElement(element *xmpp.Element) *Field {
	if element.Name() != "field" {
		return nil
	}

	field := &Field{}
	field.fvar = element.Attributes().Get("var")
	ftype := element.Attributes().Get("type")
	if ftype == "" {
		field.ftype = FieldType_TextSingle
	} else {
		field.ftype = getFieldTypeByName(ftype)
	}
	field.label = element.Attributes().Get("label")

	descElem := element.Elements().Child("desc")
	if descElem != nil {
		field.description = descElem.Text()
	}

	requiredElem := element.Elements().Child("required")
	if requiredElem != nil {
		field.required = true
	}

	var (
		valueList    []string
		optionValues []string
		optionLabels []string
	)
	for _, child := range element.Elements().All() {
		if "value" == child.Name() && child.Text() != "" {
			valueList = append(valueList, child.Text())
		}

		if "option" == child.Name() {
			optionLabels = append(optionLabels, child.Attributes().Get("label"))
			vElem := child.Elements().Child("value")
			if vElem != nil {
				optionValues = append(optionValues, vElem.Text())
			}
		}
	}

	field.values = valueList
	field.optionValues = optionValues
	field.optionLabels = optionLabels

	return field
}

func (f *Field) Element() *xmpp.Element {
	return f.getElement(true, true)
}

func (f *Field) getElement(ftype, label bool) *xmpp.Element {
	field := xmpp.NewElementName("field")

	if f.fvar != "" {
		field.SetAttribute("var", f.fvar)
	}

	if ftype && f.ftype != "" {
		field.SetAttribute("type", f.ftype.String())
	}

	if label && f.label != "" {
		field.SetAttribute("label", f.label)
	}

	if f.description != "" {
		newElem := xmpp.NewElementName("desc")
		newElem.SetText(f.description)
		field.AppendElement(newElem)
	}

	if f.required {
		newElem := xmpp.NewElementName("required")
		field.AppendElement(newElem)
	}

	for _, valueItem := range f.values {
		newElem := xmpp.NewElementName("value")
		newElem.SetText(valueItem)
		field.AppendElement(newElem)
	}

	for optionValueIdx, optionValueItem := range f.optionValues {
		optionElem := xmpp.NewElementName("option")
		if optionValueIdx < len(f.optionLabels) {
			optionElem.SetAttribute("label", f.optionLabels[optionValueIdx])
		}

		optionValueElem := xmpp.NewElementName("value")
		if optionValueItem != "" {
			optionValueElem.SetText(optionValueItem)
		}
		optionElem.AppendElement(optionValueElem)
		field.AppendElement(optionElem)
	}

	return field
}

func (f *Field) Description() string {
	return f.description
}

func (f *Field) SetDescription(description string) {
	f.description = description
}

func (f Field) Label() string {
	return f.label
}

func (f *Field) SetLabel(label string) {
	f.label = label
}

func (f *Field) OptionLabels() []string {
	return f.optionLabels
}

func (f *Field) SetOptionLabels(optionLabels []string) {
	f.optionLabels = optionLabels
}

func (f *Field) OptionValues() []string {
	return f.optionValues
}

func (f *Field) SetOptionValues(optionValues []string) {
	f.optionValues = optionValues
}

func (f *Field) Type() FieldType {
	return f.ftype
}

func (f *Field) SetType(ftype FieldType) {
	f.ftype = ftype
}

func (f *Field) Values() []string {
	return f.values
}

func (f *Field) SetValues(values []string) {
	f.values = values
}

func (f *Field) Var() string {
	return f.fvar
}

func (f *Field) SetVar(fvar string) {
	f.fvar = fvar
}

func (f *Field) Required() bool {
	return f.required
}

func (f *Field) SetRequired(required bool) {
	f.required = required
}

func (f *Field) SimpleString() string {
	var buffer bytes.Buffer
	if len(f.values) == 0 {
		buffer.WriteString("*nil*")
	} else {
		for _, val := range f.values {
			if val != "" {
				buffer.WriteString("[")
				buffer.WriteString(val)
				buffer.WriteString("]")
			}
		}
	}
	return f.Var() + "=" + buffer.String()
}

func (f *Field) RawString() string {
	return fmt.Sprintf("%+v", f)
}

type FieldType string

func (ft FieldType) String() string {
	return string(ft)
}

func getFieldTypeByName(ftype string) FieldType {
	return FieldType(ftype)
}

const (
	FieldType_Bool        = FieldType("boolean")
	FieldType_Fixed       = FieldType("fixed")
	FieldType_Hidden      = FieldType("hidden")
	FieldType_JidSingle   = FieldType("jid_single")
	FieldType_JidMulti    = FieldType("jid_multi")
	FieldType_ListMulti   = FieldType("list_multi")
	FieldType_ListSingle  = FieldType("list_single")
	FieldType_TextMulti   = FieldType("text_multi")
	FieldType_TextPrivate = FieldType("text_private")
	FieldType_TextSingle  = FieldType("text_single")
)
