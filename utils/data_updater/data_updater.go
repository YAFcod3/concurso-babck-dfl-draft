package data_updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

type ExchangeRate struct {
	Base      string             `json:"base"`
	Timestamp int64              `json:"timestamp"`
	Rates     map[string]float64 `json:"rates"`
}

func StartExchangeRateUpdater(client *redis.Client, interval time.Duration) {
	ctx := context.Background()
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	fetchExchangeRates := func() {
		resp, err := httpClient.Get("https://concurso.dofleini.com/exchange-rate/api/latest?base=USD")
		if err != nil {
			fmt.Println("Error fetching exchange rates:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("Error: received status code %d\nResponse body: %s\n", resp.StatusCode, string(body))
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}

		var exchangeRate ExchangeRate
		err = json.Unmarshal(body, &exchangeRate)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return
		}

		key := "exchange_rates"
		pipe := client.TxPipeline() // Iniciar pipeline

		// Almacenar tasas de cambio en el pipeline
		for currency, rate := range exchangeRate.Rates {
			pipe.HSet(ctx, key, currency, rate)
		}

		// Almacenar base y timestamp en el pipeline
		pipe.HSet(ctx, key, "base", exchangeRate.Base)
		pipe.HSet(ctx, key, "timestamp", exchangeRate.Timestamp)

		// Ejecutar pipeline
		_, err = pipe.Exec(ctx)
		if err != nil {
			fmt.Println("Error storing exchange rates in Redis:", err)
			return
		}

		fmt.Println("Exchange rates updated successfully!")
	}

	fetchExchangeRates() // Primera actualización

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			fetchExchangeRates()
		}
	}()
}
