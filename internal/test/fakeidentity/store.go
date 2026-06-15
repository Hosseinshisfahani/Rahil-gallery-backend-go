package fakeidentity

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type Store struct {
	mu            sync.RWMutex
	roles         map[shared.ID]*identity.Role
	rolesByName   map[string]*identity.Role
	users         map[shared.ID]*identity.User
	usersByEmail  map[string]*identity.User
	refreshTokens map[shared.ID]*identity.RefreshToken
	refreshByHash map[string]*identity.RefreshToken
}

func NewStore() *Store {
	s := &Store{
		roles:         make(map[shared.ID]*identity.Role),
		rolesByName:   make(map[string]*identity.Role),
		users:         make(map[shared.ID]*identity.User),
		usersByEmail:  make(map[string]*identity.User),
		refreshTokens: make(map[shared.ID]*identity.RefreshToken),
		refreshByHash: make(map[string]*identity.RefreshToken),
	}
	customer := &identity.Role{
		ID:          uuid.New(),
		Name:        identity.RoleCustomer,
		Description: "Store customer",
	}
	s.roles[customer.ID] = customer
	s.rolesByName[customer.Name] = customer
	return s
}

func (s *Store) SeedUser(user *identity.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := *user
	s.users[user.ID] = &copy
	if user.Email != nil {
		s.usersByEmail[normalize(*user.Email)] = &copy
	}
}

func normalize(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

type UserRepo struct{ S *Store }

func (r UserRepo) FindByID(ctx context.Context, id shared.ID) (*identity.User, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	u, ok := r.S.users[id]
	if !ok || u.DeletedAt != nil {
		return nil, shared.ErrNotFound
	}
	copy := *u
	return &copy, nil
}

func (r UserRepo) FindByEmail(ctx context.Context, email string) (*identity.User, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	u, ok := r.S.usersByEmail[normalize(email)]
	if !ok || u.DeletedAt != nil {
		return nil, shared.ErrNotFound
	}
	copy := *u
	return &copy, nil
}

func (r UserRepo) FindByPhone(_ context.Context, _ string) (*identity.User, error) {
	return nil, shared.ErrNotFound
}

func (r UserRepo) Create(ctx context.Context, user *identity.User) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if user.Email != nil {
		email := normalize(*user.Email)
		if _, exists := r.S.usersByEmail[email]; exists {
			return shared.ErrConflict
		}
	}
	copy := *user
	r.S.users[user.ID] = &copy
	if user.Email != nil {
		r.S.usersByEmail[normalize(*user.Email)] = &copy
	}
	return nil
}

func (r UserRepo) Update(ctx context.Context, user *identity.User) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.users[user.ID]; !ok {
		return shared.ErrNotFound
	}
	copy := *user
	r.S.users[user.ID] = &copy
	if user.Email != nil {
		r.S.usersByEmail[normalize(*user.Email)] = &copy
	}
	return nil
}

func (r UserRepo) SoftDelete(ctx context.Context, id shared.ID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	u, ok := r.S.users[id]
	if !ok || u.DeletedAt != nil {
		return shared.ErrNotFound
	}
	now := time.Now()
	u.DeletedAt = &now
	return nil
}

type RoleRepo struct{ S *Store }

func (r RoleRepo) FindByID(ctx context.Context, id shared.ID) (*identity.Role, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	role, ok := r.S.roles[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	copy := *role
	return &copy, nil
}

func (r RoleRepo) FindByName(ctx context.Context, name string) (*identity.Role, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	role, ok := r.S.rolesByName[name]
	if !ok {
		return nil, shared.ErrNotFound
	}
	copy := *role
	return &copy, nil
}

type RefreshRepo struct{ S *Store }

func (r RefreshRepo) Create(ctx context.Context, token *identity.RefreshToken) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	copy := *token
	r.S.refreshTokens[token.ID] = &copy
	r.S.refreshByHash[token.TokenHash] = &copy
	return nil
}

func (r RefreshRepo) FindByHash(ctx context.Context, hash string) (*identity.RefreshToken, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	t, ok := r.S.refreshByHash[hash]
	if !ok {
		return nil, shared.ErrNotFound
	}
	copy := *t
	return &copy, nil
}

func (r RefreshRepo) Revoke(ctx context.Context, id shared.ID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	t, ok := r.S.refreshTokens[id]
	if !ok || t.RevokedAt != nil {
		return shared.ErrNotFound
	}
	now := time.Now()
	t.RevokedAt = &now
	return nil
}
