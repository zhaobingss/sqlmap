package engine

import (
	"database/sql"
	"errors"
)

const commitInitStatus = 1

var initError = errors.New("not init correctly")

/// session that manage the transaction
type Session struct {
	db          *sql.DB // the database/sql.DB
	tx          *sql.Tx // transaction
	commit      int8    // commit count
	canRollback bool    // flag that indicate if the transaction can rollback
	init        bool    // flag indicate if the session is already init
}

/// create a session with sql.DB
/// @param db: sql.DB
func newSession(db *sql.DB) *Session {
	return &Session{
		db:   db,
		init: true,
	}
}

/// begin a transaction
func (s *Session) BeginTx() error {
	if !s.init {
		return initError
	}
	s.canRollback = true
	if s.tx != nil {
		s.commit++
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	s.tx = tx
	s.commit = commitInitStatus
	return nil
}

/// rollback the transaction
func (s *Session) Rollback() error {
	if !s.init {
		return initError
	}
	if s.tx == nil || !s.canRollback {
		return nil
	}

	err := s.tx.Rollback()
	if err != nil {
		return err
	}

	s.tx = nil
	return nil
}

/// commit the transaction
func (s *Session) Commit() error {
	if !s.init {
		return initError
	}
	s.canRollback = false
	if s.tx == nil {
		return nil
	}

	if s.commit != commitInitStatus {
		s.commit--
		return nil
	}

	err := s.tx.Commit()
	if err != nil {
		return err
	}

	s.tx = nil
	return nil
}

/// execute the sql with a can ignore result
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
func (s *Session) Exec(key string, data interface{}) (sql.Result, error) {
	if !s.init {
		return nil, initError
	}
	if s.tx == nil {
		return exec(key, data, s.db.Exec)
	} else {
		return exec(key, data, s.tx.Exec)
	}
}

/// execute the sql and set result to []map[string]string
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
func (s *Session) Query(key string, data interface{}) ([]map[string]string, error) {
	if !s.init {
		return nil, initError
	}
	if s.tx == nil {
		return query(key, data, s.db.Query)
	} else {
		return query(key, data, s.tx.Query)
	}
}

/// execute sql and set the result to a slice dest
/// @param the result will be set to dest, and the dest must be like eg: *[]*struct or *[]struct
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
func (s *Session) Select(dest interface{}, key string, param interface{}) error  {
	if !s.init {
		return initError
	}
	if s.tx == nil {
		return selectRows(dest, key, param, s.db.Query)
	} else {
		return selectRows(dest, key, param, s.tx.Query)
	}
}

/// execute sql and set the result to a struct dest
/// @param the result will be set to dest, and the dest must be like eg: *struct
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
/// @return error: ERR_NOT_GOT_RECORD,ERR_MORE_THAN_ONE_RECORD,...
/// ERR_NOT_GOT_RECORD indicate that not got any recode from the database
/// ERR_MORE_THAN_ONE_RECORD indicate that got more than one record from database
func (s *Session) SelectOne(dest interface{}, key string, param interface{}) error  {
	if !s.init {
		return initError
	}
	if s.tx == nil {
		return selectRow(dest, key, param, s.db.Query)
	} else {
		return selectRow(dest, key, param, s.tx.Query)
	}
}