# Gigago: Go SDK for the GigaChat API

> **[Читать документацию на русском языке](https://github.com/Role1776/gigago/blob/main/README.ru.md)**

[![Go Report Card](https://goreportcard.com/badge/github.com/Role1776/gigago)](https://goreportcard.com/report/github.com/Role1776/gigago) [![PkgGoDev](https://pkg.go.dev/badge/github.com/Role1776/gigago)](https://pkg.go.dev/github.com/Role1776/gigago)

<p align="left">
  <img src="https://github.com/Role1776/gigago/blob/main/logo.webp" width="300">
</p>

`gigago` is a lightweight and idiomatic Go SDK for the GigaChat API. It abstracts away routine tasks like authentication and request retries, allowing you to focus on your application's logic.

## Features

- **Automatic Token Management**: Seamlessly obtains and refreshes OAuth tokens in the background.
- **Smart Retries**: Automatically retries requests on authorization failures (401) after refreshing the token.
- **Flexible Configuration**: Customize the HTTP client, timeouts, API endpoints, and OAuth scope via options.
- **Full Generation Control**: Manage temperature, `top_p`, `max_tokens`, and repetition penalties.
- **Idiomatic API**: A simple and clean interface that follows Go best practices.

**Note**: Streaming is not currently supported.

---

## Installation

```bash
go get github.com/Role1776/gigago
```

## Usage

### Quick Start

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

	// 1. Create a client with your authorization key.
	// An access token will be fetched automatically.
	// Disabling certificate verification
	client, err := gigago.NewClient(ctx, "YOUR_API_KEY", gigago.WithCustomInsecureSkipVerify(true))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close() // Important: close the client to stop the background token refresher.

	// 2. Get the model you want to work with.
	model := client.GenerativeModel("GigaChat")

	// 3. (Optional) Configure the model's parameters.
	model.SystemInstruction = "You are an expert travel guide. Be concise and to the point."
	model.Temperature = 0.7

	// 4. Prepare your message.
	messages := []gigago.Message{
		{Role: gigago.RoleUser, Content: "What is the capital of France?"},
	}

	// 5. Send the request and get the response.
	resp, err := model.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("Failed to generate response: %v", err)
	}

	// 6. Print the model's response.
	if len(resp.Choices) > 0 {
		fmt.Println(resp.Choices[0].Message.Content)
	}
}
```

### Client Configuration (Options)

You can pass one or more options when creating a client to fine-tune its behavior.

```go
// Example: creating a client with a different OAuth scope.
client, err := gigago.NewClient(
    ctx,
    "YOUR_API_KEY",
    gigago.WithCustomScope("GIGACHAT_API_CORP"), // Specify a different scope
)
// ...
```

**Available Options:**

- `WithCustomURLAI(url string)`: Sets a custom URL for the AI API endpoint.
- `WithCustomURLOauth(url string)`: Sets a custom URL for the OAuth service.
- `WithCustomClient(client *http.Client)`: Uses a custom `*http.Client`.
- `WithCustomTimeout(timeout time.Duration)`: Sets a custom timeout for HTTP requests.
- `WithCustomScope(scope string)`: Specifies the OAuth scope (`GIGACHAT_API_B2B`, `GIGACHAT_API_PERS`, `GIGACHAT_API_CORP`). Defaults to `GIGACHAT_API_PERS`.
- `gigago.WithCustomInsecureSkipVerify(insecureSkipVerify bool)`: Disables certificate verification.

### Message Roles

Use the predefined role constants to manage the conversation flow:

- `gigago.RoleUser`: A message from the end-user.
- `gigago.RoleAssistant`: A response from the model.
- `gigago.RoleSystem`: A system instruction that sets the context and behavior for the model.

---

## Token Management and Client Lifecycle

You don't need to worry about OAuth tokens. `gigago` handles them automatically:

1.  **On Creation**: The client requests an access token and stores it.
2.  **In the Background**: A goroutine is launched to refresh the token 15 minutes before it expires.
3.  **On Error**: If a request returns a `401 Unauthorized` error, the client immediately attempts to refresh the token and retries the request once.

### Closing the Client

To properly stop the background token-refresh process, always call `client.Close()` when you are done with the client, typically using `defer`.

```go
defer client.Close()
```

## License

This project is licensed under the MIT License.
