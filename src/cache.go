package main

import "sync"

// Cache представляет кэш для хранения данных заказов в памяти.
type Cache struct {
	data   map[string][]byte
	mutex  sync.RWMutex
}

// NewCache создает новый экземпляр Cache.
func NewCache() *Cache {
	return &Cache{
		data: make(map[string][]byte),
	}
}

// GetOrder возвращает данные заказа с указанным идентификатором из кэша.
func (c *Cache) GetOrder(uid string) ([]byte, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	order, found := c.data[uid]
	return order, found
}

// SetOrder сохраняет данные заказа в кэше.
func (c *Cache) SetOrder(order_uid string, order []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[order_uid] = order
}