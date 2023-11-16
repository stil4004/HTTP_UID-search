package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// Database представляет базу данных PostgreSQL.
type Database struct {
	db *sql.DB
}


func createOrderTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS orders (
			order_uid TEXT PRIMARY KEY,
			data JSON
		)`
	_, err := db.Exec(query)
	return err
}

// NewDatabase создает новый экземпляр Database и устанавливает соединение с базой данных.
func NewDatabase(host string, port int, username string, password string, dbName string) *Database {
	// Формирование строки подключения
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, dbName,
	)

	// Подключение к базе данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Проверка подключения
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to the database")

	err = createOrderTable(db)
	if err != nil{
		log.Fatalf("Failed to create new database: %v", err)
	}

	return &Database{
		db: db,
	}
}

// GetOrder возвращает данные заказа с указанным идентификатором из базы данных.
func (d *Database) GetOrder(uid string) ([]byte, error) {

	query := `
		SELECT data FROM orders WHERE order_uid = $1`
	row := d.db.QueryRow(query, uid)
	
	var data []byte
	err := row.Scan(&data)
	return data, err

}

// Запись полученных данных в базу данных PostgreSQL
func (d *Database) writeDataToDB(orderUID string, data []byte) error {
	query := `
		INSERT INTO orders (order_uid, data)
		VALUES ($1, to_json($2::json))
		`
	_, err := d.db.Exec(query, orderUID, data)
	return err
}