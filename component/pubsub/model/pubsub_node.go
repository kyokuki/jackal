package model

import "time"

type PubSubNode struct {
	ServiceJid    string
	Name          string
	NodeType      string
	Title         string
	Description   string
	CreatorJid    string
	CreatedAt     time.Time
	Configuration string
	CollectionJId string
}
