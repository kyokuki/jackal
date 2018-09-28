package xep0004

import (
	"testing"
	"strings"
	"github.com/stretchr/testify/require"
	"github.com/ortuman/jackal/xmpp"
)

func TestField_NewFieldFromElement(t *testing.T) {
	docSrc := ""
	docSrc = `<field var="maxsubs" type="list-single" label="Maximum number of subscribers">` +
		`<value>20</value>` +
		`<option label="10"><value>10</value></option>` +
		`<option label="20"><value>20</value></option>` +
		`<option label="30"><value>30</value></option>` +
		`<option label="50"><value>50</value></option>` +
		`<option label="100"><value>100</value></option>` +
		`<option label="None"><value>none</value></option>` +
		`</field>`

	p := xmpp.NewParser(strings.NewReader(docSrc), xmpp.DefaultMode, 0)
	parent, err := p.ParseElement()

	require.Nil(t, err)
	require.NotNil(t, parent)

	elem, ok := parent.(*xmpp.Element)
	require.Equal(t, true, ok)

	field := NewFieldFromElement(elem)
	require.NotNil(t, field)

	require.Equal(t, "list-single", field.ftype.String())
	require.Equal(t, "Maximum number of subscribers", field.label)
	require.Equal(t, "maxsubs", field.fvar)
	require.Equal(t, []string{"20"}, field.values)
	require.Equal(t, []string{"10", "20", "30", "50", "100", "None"}, field.optionLabels)
	require.Equal(t, []string{"10", "20", "30", "50", "100", "none"}, field.optionValues)
}

func TestField_Element(t *testing.T) {
	docSrc := ""
	docSrc = `<field var="maxsubs" type="list-single" label="Maximum number of subscribers">` +
		`<value>20</value>` +
		`<option label="10"><value>10</value></option>` +
		`<option label="20"><value>20</value></option>` +
		`<option label="30"><value>30</value></option>` +
		`<option label="50"><value>50</value></option>` +
		`<option label="100"><value>100</value></option>` +
		`<option label="None"><value>none</value></option>` +
		`</field>`

	p := xmpp.NewParser(strings.NewReader(docSrc), xmpp.DefaultMode, 0)
	parent, err := p.ParseElement()

	require.Nil(t, err)
	require.NotNil(t, parent)

	elem, ok := parent.(*xmpp.Element)
	require.Equal(t, true, ok)

	field := NewFieldFromElement(elem)
	require.NotNil(t, field)

	toElem := field.Element()
	require.Equal(t, docSrc, toElem.String())
}
