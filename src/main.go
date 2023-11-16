package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/nats-io/stan.go"
	"github.com/spf13/viper"
)

func orderFormHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("../design/html/index.html")
	if err != nil{
		fmt.Println(err)
		return
	}

	tmpl.Execute(w, "index.html")
}

func main() {
	// Чтение конфигурации из файла
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	// Подключение к nats-streaming
	nc, err := stan.Connect(
		viper.GetString("nats_cluster"),
		viper.GetString("nats_client_id"),
		stan.NatsURL(viper.GetString("nats_url")),
	)
	if err != nil {
		log.Fatalf("Failed to connect to nats-streaming: %v", err)
	}
	defer nc.Close()

	// Создание экземпляра сервиса
	service := NewOrderService(nc)

	// Запуск http-сервера
	router := mux.NewRouter()
	router.HandleFunc("/", orderFormHandler)
	router.HandleFunc("/order", service.GetOrderHandler).Methods("GET")
	

	// -t отвечает за запуск тестов
	args := os.Args
	if len(args) > 1 && args[1] == "-t"{
		go func(){
			time.Sleep(10 * time.Second)
			test()
		}()
	}	

	// Для большего функционала можем обрабатывать так:
	// switch args[0]{
	// case "-t":
	// 	funcT()
	// case "-c":
	// 	funcC()
	// case "-cm":
	// 	funcCM()
	// }

	log.Printf("Starting server on port %s...\n", viper.GetString("server_port"))
	err = http.ListenAndServe(":" + viper.GetString("server_port"), router)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}