Claro, es posible separar el código para que el `main` solo gestione las APIs y los controladores, mientras que la lógica para actualizar las tasas de cambio quede en un archivo o módulo independiente. Esto ayuda a organizar el código y hacerlo más modular y mantenible.

Aquí te muestro cómo podrías estructurarlo:

1. **Crear un Archivo `data_updater.go`**: Este archivo manejará la actualización periódica de las tasas de cambio y su almacenamiento en Redis.

2. **`main.go`**: Este archivo se centrará en crear el servidor, definir las rutas de la API y delegar las tareas al módulo de actualización.

### Paso 1: Crear `data_updater.go`

Este archivo contendrá la lógica para actualizar las tasas de cambio y almacenarlas en Redis.

En este archivo:
- `StartExchangeRateUpdater` se encarga de actualizar las tasas de cambio cada intervalo de tiempo (`interval`).
- La función `fetchExchangeRates` contiene la lógica para obtener los datos de la API y guardarlos en Redis.
- Este proceso de actualización se ejecuta en una goroutine para no bloquear el resto de la aplicación.

### Paso 2: Configurar `main.go`

Ahora, en `main.go`, puedes centrarte en definir las rutas de la API y delegar la actualización de las tasas de cambio a `StartExchangeRateUpdater` en `data_updater.go`.



### Explicación:

1. **Inicialización de Redis**: Se conecta a Redis al inicio.
2. **Inicio del Proceso de Actualización**: Se llama a `StartExchangeRateUpdater` para comenzar a actualizar las tasas de cambio cada minuto.
3. **Rutas de la API con Fiber**: La ruta `/exchange-rates/:currency` permite obtener el tipo de cambio de una moneda específica desde Redis.

Ahora, tu aplicación estará organizada y el `main` estará enfocado en la configuración del servidor y la API, mientras que la lógica de actualización de datos estará en `data_updater.go`.
