package base

import (
	"strconv"
	"github.com/ortuman/jackal/module/xep0004"
)

type LeafNodeConfig struct {
	abstractNodeConfig
}

type CollectionNodeConfig struct {
	abstractNodeConfig
}

type DefaultNodeConfig struct {
	LeafNodeConfig
}

func NewLeafNodeConfig(nodeName string) *LeafNodeConfig {
	lf := &LeafNodeConfig{}
	lf.init(nodeName)
	return lf
}

func (lf *LeafNodeConfig) IsPersistItem() bool {
	exist, filed := lf.Form().Field("pubsub#persist_items")
	if exist < 0 {
		return false
	}

	if len(filed.Values) > 0 && filed.Values[0] == "1" {
		return true
	}
	return false
}

func (lf *LeafNodeConfig) MaxItems() int {
	value := ""
	exist, filed := lf.Form().Field("pubsub#max_items")
	if exist < 0 {
		return 0
	}

	if len(filed.Values) > 0 {
		value = filed.Values[0]
	}

	ret, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return  ret
}

func NewCollectionNodeConfig(nodeName string) *CollectionNodeConfig {
	cf := &CollectionNodeConfig{}
	cf.init(nodeName)
	cf.form.AddField(xep0004.NewFieldTextMulti("pubsub#children", []string{""}, ""))
	return cf
}

func (cf *CollectionNodeConfig) SetChildren(children []string) {
	cf.form.AddField(xep0004.Field{
		Var:"pubsub#children",
		Values:children,
	})
}

func NewDefaultNodeConfig(nodeName string) *DefaultNodeConfig {
	df := &DefaultNodeConfig{}
	df.init(nodeName)
	df.form.AddField(xep0004.NewFieldTextMulti("pubsub#children", []string{""}, ""))
	return df
}