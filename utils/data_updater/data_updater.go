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

	// Función para obtener y almacenar tasas de cambio
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

		// Usar una única clave para almacenar las tasas
		key := "exchange_rates"

		// Almacenar en Redis
		for currency, rate := range exchangeRate.Rates {
			err := client.HSet(ctx, key, currency, rate).Err()
			if err != nil {
				fmt.Println("Error storing exchange rate in Redis:", err)
			}
		}

		// Actualizar base y timestamp
		err = client.HSet(ctx, key, "base", exchangeRate.Base, "timestamp", exchangeRate.Timestamp).Err()
		if err != nil {
			fmt.Println("Error storing exchange metadata in Redis:", err)
		}

		fmt.Println("Exchange rates updated successfully!")
	}

	// Ejecuto la primera actualización al iniciar
	fetchExchangeRates()

	// Actualización periódica
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			fetchExchangeRates()
		}
	}()
}
