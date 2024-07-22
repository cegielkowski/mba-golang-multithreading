package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const cep = "17128072"

type Response struct {
	Source   string
	Response string
}

func main() {
	// Criação de um contexto com cancelamento
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url1 := "https://brasilapi.com.br/api/cep/v1/" + cep
	url2 := "http://viacep.com.br/ws/" + cep + "/json/"

	resChan := make(chan Response)
	errChan := make(chan error)

	// Inicia goroutines para buscar dados das APIs
	go fetchData(ctx, url1, resChan, errChan)
	go fetchData(ctx, url2, resChan, errChan)

	// Seleciona a resposta mais rápida ou um timeout
	select {
	case <-time.After(1 * time.Second):
		log.Println("Request timeout")
		return
	case err := <-errChan:
		log.Printf("Error received: %v\n", err)
		return
	case res := <-resChan:
		fmt.Println("Response received:")
		fmt.Printf("Source: %s\n", res.Source)
		fmt.Printf("Content: %s\n", res.Response)
	}

	// Fechamento dos canais após uso
	close(resChan)
	close(errChan)
}

func fetchData(ctx context.Context, url string, resChan chan<- Response, errChan chan<- error) {
	// Criação da requisição com contexto
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		errChan <- err
		return
	}

	// Envio da requisição HTTP
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		errChan <- err
		return
	}
	defer res.Body.Close()

	// Verificação do status da resposta
	if res.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("error: received status code %d from %s", res.StatusCode, url)
		return
	}

	// Leitura da resposta
	respValue, err := io.ReadAll(res.Body)
	if err != nil {
		errChan <- err
		return
	}

	// Envio do resultado para o canal de resposta
	resChan <- Response{
		Source:   url,
		Response: string(respValue),
	}
}
