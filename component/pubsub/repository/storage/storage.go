package storage

import (
	"github.com/ortuman/jackal/component/pubsub/repository/storage/sql"
	"github.com/ortuman/jackal/component/pubsub/repository/storage/interface"
)


var (
	instanceDao _interface.IPubSubDao
	isInited bool
)

func Instance() _interface.IPubSubDao {
	return instanceDao
}

func InitStorage(mysql string) (error) {
	instanceDao = sql.New(mysql, 16)
	isInited = true
	return nil
}