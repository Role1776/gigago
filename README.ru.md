# Gigago: Go SDK для GigaChat API
[![Go Report Card](https://goreportcard.com/badge/github.com/Role1776/gigago)](https://goreportcard.com/report/github.com/Role1776/gigago) [![PkgGoDev](https://pkg.go.dev/badge/github.com/Role1776/gigago)](https://pkg.go.dev/github.com/Role1776/gigago)
<p align="left">
  <img src="https://github.com/Role1776/gigago/blob/main/logo.webp" width="300">
</p>



`gigago` — это лёгкий и идиоматичный Go SDK для GigaChat API. Он абстрагирует рутинные задачи, такие как аутентификация и повторные запросы, позволяя вам сосредоточиться на логике вашего приложения.


## Возможности

- **Автоматическое управление токенами**: Прозрачное получение и фоновое обновление OAuth-токенов.
- **Умные повторы**: Автоматический повтор запроса при ошибке авторизации (401) с обновлением токена.
- **Гибкая конфигурация**: Настройка HTTP-клиента, таймаутов, эндпоинтов и OAuth-scope через опции.
- **Полный контроль над генерацией**: Управление температурой, `top_p`, `max_tokens` и штрафами за повторения.
- **Идиоматичный API**: Простой и понятный интерфейс, следующий лучшим практикам Go.
  
**Примечание**: Потоковая передача в настоящее время не поддерживается. 

---


## Установка

```bash
go get github.com/Role1776/gigago
```

## Использование

### Быстрый старт

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Role1776/gigago"
)

func main() {
	ctx := context.Background()

	// 1. Создаем клиент с вашим авторизационным ключом.
	// Токен доступа будет получен автоматически.
	// Отключаем проверку сертификата
	client, err := gigago.NewClient(ctx, "YOUR_API_KEY", gigago.WithCustomInsecureSkipVerify(true))
	if err != nil {
		log.Fatalf("Ошибка создания клиента: %v", err)
	}
	defer client.Close() // Важно закрыть клиент для остановки фонового обновления токена.

	// 2. Получаем модель, с которой будем работать.
	model := client.GenerativeModel("GigaChat")

	// 3. (Опционально) Настраиваем параметры модели.
	model.SystemInstruction = "Ты — опытный гид по путешествиям. Отвечай кратко и по делу."
	model.Temperature = 0.7

	// 4. Формируем сообщение для отправки.
	messages := []gigago.Message{
		{Role: gigago.RoleUser, Content: "Какая столица у Франции?"},
	}

	// 5. Отправляем запрос и получаем ответ.
	resp, err := model.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("Ошибка генерации ответа: %v", err)
	}

	// 6. Печатаем ответ модели.
	
    fmt.Println(resp.Choices[0].Message.Content)
	
}
```

### Настройка клиента (Options)

При создании клиента можно передать одну или несколько опций для тонкой настройки его поведения.

```go

client, err := gigago.NewClient(
    ctx,
    "YOUR_API_KEY",
    gigago.WithCustomScope("GIGACHAT_API_CORP"), // Указание другого scope
)
// ...
```

**Доступные опции:**

- `WithCustomURLAI(url string)`: Задать URL для API генерации.
- `WithCustomURLOauth(url string)`: Задать URL для OAuth-сервиса.
- `WithCustomClient(client *http.Client)`: Использовать собственный `*http.Client`.
- `WithCustomTimeout(timeout time.Duration)`: Установить таймаут для HTTP-запросов.
- `WithCustomScope(scope string)`: Указать `scope` для получения токена (`GIGACHAT_API_PERS` или `GIGACHAT_API_CORP`). По дефолту стоит GIGACHAT_API_PERS. 

### Роли сообщений

Для управления диалогом используйте предопределенные константы ролей:

- `gigago.RoleUser`: Сообщение от пользователя.
- `gigago.RoleAssistant`: Ответ от модели.
- `gigago.RoleSystem`: Системная инструкция, задающая контекст и поведение модели.

---
## Управление токенами и жизненный цикл клиента

Вам не нужно беспокоиться об OAuth-токенах.`gigago` управляет ими полностью автоматически:

1.  **При создании**: Клиент запрашивает токен доступа и сохраняет его.
2.  **В фоне**: Запускается фоновый процесс, который обновляет токен за 15 минут до его истечения.
3.  **При ошибке**: Если запрос возвращает ошибку `401 Unauthorized`, клиент немедленно пытается обновить токен и повторяет запрос еще один раз.

### Закрытие клиента

Чтобы корректно остановить фоновый процесс обновления токена, всегда вызывайте `client.Close()` при завершении работы с клиентом.

```go
defer client.Close()
```

## Лицензия

Проект распространяется под лицензией MIT.
