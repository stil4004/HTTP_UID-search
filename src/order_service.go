package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/nats-io/stan.go"
	"github.com/spf13/viper"
)

// OrderService представляет сервис обработки заказов.
type OrderService struct {
	nc   stan.Conn
	db   *Database
	cache *Cache
	mutex sync.Mutex
}

// NewOrderService создает новый экземпляр OrderService.
func NewOrderService(nc stan.Conn) *OrderService {
	// Подключение к базе данных PostgreSQL
	db := NewDatabase(
		viper.GetString("db_host"),
		viper.GetInt("db_port"),
		viper.GetString("db_username"),
		viper.GetString("db_password"),
		viper.GetString("db_name"),
	)

	data, err := os.ReadFile("model.json")
	if err != nil {
    	log.Printf("Couldn't convert data from JSON %v", err)
	}
	db.writeDataToDB("b563feb7b2b84b6test", data)

	
	// Создание кэша
	cache := NewCache()

	return &OrderService{
		nc:   nc,
		db:   db,
		cache: cache,
	}
}

// GetOrderHandler обработчик запроса на получение данных заказа.
func (s *OrderService) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	// Извлечение идентификатора заказа из URL

	idStr := r.URL.Query().Get("id")
	

	// Попытка получить данные из кэша
	order, found := s.cache.GetOrder(idStr)
	if found {
		// Возвращение данных из кэша
		writeResponse(w, order)

		// Отправьте сообщение в канал с именем "cache_сhannel".
		err := s.nc.Publish("cache_channel", []byte("Downloaded from cache"))
		if err != nil {
			log.Printf("Failed to publish message: %v\n", err)
		}
		return
	}

	// Попытка получить данные из базы данных
	order, err := s.db.GetOrder(idStr)
	if err != nil {
		log.Printf("Failed to get order from database: %v", err)
		http.Error(w, "Wrong order UID", http.StatusInternalServerError)
		return
	}

	// Сохранение данных в кэш
	s.cache.SetOrder(idStr, order)

	// Возвращение данных
	writeResponse(w, order)
}

// writeResponse отправляет данные заказа в формате JSON в ответ на запрос.
func writeResponse(w http.ResponseWriter, order []byte) {
	
	// Установка заголовков для правильного вывода JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Вывод JSON на странице
	fmt.Fprint(w, string(order))

}