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

The application is deployed to production using GitHub Actions and the cloudlab deployment workflow.

### Prerequisites

- Docker Swarm initialized on the cloudlab server
- GitHub Actions secrets configured
- Bot images published to GitHub Container Registry (ghcr.io)

### Required GitHub Secrets

Set up the following secrets in your GitHub repository (Settings → Secrets and variables → Actions):

**SSH Connection**:
- `DO_SSH_PRIVATE_KEY` - SSH private key for accessing the cloudlab server
- `DO_HOST` - Cloudlab server IP address (e.g., `167.172.190.133`)
- `DO_USER` - SSH username for deployment (e.g., `deployer`)

**Application Secrets**:
- `TELEGRAM_TOKEN` - Telegram bot token from [@BotFather](https://t.me/botfather)
- `ANTHROPIC_API_KEY` - Anthropic API key from [Anthropic Console](https://console.anthropic.com/)
- `CODECOV_TOKEN` - Codecov token for coverage reports

**Environment Variable Mapping**:
- `TELEGRAM_TOKEN` (GitHub secret) → `BOT_TELEGRAM_TOKEN` (container)
- `ANTHROPIC_API_KEY` (GitHub secret) → `AI_API_KEY` (container)

### Deployment Process

The deployment is fully automated via GitHub Actions:

1. **On Pull Request**: Build and test the Docker image (amd64 only)
2. **On Tag Push** (e.g., `v1.0.0`):
   - Build multi-arch Docker images (amd64, arm64)
   - Push images to GitHub Container Registry
   - Deploy to production using the cloudlab workflow
   - Automatic rollback on failure

**To deploy a new version**:
```bash
git tag v1.0.0
git push origin v1.0.0
```

**Manual deployment** (if needed):
```bash
# SSH to the cloudlab server
ssh -p 1923 deployer@167.172.190.133

# Navigate to deployment directory
cd ~/cloudlab/stacks/helpmypet

# Deploy the stack
docker stack deploy -c docker-compose.yml helpmypet
```

### Monitoring

**Check service status**:
```bash
ssh -p 1923 deployer@167.172.190.133
docker stack ps helpmypet
```

**View service logs**:
```bash
# Bot logs
docker service logs -f helpmypet_help-my-pet

# Redis logs
docker service logs -f helpmypet_redis
```

**Health check**:
```bash
# Check if services are running
docker service ls | grep helpmypet
```

### Rollback

If deployment fails, the system automatically rolls back to the previous version. For manual rollback:

```bash
docker service update --rollback helpmypet_help-my-pet
```

## Project Structure

- `cmd/help-my-pet/` - Main application entry point
- `pkg/`
  - `bot/` - Telegram bot service implementation
  - `core/` - Core business logic and AI service
  - `cmd/` - Command line interface and configuration

## License

This project is licensed under the MIT License - see the LICENSE file for details.
