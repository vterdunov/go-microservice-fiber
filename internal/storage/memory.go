package storage

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vterdunov/go-microservice-fiber/internal/model"
)

var (
	ErrNotFound = errors.New("user not found")
)

type MemoryStorage struct {
	data   sync.Map
	nextID atomic.Uint64
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

func (s *MemoryStorage) Create(req model.CreateUserRequest) model.User {
	id := s.nextID.Add(1)
	user := model.User{
		ID:        id,
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}
	s.data.Store(id, user)
	return user
}

func (s *MemoryStorage) Get(id uint64) (model.User, error) {
	val, ok := s.data.Load(id)
	if !ok {
		return model.User{}, ErrNotFound
	}
	return val.(model.User), nil
}

func (s *MemoryStorage) GetAll() []model.User {
	users := make([]model.User, 0)
	s.data.Range(func(_, value any) bool {
		users = append(users, value.(model.User))
		return true
	})
	return users
}

func (s *MemoryStorage) Update(id uint64, req model.UpdateUserRequest) (model.User, error) {
	val, ok := s.data.Load(id)
	if !ok {
		return model.User{}, ErrNotFound
	}
	user := val.(model.User)
	user.Name = req.Name
	user.Email = req.Email
	s.data.Store(id, user)
	return user, nil
}

func (s *MemoryStorage) Delete(id uint64) error {
	_, ok := s.data.Load(id)
	if !ok {
		return ErrNotFound
	}
	s.data.Delete(id)
	return nil
}
