package storage

import (
	"fmt"
	"github.com/ortuman/jackal/storage"
	"github.com/ortuman/jackal/component/pubsub/storage/sql"
)


type IPubSubDao interface {
	tmpFuncNeverUse()
}

var (
	instanceDao IPubSubDao
	isInited bool
)

func Instance() IPubSubDao {
	return instanceDao
}

func InitStorage(config *storage.Config) (error) {

	switch config.Type {
	case storage.BadgerDB:
		// TODO
	case storage.MySQL:
		host := config.MySQL.Host
		user := config.MySQL.User
		pass := config.MySQL.Password
		db := config.MySQL.Database
		poolSize := config.MySQL.PoolSize
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pass, host, db)
		instanceDao = sql.New(dsn, poolSize)
		isInited = true
	case storage.Memory:
		// TODO
	default:
		return fmt.Errorf("storage: unrecognized storage type: %s", config.Type)
	}

	return nil
}