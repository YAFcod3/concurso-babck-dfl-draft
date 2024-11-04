package generate_transaction_code

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

type CodeGenerator struct {
	counter  int           // Contador en memoria
	Client   *redis.Client // Cliente de Redis
	mu       sync.Mutex    // Mutex para controlar el acceso concurrente
	lastHour string        // Última hora almacenada
}

// Cargar el contador y la última hora desde Redis
func (cg *CodeGenerator) LoadLastCounter() {
	ctx := context.Background()
	vals, err := cg.Client.HGetAll(ctx, "transaction_data").Result()
	if err == nil {
		cg.counter, _ = strconv.Atoi(vals["counter"])
		cg.lastHour = vals["last_hour"]
	}
}

// Generar el código
func (cg *CodeGenerator) GenerateCode() (string, error) {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	now := time.Now()
	currentHour := now.Format("06010215") // Formato YYMMDDHH

	// Si la hora cambió, reiniciar el contador
	if currentHour != cg.lastHour {
		cg.counter = 0            // Reiniciar el contador
		cg.lastHour = currentHour // Actualizar la última hora

		// Guardar el contador y la última hora en Redis
		ctx := context.Background()
		err := cg.Client.HSet(ctx, "transaction_data", "counter", cg.counter, "last_hour", cg.lastHour).Err()
		if err != nil {
			log.Println("Error al actualizar datos en Redis:", err)
		}
	} else if cg.counter == 100000000 {
		// Si el contador ha llegado a 100000000, detener la generación de códigos hasta la próxima hora
		return "", fmt.Errorf("límite de códigos alcanzado para la hora actual")
	}

	// Generar el código en el formato requerido
	code := fmt.Sprintf("T%s%08d", currentHour, cg.counter)
	cg.counter++ // Incrementar el contador en memoria para la próxima llamada

	// Guardar el nuevo valor del contador en Redis  // ! ver si mejorar
	ctx := context.Background()
	err := cg.Client.HSet(ctx, "transaction_data", "counter", cg.counter).Err()
	if err != nil {
		log.Println("Error al actualizar counter:", err)
	}

	return code, nil
}
