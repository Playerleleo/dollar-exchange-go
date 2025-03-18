package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRate struct {
	Bid string `json:"bid"`
}

type APIResponse struct {
	USDBRL ExchangeRate `json:"USDBRL"`
}

func main() {
	db, err := sql.Open("sqlite3", "cotacoes.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY AUTOINCREMENT, bid TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
		if err != nil {
			http.Error(w, "Erro ao criar requisição", http.StatusInternalServerError)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Erro ao obter cotação", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Verifica o status HTTP da resposta
		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Erro na resposta da API", http.StatusBadGateway)
			return
		}

		var apiResp APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			http.Error(w, "Erro ao decodificar resposta", http.StatusInternalServerError)
			return
		}

		ctxDB, cancelDB := context.WithTimeout(context.Background(), 100*time.Millisecond) // Timeout ajustado
		defer cancelDB()

		_, err = db.ExecContext(ctxDB, "INSERT INTO cotacoes (bid) VALUES (?)", apiResp.USDBRL.Bid)
		if err != nil {
			log.Println("Erro ao salvar no banco:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"bid": apiResp.USDBRL.Bid})
	})

	fmt.Println("Servidor rodando na porta 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Erro ao iniciar o servidor:", err)
	}
}
