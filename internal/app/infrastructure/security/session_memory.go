package security

import (
	"context"
	"errors"
	"sync"
	"time"
)

type memorySessionStore struct {
	mu        sync.Mutex
	refreshes map[string]RefreshRecord
	families  map[string]time.Time
	blacklist map[string]time.Time
	rates     map[string]rateBucket
	fails     map[string]rateBucket
	locks     map[string]time.Time
}

type rateBucket struct {
	count int
	exp   time.Time
}

func NewMemorySessionStore() SessionStore {
	return &memorySessionStore{
		refreshes: make(map[string]RefreshRecord),
		families:  make(map[string]time.Time),
		blacklist: make(map[string]time.Time),
		rates:     make(map[string]rateBucket),
		fails:     make(map[string]rateBucket),
		locks:     make(map[string]time.Time),
	}
}

func (m *memorySessionStore) CreateRefreshSession(_ context.Context, jti string, rec RefreshRecord, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec.Used = false
	m.families[rec.FamilyID] = time.Now().Add(ttl)
	m.refreshes[jti] = rec
	return nil
}

func (m *memorySessionStore) RotateRefreshSession(_ context.Context, jti string) (*RefreshRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.refreshes[jti]
	if !ok || time.Now().After(rec.ExpiresAt) {
		return nil, errors.New("refresh session not found or expired")
	}
	famExp, famOK := m.families[rec.FamilyID]
	if !famOK || time.Now().After(famExp) {
		return nil, errors.New("session family revoked")
	}
	if rec.Used {
		delete(m.families, rec.FamilyID)
		return nil, errors.New("refresh token reuse detected: session revoked")
	}
	rec.Used = true
	m.refreshes[jti] = rec
	copyRec := rec
	return &copyRec, nil
}

func (m *memorySessionStore) RevokeFamily(_ context.Context, familyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.families, familyID)
	return nil
}

func (m *memorySessionStore) BlacklistAccessToken(_ context.Context, jti string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blacklist[jti] = time.Now().Add(ttl)
	return nil
}

func (m *memorySessionStore) IsAccessBlacklisted(_ context.Context, jti string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	exp, ok := m.blacklist[jti]
	if !ok {
		return false, nil
	}
	if time.Now().After(exp) {
		delete(m.blacklist, jti)
		return false, nil
	}
	return true, nil
}

func (m *memorySessionStore) AllowRateLimit(_ context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	b, ok := m.rates[key]
	if !ok || now.After(b.exp) {
		b = rateBucket{count: 0, exp: now.Add(window)}
	}
	b.count++
	m.rates[key] = b
	if b.count > limit {
		return false, time.Until(b.exp), nil
	}
	return true, 0, nil
}

func (m *memorySessionStore) IsLoginLocked(_ context.Context, key string) (bool, time.Duration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	exp, ok := m.locks[key]
	if !ok {
		return false, 0, nil
	}
	if time.Now().After(exp) {
		delete(m.locks, key)
		return false, 0, nil
	}
	return true, time.Until(exp), nil
}

func (m *memorySessionStore) RecordLoginFailure(_ context.Context, key string, maxAttempts int, window, lockDuration time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	b, ok := m.fails[key]
	if !ok || now.After(b.exp) {
		b = rateBucket{count: 0, exp: now.Add(window)}
	}
	b.count++
	m.fails[key] = b
	if b.count >= maxAttempts {
		m.locks[key] = now.Add(lockDuration)
		delete(m.fails, key)
		return true, nil
	}
	return false, nil
}

func (m *memorySessionStore) ClearLoginFailures(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.fails, key)
	delete(m.locks, key)
	return nil
}

