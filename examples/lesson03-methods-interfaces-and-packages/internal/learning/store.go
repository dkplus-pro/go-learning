package learning

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyName = errors.New("learner name is required")
	ErrNotFound  = errors.New("learner not found")
)

// LearnerStore 是调用方需要的最小能力集合。
// 后续可以把 MemoryStore 替换成数据库实现，而使用方只依赖这个接口。
type LearnerStore interface {
	Save(Learner) error
	FindByName(string) (Learner, error)
}

// MemoryStore 是一个简单的内存实现，适合教学和单元测试。
type MemoryStore struct {
	items map[string]Learner
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]Learner)}
}

func (s *MemoryStore) Save(learner Learner) error {
	if learner.Name == "" {
		return ErrEmptyName
	}

	s.items[learner.Name] = learner
	return nil
}

func (s *MemoryStore) FindByName(name string) (Learner, error) {
	learner, ok := s.items[name]
	if !ok {
		return Learner{}, fmt.Errorf("%w: %s", ErrNotFound, name)
	}

	return learner, nil
}

var _ LearnerStore = (*MemoryStore)(nil)
