package main

import (
	//データをある一定の規則に従って、別の形式のデータに変換する
	"encoding/json"
	"strconv"
	"strings"

	//主に文字列の入出力と、フォーマット(書式設定)に関する機能を提供するパッケージです。
	"fmt"
	"io/ioutil"
	"log"

	//HTTPクライアントとサーバーの実装を提供していて、GET、POSTリクエスト、Formデータの送信など色々できます。
	"net/http"
)

const (
	// ACCESS_TOKEN use to access Telegram API
	ACCESS_TOKEN = "5699310931:AAHYRrVzTgL-9fbTeDfllNn_Hl1QikvwPhA"
	// TELEGRAM_API_URL is a base URL for Telegram API
	TELEGRAM_API_URL = "https://api.telegram.org"
)

// Message data structure for telegram message
type Message struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID           int    `json:"id"`
			IsBot        bool   `json:"is_bot"`
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			Username     string `json:"username"`
			LanguageCode string `json:"language_code"`
		} `json:"from"`
		Chat struct {
			ID        int    `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Username  string `json:"username"`
			Type      string `json:"type"`
		} `json:"chat"`
		Date     int    `json:"date"`
		Text     string `json:"text"`
		Entities []struct {
			Offset int    `json:"offset"`
			Length int    `json:"length"`
			Type   string `json:"type"`
		} `json:"entities"`
	} `json:"message"`
}

// webhook is a handler for Webhook server
//webhookメソッドでリクエストはWとｒ
func webhook(w http.ResponseWriter, r *http.Request) {
	// return all with status code 200
	//
	w.WriteHeader(http.StatusOK)

	// read body in the request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to read body: %v", err)
	}

	// initiate Message data structure to message variable
	// unmarshal []byte data into message
	var message Message
	//json.Marshalは構造体をjsonに変換
	if err := json.Unmarshal(body, &message); err != nil {
		log.Printf("failed to unmarshal body: %v", err)
		return
	}

	// 1000 jpy usd
	fmt.Println(message.Message.Text)

	// https://www.tutorialspoint.com/strings-split-function-in-golang
	//https://cdn.jsdelivr.net/gh/fawazahmed0/currency-api@1/latest/currencies/jpy/usd.json
	// need split -> [1000, jpy, usd]
	split := strings.Split(message.Message.Text, " ")

	resp, err := http.Get("https://cdn.jsdelivr.net/gh/fawazahmed0/currency-api@1/latest/currencies/" + split[1] + "/" + split[2] + ".json")

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	// initiate Message data structure to message variable
	// unmarshal []byte data into message
	var rate map[string]interface{}
	//json.Marshalは構造体をjsonに変換
	if err := json.Unmarshal(body, &rate); err != nil {
		log.Printf("failed to unmarshal body: %v", err)
		return
	}

	// rate["idr"].(float64) * 200

	// send message to end-user
	//メッセージの送信にはSendMessage関数を使用する

	number := split[0]
	//文字列をfloat型に変換する
	numberFloat, _ := strconv.ParseFloat(number, 64)

	err = sendMessage(message.Message.Chat.ID, fmt.Sprintf("%f", numberFloat*rate[split[2]].(float64)))
	if err != nil {
		log.Printf("failed to send message: %v", err)
	}

	return
}

// sendMessage sends a message to end-user
func sendMessage(chatID int, message string) error {
	// setup http request
	//stringと他の型をstring型として一緒に扱うことができるようにするfmt.Sprintfという関数
	url := fmt.Sprintf("%s/bot%s/sendMessage?chat_id=%d&text=%s", TELEGRAM_API_URL, ACCESS_TOKEN, chatID, message)
	//http.NewRequestでリクエストの準備をし、client.Doでリクエストを発行するような流れ
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed wrap request: %w", err)
	}

	// send http request
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed send request: %w", err)
	}
	//なぜクローズ？
	defer res.Body.Close()

	// print response
	log.Printf("message sent successfully?\n%#v", res)

	return nil
}

func main() {
	// create the handler
	handler := http.NewServeMux()
	handler.HandleFunc("/", webhook)

	// configure http server
	srv := &http.Server{
		Handler: handler,
		Addr:    fmt.Sprintf("localhost:%d", 3000),
	}

	// start http server
	log.Printf("http server listening at %v", srv.Addr)
	//Webサーバを構築してくれるListenAndServe
	if err := srv.ListenAndServe(); err != nil {
		//Fatalfは対象の関数のテストに失敗した記録を残し、後続のテストは実行しない。
		log.Fatalf("failed to serve: %v", err)
	}
}
