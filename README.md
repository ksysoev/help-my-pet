# Help My Pet

An AI-powered Telegram bot that helps pet owners by answering questions about pet health using the Anthropic Claude AI model.

## Features

- Answers pet health-related questions using Anthropic's Claude AI
- Provides helpful and compassionate advice
- Always recommends consulting with a veterinarian for serious medical conditions
- Easy to use through Telegram messaging

## Prerequisites

- Go 1.21 or later
- Telegram Bot Token (obtain from [@BotFather](https://t.me/BotFather))
- Anthropic API Key (obtain from [Anthropic Console](https://console.anthropic.com/))

## Installation

1. Clone the repository:
```bash
git clone https://github.com/ksysoev/help-my-pet.git
cd help-my-pet
```

2. Install dependencies:
```bash
go mod download
```

3. Create a configuration file:
```bash
cp config.yaml config.local.yaml
```

4. Edit `config.local.yaml` and add your API keys:
```yaml
bot:
  telegram_token: "your-telegram-bot-token"

ai:
  anthropic_key: "your-anthropic-api-key"
  model: "claude-2"
```

## Usage

1. Start the bot:
```bash
go run cmd/help-my-pet/main.go bot --config config.local.yaml
```

2. Open Telegram and start chatting with your bot
3. Ask any pet health-related questions

## Development

- Run tests:
```bash
go test ./...
```

- Run linter:
```bash
golangci-lint run
```

## Project Structure

- `cmd/help-my-pet/` - Main application entry point
- `pkg/`
  - `bot/` - Telegram bot service implementation
  - `core/` - Core business logic and AI service
  - `cmd/` - Command line interface and configuration

## License

This project is licensed under the MIT License - see the LICENSE file for details.
