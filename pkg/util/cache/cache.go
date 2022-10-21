package cache

import (
	"bytes"
	"encoding/gob"
	"runtime/debug"

	"github.com/coocood/freecache"
)

type Cache struct {
	store   *freecache.Cache
	encoder *gob.Encoder
	decoder *gob.Decoder
	buffer  *bytes.Buffer
}

// Default time on cache in secs
const defaultExpireTime = 60

var cache Cache

func Initilize() {
	cache = Cache{
		store:   freecache.NewCache(0),
		buffer:  new(bytes.Buffer),
		encoder: gob.NewEncoder(cache.buffer),
		decoder: gob.NewDecoder(cache.buffer),
	}
	debug.SetGCPercent(20)
}
func Add[T any](key string, value T) (err error) {
	cKey := []byte(key)
	err = cache.encoder.Encode(&value)
	if err != nil {
		return
	}
	err = cache.store.Set(cKey, cache.buffer.Bytes(), defaultExpireTime)
	if err != nil {
		return
	}
	cache.buffer.Reset()
	return
}

func Get[T any](key string, holder *T) (err error) {
	cKey := []byte(key)
	value, err := cache.store.Get(cKey)
	if err != nil {
		return
	}
	_, err = cache.buffer.Write(value)
	if err != nil {
		return
	}
	err = cache.decoder.Decode(holder)
	if err != nil {
		return
	}
	cache.buffer.Reset()
	return
}
