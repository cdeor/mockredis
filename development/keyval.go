package main

import "sync"

type KV struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewKV() *KV {
	return &KV{
		data: make(map[string][]byte),
	}
}

func (kv *KV) SET(key string, val []byte) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.data[key] = val
}

func (kv *KV) GET(key string) ([]byte, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	val, ok := kv.data[key]
	return val, ok
}

func (kv *KV) DEL(keys []string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	for _, key := range keys {
		delete(kv.data, key)
	}
}

func (kv *KV) KEYS() []string {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	keys := make([]string, 0, len(kv.data))
	for k := range kv.data {
		keys = append(keys, k)
	}
	return keys
}
