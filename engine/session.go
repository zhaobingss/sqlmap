package engine

import (
	"database/sql"
	"errors"
)

const commitInitStatus = 1

var initError = errors.New("not init correctly")

/// session会话，管理事务
type Session struct {
	db          *sql.DB // 数据库连接
	tx          *sql.Tx // 事务
	commit      int8    // 本session提交的次数
	canRollback bool    // 标记是否可以回滚
	init        bool    // 是否初始化
}

/// 创建session
func newSession(db *sql.DB) *Session {
	return &Session{
		db:   db,
		init: true,
	}
}

/// 开始事务
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

/// 回滚事务
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

/// 提交事务
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

/// 非查询语句
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

/// 查询语句
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