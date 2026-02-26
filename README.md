# Igris Inertial Go SDK

Go client for [Igris Inertial](https://igris-inertial.com) -- the AI inference gateway with multi-provider routing, SLO enforcement, fleet management, and BYOK vault.

## Installation

```bash
go get github.com/igris-inertial/go-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    igris "github.com/igris-inertial/go-sdk"
)

func main() {
    client := igris.NewClient(
        "https://api.igris-inertial.com",
        "your-api-key",
    )

    resp, err := client.Infer(context.Background(), &igris.InferRequest{
        Model: "gpt-4",
        Messages: []igris.Message{
            {Role: "user", Content: "Hello!"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(resp.Choices[0].Message.Content)
}
```

## Features

- **Zero dependencies** -- Uses only the standard library
- **Context support** on all methods
- **Functional options** for client configuration
- **Multi-provider inference** -- Route across OpenAI, Anthropic, Google, and more
- **Provider management** -- Register, test, and monitor providers
- **BYOK vault** -- Securely store and rotate your own API keys
- **Fleet management** -- Register and monitor inference agents
- **Usage tracking** -- Monitor costs and token usage

## API

```go
client := igris.NewClient(baseURL, apiKey, igris.WithTenantID("t1"))

// Inference
client.Infer(ctx, &igris.InferRequest{...})
client.ChatCompletion(ctx, &igris.InferRequest{...})
client.ListModels(ctx)
client.Health(ctx)

// Providers
client.Providers.Register(ctx, &igris.ProviderConfig{...})
client.Providers.List(ctx)
client.Providers.Test(ctx, &igris.ProviderConfig{...})
client.Providers.Update(ctx, id, config)
client.Providers.Delete(ctx, id)
client.Providers.Health(ctx, id)

// Vault
client.Vault.Store(ctx, &igris.VaultStoreRequest{...})
client.Vault.List(ctx)
client.Vault.Rotate(ctx, "openai")
client.Vault.Delete(ctx, "openai")

// Fleet
client.Fleet.Register(ctx, config)
client.Fleet.Agents(ctx)
client.Fleet.Health(ctx)
client.Fleet.Telemetry(ctx, fleetID, data)

// Usage & Audit
client.Usage.Current(ctx)
client.Usage.History(ctx, params)
client.Audit.List(ctx, params)
```

## Execution Receipt Verification (v2.2.0+)

When Overture is backed by a Runtime instance, inference responses include an
`ExecutionReceipt` with signed resource-accounting data. Verification is
opt-in and requires the Runtime's public key.

```go
import (
    "crypto/ed25519"
    "encoding/hex"

    igris "github.com/igris-inertial/go-sdk"
)

resp, err := client.Infer(ctx, req)
// ...

if resp.ExecutionReceipt != nil {
    pubKeyHex := "..." // matches IGRIS_RUNTIME_PUBLIC_KEY
    pubKeyBytes, _ := hex.DecodeString(pubKeyHex)
    pub := ed25519.PublicKey(pubKeyBytes)

    if err := igris.VerifyReceipt(resp.ExecutionReceipt, pub); err != nil {
        log.Printf("receipt verification failed: %v", err)
    }

    fmt.Printf("cpu=%dms wall=%dms violation=%v\n",
        resp.ExecutionReceipt.CpuTimeMs,
        resp.ExecutionReceipt.WallTimeMs,
        resp.ExecutionReceipt.ViolationOccurred,
    )
}
```

`VerifyReceipt` is never called automatically. Responses from servers that do
not emit receipts decode normally — `ExecutionReceipt` will be `nil`.

## Changelog

### v2.2.0
- Added `ExecutionReceipt` field to `InferResponse`
- Added `VerifyReceipt(receipt *ExecutionReceipt, pub ed25519.PublicKey) error`

## License

MIT
