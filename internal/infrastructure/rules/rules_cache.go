package rules

import (
    "os"
    "sync"
    "time"
)

type Cache struct {
    mu      sync.RWMutex
    mtime   time.Time
    content []byte
}

func (c *Cache) LoadIfChanged(path string) (changed bool, data []byte, err error) {
    fi, err := os.Stat(path)
    if err != nil { return false, nil, err }
    c.mu.RLock()
    same := !fi.ModTime().After(c.mtime)
    c.mu.RUnlock()
    if same { return false, nil, nil }
    b, err := os.ReadFile(path)
    if err != nil { return false, nil, err }
    c.mu.Lock()
    c.mtime = fi.ModTime()
    c.content = b
    c.mu.Unlock()
    return true, b, nil
}
