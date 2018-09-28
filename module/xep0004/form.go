package xep0004

import "github.com/ortuman/jackal/xmpp"

type Form struct {
	instruction string
	title       string
	ftype       string

	fields Fields
}

func New(ftype string, title string, instruction string) *Form {
	f := &Form{
		instruction: instruction,
		title:       title,
		ftype:       ftype,
	}
	f.fields.Init()
	return f
}

func NewFormFromElement(element *xmpp.Element) *Form {
	form := &Form{}
	form.fields.Init()

	ftype := element.Attributes().Get("type")
	if ftype != "" {
		form.ftype = ftype
	}

	for _, itemElem := range element.Elements().All() {
		switch itemElem.Name() {
		case "title":
			form.SetTitle(itemElem.Text())
		case "instructions":
			form.SetInstruction(itemElem.Text())
		case "field":
			sub, ok := itemElem.(*xmpp.Element)
			if ok {
				field := NewFieldFromElement(sub)
				form.fields.AddField(field)
			}
		}
	}

	return form
}

func (f *Form) Element() *xmpp.Element {
	form := xmpp.NewElementNamespace("x", "jabber:x:data")

	if f.ftype != "" {
		form.SetAttribute("type", f.ftype)
	}

	if f.title != "" {
		newElem := xmpp.NewElementName("title")
		newElem.SetText(f.title)
		form.AppendElement(newElem)
	}

	if f.instruction != "" {
		newElem := xmpp.NewElementName("instructions")
		newElem.SetText(f.instruction)
		form.AppendElement(newElem)
	}

	for _, fieldItem := range f.fields.All() {
		form.AppendElement(fieldItem.Element())
	}

	return form
}

func (f *Form) Instruction() string {
	return f.instruction
}

func (f *Form) SetInstruction(instruction string) {
	f.instruction = instruction
}

func (f *Form) Title() string {
	return f.title
}

func (f *Form) SetTitle(title string) {
	f.title = title
}

func (f *Form) Type() string {
	return f.ftype
}

func (f *Form) SetType(ftype string) {
	f.ftype = ftype
}

func (f *Form) Contains(fvar string) bool {
	return f.fields.Contains(fvar)
}

func (f *Form) AddField(field *Field) {
	f.fields.AddField(field)
}

func (f *Form) RemoveField(fvar string) *Field {
	return f.fields.RemoveField(fvar)
}
func (f *Form) Field(fvar string) *Field {
	return f.fields.Field(fvar)
}
func (f *Form) AllFields() []*Field {
	return f.fields.All()
}

