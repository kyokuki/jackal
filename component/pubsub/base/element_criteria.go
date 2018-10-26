package base

import "github.com/ortuman/jackal/xmpp"

type Criteria interface {
	Matches(element xmpp.XElement) bool
}

type OrElementCriteria struct {
	crits []*ElementCriteria
}

type ElementCriteria struct {
	attrs        map[string]string
	name         string
	cdata        string
	nextCriteria Criteria
}

func (ec *ElementCriteria) SetName(name string) *ElementCriteria {
	ec.name = name
	return ec
}

func (ec *ElementCriteria) AddAttr(attrName string, attrValue string) *ElementCriteria {
	if ec.attrs == nil {
		ec.attrs = make(map[string]string)
	}
	ec.attrs[attrName] = attrValue
	return ec
}

func (ec *ElementCriteria) SetCDATA(cdata string) *ElementCriteria {
	ec.cdata = cdata
	return ec
}

func (ec *ElementCriteria) AddCriteria(next Criteria) *ElementCriteria {
	if ec.nextCriteria == nil {
		ec.nextCriteria = next
	}
	return ec
}

func (ec *ElementCriteria) Matches(element xmpp.XElement) bool {
	if ec.name != "" && ec.name != element.Name() {
		return false
	}

	if ec.cdata == "" || element.Text() != "" && element.Text() == ec.cdata {
		var result bool = true

		for attrKey, attrValue := range ec.attrs {
			if element.Attributes().Get(attrKey) != attrValue {
				result = false
				break
			}
		}

		if ec.nextCriteria != nil {
			subResult := false
			subElements := element.Elements().All()
			for _, elemItem := range subElements {
				if ec.nextCriteria.Matches(elemItem) {
					subResult = true
					break
				}
			}

			result = result && subResult
		}

		return result
	}

	return false
}

func (oec *OrElementCriteria) AddCriteria(next *ElementCriteria) {
	oec.crits = append(oec.crits, next)
}

func (oec *OrElementCriteria) Matches(element xmpp.XElement) bool {
	for _, item := range oec.crits {
		if item.Matches(element) {
			return true
		}
	}
	return false
}