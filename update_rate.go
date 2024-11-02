// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"time"

// 	"github.com/go-redis/redis/v8"
// )

// type ExchangeRate struct {
// 	Rates map[string]float64 `json:"rates"`
// }

// func main() {
// 	ctx := context.Background()
// 	client := redis.NewClient(&redis.Options{
// 		Addr: "localhost:6379",
// 	})

// 	// Defininicion de  función para obtener los tipos de cambio
// 	fetchExchangeRates := func() {
// 		resp, err := http.Get("https://concurso.dofleini.com/exchange-rate/api/latest?base=USD")
// 		if err != nil {
// 			fmt.Println("Error fetching exchange rates:", err)
// 			return
// 		}
// 		defer resp.Body.Close()

// 		// Verifica el código de estado de la respuesta
// 		if resp.StatusCode != http.StatusOK {
// 			fmt.Printf("Error: received status code %d\n", resp.StatusCode)
// 			body, _ := ioutil.ReadAll(resp.Body)
// 			fmt.Println("Response body:", string(body))
// 			return
// 		}

// 		// Lee el cuerpo de la respuesta
// 		body, err := ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			fmt.Println("Error reading response body:", err)
// 			return
// 		}

// 		// Convierte el cuerpo a JSON
// 		var exchangeRate ExchangeRate
// 		err = json.Unmarshal(body, &exchangeRate)
// 		if err != nil {
// 			fmt.Println("Error unmarshalling JSON:", err)
// 			fmt.Println("Response body:", string(body)) // Imprime el cuerpo para depuración
// 			return
// 		}

// 		// Imprime las tasas de cambio obtenidas
// 		fmt.Println("Exchange rates:", exchangeRate.Rates)

// 		// Almacena las tasas de cambio en Redis
// 		for currency, rate := range exchangeRate.Rates {
// 			err := client.Set(ctx, currency, rate, 0).Err()
// 			if err != nil {
// 				fmt.Println("Error storing exchange rate in Redis:", err)
// 			}
// 		}

// 		fmt.Println("Exchange rates updated successfully!")
// 	}

// 	// Obtiene las tasas de cambio inmediatamente al iniciar
// 	fetchExchangeRates()
// 	var count int = 0

// 	// Crea un ticker para obtener las tasas de cambio cada 1 minuto
// 	ticker := time.NewTicker(1 * time.Minute)

// 	defer ticker.Stop()

// 	// Ejecuta la función de obtención cada vez que el ticker se activa
// 	for range ticker.C {
// 		count++
// 		fmt.Println(count)
// 		fetchExchangeRates()
// 	}
// }
