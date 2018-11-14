package model

import "time"

type NodeMeta struct {
	NodeId int64
	Name string
	NodeConfig string
	Creator string
	CreateDate time.Time
}
