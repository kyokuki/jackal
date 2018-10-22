package sql

import (
	"time"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ortuman/jackal/log"
	"crypto/sha1"
	"fmt"
	"github.com/ortuman/jackal/component/pubsub/repository/storage/interface"
)

// Storage represents a SQL storage sub system.
type Storage struct {
	db     *sql.DB
	doneCh chan chan bool
}

// New returns a SQL storage instance.
func New(dsn string, poolSize int) _interface.IPubSubDao {
	var err error
	s := &Storage{
		doneCh: make(chan chan bool),
	}
	s.db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("%v", err)
	}
	s.db.SetMaxOpenConns(poolSize) // set max opened connection count

	if err := s.db.Ping(); err != nil {
		log.Fatalf("%v", err)
	}
	go s.loop()

	return s
}

// NewMock returns a mocked SQL storage instance.
func NewMock() (*Storage, sqlmock.Sqlmock) {
	var err error
	var sqlMock sqlmock.Sqlmock
	s := &Storage{}
	s.db, sqlMock, err = sqlmock.New()
	if err != nil {
		log.Fatalf("%v", err)
	}
	return s, sqlMock
}

// Close shuts down SQL storage sub system.
func (s *Storage) Close() error {
	ch := make(chan bool)
	s.doneCh <- ch
	<-ch
	return nil
}

func (s *Storage) loop() {
	tc := time.NewTicker(time.Second * 15)
	defer tc.Stop()
	for {
		select {
		case <-tc.C:
			err := s.db.Ping()
			if err != nil {
				log.Error(err)
			}
		case ch := <-s.doneCh:
			s.db.Close()
			close(ch)
			return
		}
	}
}

func (s *Storage) inTransaction(f func(tx *sql.Tx) error) error {
	tx, txErr := s.db.Begin()
	if txErr != nil {
		return txErr
	}
	if err := f(tx); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (s *Storage) Sha1(str string) string {
	sum := sha1.Sum([]byte(str))
	ret := fmt.Sprintf("%x", sum)
	return ret
}


