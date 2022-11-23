package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	serverUrl       string            = "http://localhost:8080/cotacao"
	fileName        string            = "cotacao.txt"
	requestsTimeout                   = time.Millisecond * 300
	moedas          map[string]string = map[string]string{
		"USD-BRL": "Dólar",
		"CAD-BRL": "Dólar Canadense",
		"EUR-BRL": "Euro",
		"GBP-BRL": "Libra Esterlina",
		"ARS-BRL": "Peso Argentino",
		"BTC-BRL": "Bitcoin",
		"LTC-BRL": "Litecoin",
		"JPY-BRL": "Iene Japonês",
		"CHF-BRL": "Franco Suíço",
		"AUD-BRL": "Dólar Australiano",
		"CNY-BRL": "Yuan Chinês",
		"ILS-BRL": "Novo Shekel Israelense",
		"ETH-BRL": "Ethereum",
		"XRP-BRL": "XRP",
	}
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), requestsTimeout)
	defer cancel()
	moeda := ""
	moedaDesc := "Dólar"

	for _, m := range os.Args[1:] {
		moeda = m
		break
	}

	url := serverUrl
	if moeda != "" {
		url += "?moeda=" + moeda
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Erro ao criar request] %v\n", err.Error())
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Erro na request] %v\n", err.Error())
		return
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Erro ao ler response body] %v\n", err.Error())
		return
	}

	if res.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "[Erro ao buscar cotação] %v\n", string(resBody))
		return
	}

	cotacaoFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Erro ao abrir arquivo %v] %v\n", fileName, err)
		return
	}
	defer cotacaoFile.Close()

	if desc, ok := moedas[moeda]; ok {
		moedaDesc = desc
	}

	cotacaoDesc := fmt.Sprintf("%v: %v\n", moedaDesc, string(resBody))
	_, err = cotacaoFile.WriteString(cotacaoDesc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Erro ao salvar cotação] %v\n", err)
		return
	}
	fmt.Println(cotacaoDesc)

}
