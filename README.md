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
  telegram_token: "your-telegram-bot-token" # Required: Your Telegram bot token from BotFather
  rate_limit:
    user_hourly_limit: 5   # Maximum number of requests per hour per user
    user_daily_limit: 15   # Maximum number of requests per day per user
    global_daily_limit: 1000  # Maximum total requests per day across all users
    whitelist_ids: []      # List of user IDs exempt from rate limiting

ai:
  model: "claude-2"      # Optional: Anthropic model to use (default: claude-2)
  api_key: ""           # Required: Your Anthropic API key
  max_tokens: 1000      # Optional: Maximum tokens in response (default: 1000)
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

## Docker

You can run the bot using Docker in the following ways:

### Pull from GitHub Packages

1. Pull the image:
```bash
docker pull ghcr.io/ksysoev/help-my-pet:latest
```

2. Create your config file as described in the Installation section

3. Run the container:
```bash
docker run -v $(pwd)/config.local.yaml:/app/config.yaml ghcr.io/ksysoev/help-my-pet bot
```

### Build Locally

1. Build the image:
```bash
docker build -t help-my-pet .
```

2. Run the container:
```bash
docker run -v $(pwd)/config.local.yaml:/app/config.yaml help-my-pet bot
```

## Deployment

The application can be automatically deployed to Digital Ocean using GitHub Actions. The deployment process is triggered automatically when changes are pushed to the main branch.

### Prerequisites

1. A Digital Ocean droplet with Docker and Docker Compose installed
2. SSH access to the droplet
3. GitHub repository secrets configured

### Required GitHub Secrets

Set up the following secrets in your GitHub repository (Settings -> Secrets and variables -> Actions):

- `DO_SSH_PRIVATE_KEY`: SSH private key for accessing the Digital Ocean droplet
- `DO_HOST`: Your Digital Ocean droplet's IP address or hostname
- `DO_USER`: SSH user for the Digital Ocean droplet
- `ANTHROPIC_API_KEY`: Your Anthropic API key for the deployed instance

### Deployment Process

1. Push your changes to the main branch
2. GitHub Actions will automatically:
   - Build and push the Docker image to GitHub Container Registry
   - Deploy the latest version to your Digital Ocean droplet
   - Set up the configuration and start the container

### Manual Deployment

If you need to deploy manually, you can:

1. SSH into your Digital Ocean droplet
2. Pull the latest image:
   ```bash
   docker pull ghcr.io/ksysoev/help-my-pet:main
   ```
3. Update the container:
   ```bash
   docker-compose pull && docker-compose up -d
   ```

## Project Structure

- `cmd/help-my-pet/` - Main application entry point
- `pkg/`
  - `bot/` - Telegram bot service implementation
  - `core/` - Core business logic and AI service
  - `cmd/` - Command line interface and configuration

## License

This project is licensed under the MIT License - see the LICENSE file for details.
