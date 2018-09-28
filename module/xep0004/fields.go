package xep0004

import (
	"bytes"
)

type Fields struct {
	fields      []*Field
	fieldsByVar map[string]*Field
}

func NewFields() *Fields {
	fields := &Fields{
		fields:      nil,
		fieldsByVar: make(map[string]*Field),
	}

	return fields
}

func (fs *Fields) AddField(field *Field) {
	if field == nil {
		return
	}

	cf, ok := fs.fieldsByVar[field.Var()]
	if ok {
		filtered := fs.fields[:0]
		for _, item := range fs.fields {
			if item != cf {
				filtered = append(filtered, item)
			}
		}
		fs.fields = filtered
	} else {
		fs.fields = append(fs.fields, field)
	}

	if field.Var() != "" {
		fs.fieldsByVar[field.Var()] = field
	}
}

func (fs *Fields) RemoveField(fvar string) *Field {
	var removed *Field
	cf, ok := fs.fieldsByVar[fvar]
	if ok {
		filtered := fs.fields[:0]
		for _, item := range fs.fields {
			if item != cf {
				filtered = append(filtered, item)
			} else {
				removed = item
			}
		}
		fs.fields = filtered
		delete(fs.fieldsByVar, fvar)
	}

	return removed
}

func (fs *Fields) Field(fvar string) *Field {
	cf, ok := fs.fieldsByVar[fvar]
	if ok {
		return cf
	}
	return nil
}

func (fs *Fields) All() []*Field {
	return fs.fields
}

func (fs *Fields) Init() {
	fs.fields = nil
	fs.fieldsByVar = make(map[string]*Field)
}

func (fs *Fields) Clear() {
	fs.fields = nil
	fs.fieldsByVar = make(map[string]*Field)
}

func (fs *Fields) Contains(fvar string) bool {
	_, ok := fs.fieldsByVar[fvar]
	return ok
}

func (fs *Fields) RawString() string {
	var bufferFields bytes.Buffer
	bufferFields.WriteString("fields=[")
	if len(fs.fields) == 0 {
		bufferFields.WriteString("*nil*")
	} else {
		for _, f := range fs.fields {
			bufferFields.WriteString(" ")
			bufferFields.WriteString(f.RawString())
		}
	}
	bufferFields.WriteString("]")

	var bufferFieldsByVar bytes.Buffer
	bufferFieldsByVar.WriteString("fieldsByVar={")
	if len(fs.fields) == 0 {
		bufferFieldsByVar.WriteString("*nil*")
	} else {
		for k, v := range fs.fieldsByVar {
			bufferFieldsByVar.WriteString(" ")
			bufferFieldsByVar.WriteString( k + ":")
			bufferFieldsByVar.WriteString(v.RawString())
		}
	}
	bufferFieldsByVar.WriteString("}")

	return  bufferFields.String() + " " + bufferFieldsByVar.String()
}

func (fs *Fields) SimpleString() string {
	var buffer bytes.Buffer
	if len(fs.fields) == 0 {
		buffer.WriteString("*nil*")
	} else {
		for _, f := range fs.fields {
			buffer.WriteString(f.SimpleString())
		}
	}
	return "Fields{fields=" + buffer.String() + "}"
}
