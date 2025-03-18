package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatal("Erro ao criar requisição:", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Erro ao fazer requisição ao servidor:", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Erro ao ler resposta:", err)
	}

	var data Response
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal("Erro ao decodificar JSON:", err)
	}

	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Fatal("Erro ao criar arquivo:", err)
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar: %s\n", data.Bid))
	if err != nil {
		log.Fatal("Erro ao escrever no arquivo:", err)
	}

	fmt.Println("Cotação salva em cotacao.txt")
}
