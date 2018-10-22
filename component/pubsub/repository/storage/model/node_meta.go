package model

import "time"

type NodeMeta struct {
	NodeId int64
	NodeConfig string
	Creator string
	CreateDate time.Time
}
