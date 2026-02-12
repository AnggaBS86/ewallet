package service

import (
	"strconv"
	"sync"
	"time"

	"ewallet/internal/dto"
	cache "github.com/AnggaBS86/gocachemem"
)

// HistoryCache is a local-process cache for transaction history responses.
// It is intentionally simple (in-memory) and must be invalidated on balance-changing operations.
type HistoryCache interface {
	Get(userID uint, page, limit int) (*dto.TransactionHistoryResponse, bool)
	Set(userID uint, page, limit int, resp *dto.TransactionHistoryResponse)
	InvalidateUser(userID uint)
}

type historyCacheEntry struct {
	value     *dto.TransactionHistoryResponse
	expiresAt time.Time
}

// inMemoryHistoryCache is a best-effort cache backed by gocachemem.
// It is safe for concurrent use.
type inMemoryHistoryCache struct {
	mu       sync.RWMutex
	ttl      time.Duration
	store    *cache.Cache
	userKeys map[uint]map[string]struct{}
}

func NewInMemoryHistoryCache(ttl time.Duration) HistoryCache {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return &inMemoryHistoryCache{
		ttl:      ttl,
		store:    cache.GoCacheMem(),
		userKeys: make(map[uint]map[string]struct{}),
	}
}

func (c *inMemoryHistoryCache) key(userID uint, page, limit int) string {
	return strconv.FormatUint(uint64(userID), 10) + ":" + strconv.Itoa(page) + ":" + strconv.Itoa(limit)
}

func (c *inMemoryHistoryCache) Get(userID uint, page, limit int) (*dto.TransactionHistoryResponse, bool) {
	k := c.key(userID, page, limit)
	val, ok := c.store.Get(k)
	if !ok {
		return nil, false
	}
	entry, typeOK := val.(historyCacheEntry)
	if !typeOK {
		c.store.Delete(k)
		return nil, false
	}
	return cloneHistoryResponse(entry.value), true
}

func (c *inMemoryHistoryCache) Set(userID uint, page, limit int, resp *dto.TransactionHistoryResponse) {
	k := c.key(userID, page, limit)
	c.store.Set(k, historyCacheEntry{
		value:     cloneHistoryResponse(resp),
		expiresAt: time.Now().Add(c.ttl),
	}, c.ttl)

	c.mu.Lock()
	if _, ok := c.userKeys[userID]; !ok {
		c.userKeys[userID] = make(map[string]struct{})
	}
	c.userKeys[userID][k] = struct{}{}
	c.mu.Unlock()
}

func (c *inMemoryHistoryCache) InvalidateUser(userID uint) {
	c.mu.Lock()
	keys := c.userKeys[userID]
	delete(c.userKeys, userID)
	c.mu.Unlock()

	for k := range keys {
		c.store.Delete(k)
	}
}

func cloneHistoryResponse(in *dto.TransactionHistoryResponse) *dto.TransactionHistoryResponse {
	if in == nil {
		return nil
	}
	out := &dto.TransactionHistoryResponse{
		Transactions: make([]dto.TransactionHistoryItem, len(in.Transactions)),
		Pagination:   in.Pagination,
	}
	copy(out.Transactions, in.Transactions)
	return out
}
