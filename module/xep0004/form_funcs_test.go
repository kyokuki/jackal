package xep0004

import (
"testing"
"strings"
"github.com/stretchr/testify/require"
"github.com/ortuman/jackal/xmpp"
)

func TestForm(t *testing.T) {
	docSrc := ""+
		`<x xmlns="jabber:x:data" type="form">` +
		`<title>Bot Configuration</title>` +
		`<instructions>Fill out this form to configure your new bot!</instructions>` +
		`<field var="FORM_TYPE" type="hidden">` +
		`<value>jabber:bot</value>` +
		`</field>` +
		`<field type="fixed"><value>Section 1: Bot Info</value></field>` +
		`<field var="botname" type="text-single" label="The name of your bot"/>` +
		`<field var="description" type="text-multi" label="Helpful description of your bot"/>` +
		`<field var="public" type="boolean" label="Public bot?">` +
		`<required/>` +
		`</field>` +
		`<field var="password" type="text-private" label="Password for special access"/>` +
		`<field type="fixed"><value>Section 2: Features</value></field>` +
		`<field var="features" type="list-multi" label="What features will the bot support?">` +
		`<value>news</value>` +
		`<value>search</value>` +
		`<option label="Contests"><value>contests</value></option>` +
		`<option label="News"><value>news</value></option>` +
		`<option label="Polls"><value>polls</value></option>` +
		`<option label="Reminders"><value>reminders</value></option>` +
		`<option label="Search"><value>search</value></option>` +
		`</field>` +
		`<field type="fixed"><value>Section 3: Subscriber List</value></field>` +
		`<field var="maxsubs" type="list-single" label="Maximum number of subscribers">` +
		`<value>20</value>` +
		`<option label="10"><value>10</value></option>` +
		`<option label="20"><value>20</value></option>` +
		`<option label="30"><value>30</value></option>` +
		`<option label="50"><value>50</value></option>` +
		`<option label="100"><value>100</value></option>` +
		`<option label="None"><value>none</value></option>` +
		`</field>` +
		`<field type="fixed"><value>Section 4: Invitations</value></field>` +
		`<field var="invitelist" type="jid-multi" label="People to invite">` +
		`<desc>Tell all your friends about your new bot!</desc>` +
		`</field>` +
		`</x>`

	p := xmpp.NewParser(strings.NewReader(docSrc), xmpp.DefaultMode, 0)
	parent, err := p.ParseElement()
	require.Nil(t, err)
	require.NotNil(t, parent)

	elem, ok := parent.(*xmpp.Element)
	require.Equal(t, true, ok)

	// test Form.NewFormFromElement
	form, err := NewFormFromElement(elem)
	require.NotNil(t, form)

	// test Form.Elements()
	//toElem := form.Element()
	//require.Equal(t, docSrc, toElem.String())

	// test Form.Field
	//_, maxsubsField := form.Field("maxsubs")
	//fmt.Println(maxsubsField)

	// test Form.Field
	//_, featuresField := form.Field("features")
	//fmt.Println(featuresField)

	// test Form.Contains
	require.Equal(t, true, form.Contains("features"))
	require.Equal(t, false, form.Contains("no-exist"))

	// test Form.RemoveField
	_, removed := form.RemoveField("features")
	require.Equal(t, false, form.Contains("features"))

	// test Form.AddField
	form.AddField(removed)
	require.Equal(t, true, form.Contains("features"))
}