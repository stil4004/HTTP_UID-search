package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/nats-io/stan.go"
	"github.com/spf13/viper"
)


func test() {
	// Вывод полученных данных
	fmt.Println("Starting test...")

	// Обычная проверка на поиск данных и их корректность
	err := baseSearchTest()
	if err != nil{
		fmt.Printf("Test 1 wrong: %v\n", err)
		return
	}
	
	// Проверяем что повторный запрос берется из кэша
	err = cacheSearchTest()
	if err != nil{
		fmt.Printf("Test 2 wrong: %v\n", err)
		return
	}

	// Проверяем поведение сервиса на неправильный запрос
	err = wrongSearchTest()
	if err != nil{
		fmt.Printf("Test 3 wrong: %v\n", err)
		return
	}
	
	fmt.Println("All tests passed ✅")

}

func baseSearchTest() error {

	// URL сервера
	url := "http://" + viper.GetString("server_url") + ":" + viper.GetString("server_port") + "/order"

	// Параметры запроса
	id := "b563feb7b2b84b6test"

	// Формирование полного URL с параметром id
	requestURL := fmt.Sprintf("%s?id=%s", url, id)

	// Выполнение GET-запроса
	response, err := http.Get(requestURL)
	if err != nil {
		//fmt.Println("Ошибка при выполнении запроса:", err)
		return err
	}
	defer response.Body.Close()

	// Чтение тела ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	
	var test_order, right_order Order

	// Парсинг JSON-ответа
	err = json.Unmarshal(body, &test_order)
	if err != nil {
		//fmt.Println("Ошибка при парсинге JSON:", err)
		return err
	}

	// Парсинг "правильного" JSON
	data, err := os.ReadFile("model.json")
	if err != nil {
    	log.Printf("Couldn't convert data from JSON %v", err)
	}
	err = json.Unmarshal(data, &right_order)
	if err != nil {
		return err
	}

	// Проверка на корректность файлов
	if right_order.OrderUID == test_order.OrderUID{
		fmt.Println("Test 1 passed✅")
		return nil
	}
	return errors.New("wrong answer❌")
}

// Тестим, что при повторной проверке данные выдаются из кэша
func cacheSearchTest() error{
	nc, err := stan.Connect(
		viper.GetString("nats_cluster"),
		"unit-test",
		stan.NatsURL(viper.GetString("nats_url")),
	)
	if err != nil {
		return err
	}
	defer nc.Close()

	// Создаем канал и подписываемся на него
	messages := make(chan *stan.Msg)

	sub, err := nc.Subscribe("cache_channel", func(msg *stan.Msg) {
		messages <- msg
	})

	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	// Делаем повторный запрос
	url := "http://" + viper.GetString("server_url") + ":" + viper.GetString("server_port") + "/order"

	// Параметры запроса
	id := "b563feb7b2b84b6test"

	// Формирование полного URL с параметром uid
	requestURL := fmt.Sprintf("%s?id=%s", url, id)

	// Выполнение GET-запроса
	response, err := http.Get(requestURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	//

	// Ждите сообщения и проверьте его содержимое.
	msg := <-messages
	if string(msg.Data) == "Downloaded from cache" {
			fmt.Println("Test 2 passed✅")
			return nil
		} else {
			return errors.New("wrong answer❌")
	}
}

func wrongSearchTest() error{

// URL сервера
	url := "http://" + viper.GetString("server_url") + ":" + viper.GetString("server_port") + "/order"

	// Параметры запроса
	id := "123"

	// Формирование полного URL с параметром id
	requestURL := fmt.Sprintf("%s?id=%s", url, id)

	// Выполнение GET-запроса
	response, err := http.Get(requestURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Чтение тела ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// Проверка на корректность некорректности)
	if strings.Contains(string(body), "Wrong order UID"){
		fmt.Println("Test 3 passed✅")
		return nil
	}
	return errors.New("wrong answer❌")
}
