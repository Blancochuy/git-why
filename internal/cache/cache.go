package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Cache struct {
	data     map[string]cacheEntry
	mu       sync.RWMutex
	cacheDir string
}

type cacheEntry struct {
	Data      []byte    `json:"data"`
	ExpiresAt time.Time `json:"expires_at"`
}

func New() *Cache {
	cacheDir := os.Getenv("GITWHY_CACHE_DIR")
	if cacheDir == "" {
		homeDir, _ := os.UserHomeDir()
		cacheDir = filepath.Join(homeDir, ".cache", "git-why")
	}

	c := &Cache{
		data:     make(map[string]cacheEntry),
		cacheDir: cacheDir,
	}

	os.MkdirAll(cacheDir, 0755)

	c.loadFromDisk()

	go c.cleanupLoop()

	return c
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.data[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Data, true
}

func (c *Cache) Set(key string, data []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}

	go c.saveToDisk(key, c.data[key])
}

func (c *Cache) loadFromDisk() {
	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(c.cacheDir, file.Name()))
		if err != nil {
			continue
		}

		var entry cacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		if time.Now().Before(entry.ExpiresAt) {
			c.data[file.Name()] = entry
		}
	}
}

func (c *Cache) saveToDisk(key string, entry cacheEntry) {
	entryData, err := json.Marshal(entry)
	if err != nil {
		return
	}

	safeKey := sanitizeFilename(key)
	os.WriteFile(filepath.Join(c.cacheDir, safeKey), entryData, 0644)
}

func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.ExpiresAt) {
			delete(c.data, key)
			safeKey := sanitizeFilename(key)
			os.Remove(filepath.Join(c.cacheDir, safeKey))
		}
	}
}

func sanitizeFilename(key string) string {
	result := make([]byte, 0, len(key))
	for _, b := range []byte(key) {
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '-' || b == '_' || b == ':' {
			result = append(result, b)
		} else {
			result = append(result, '_')
		}
	}
	return string(result)
}
