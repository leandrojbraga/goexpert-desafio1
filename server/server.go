package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	dbName          string = "cotacao.db"
	urlCotacao      string = "https://economia.awesomeapi.com.br/json/last/"
	requestsTimeout        = time.Millisecond * 200
	dbTimeout              = time.Millisecond * 10
)

type Cotacao struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Bid        string `json:"bid"`
	CreateDate string `json:"create_date"`
}

func main() {
	err := createDb()
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", CotacaoHandler)

	println("Server started")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	moeda := r.URL.Query().Get("moeda")
	if moeda == "" {
		moeda = "USD-BRL"
	}

	cotacao, err := getCotacao(&moeda)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	err = saveCotacao(cotacao)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%v", cotacao.Bid)))
}

func getCotacao(moeda *string) (*Cotacao, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestsTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%v%v", urlCotacao, *moeda), nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		var responseErro map[string]any
		err = json.Unmarshal(resBody, &responseErro)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf(fmt.Sprintf("%v", responseErro["message"]))
	}

	var cotacoes map[string]Cotacao
	err = json.Unmarshal(resBody, &cotacoes)
	if err != nil {
		return nil, err
	}

	var cotacao Cotacao
	for _, c := range cotacoes {
		cotacao = c
	}

	return &cotacao, nil
}

func saveCotacao(c *Cotacao) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare("insert into cotacoes (code, codein, bid, create_date) values (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, c.Code, c.Codein, c.Bid, c.CreateDate)
	if err != nil {
		return err
	}

	return nil
}

func createDb() error {
	if _, err := os.Stat(dbName); err != nil {
		println("Creating DB")
		defer println("DB Created")
		file, err := os.Create(dbName)
		if err != nil {
			return err
		}
		file.Close()

		db, err := sql.Open("sqlite3", dbName)
		if err != nil {
			return err
		}
		defer db.Close()

		stmt, err := db.Prepare("CREATE TABLE cotacoes(code TEXT NOT NULL, codein TEXT NOT NULL, bid DECIMAL(8,2) NOT NULL, create_date TEXT DATETIME NOT NULL)")
		if err != nil {
			return err
		}

		_, err = stmt.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
