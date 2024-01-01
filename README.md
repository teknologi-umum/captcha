# Captcha

CAPTCHA stands for "Completely Automated Public Turing test to tell Computers and Humans Apart". It's a security measure
that helps protect online services from spam and abuse by automated bots. Captchas typically present a challenge that's
easy for humans to solve but difficult for bots, such as recognizing distorted text or images.

This Telegram bot acts as a gatekeeper for your groups. When a new user tries to join, it presents them with a captcha
challenge. Only if they solve the challenge correctly will they be allowed to enter the group. This helps to deter spam
bots and protect your group from unwanted intrusions.

A hosted version of the bot is available at [@TeknumCaptchaBot](https://t.me/TeknumCaptchaBot). However, due to resource
constraints, it's not currently open for public invitations. If you're interested in using the bot, you'll need to
self-host it.

## Self-Hosting Guide

There are a few ways to self-host the captcha bot. Through these ways, you can run the bot on any kind of platform,
either on a dedicated virtual machine, Kubernetes, or a PaaS platform
like [Fly](https://fly.io) or [Heroku](https://www.heroku.com/).

Some configuration are needed to make the bot running. One mandatory configuration is Telegram Bot Token, in which you
can acquire from [@BotFather](https://t.me/BotFather). See Telegram's guide
to [Obtain Your Bot Token](https://core.telegram.org/bots/tutorial#obtain-your-bot-token).

The bot token can be configured on the environment variable as `BOT_TOKEN` or in a configuration file.

#### Configuration

There are 2 ways to configure the bot: through environment variables or through a configuration file.

##### File

Create a JSON or YAML file. Then run the bot with `--configuration-file=/path/to/config.json` or environment variable
of `CONFIGURATION_FILE=/path/to/config.yaml`.

Bear in mind that the only required field is `bot_token`. Everything else is optional.

```yaml
environment: production  # Assuming default value
bot_token: ""            # Required field
feature_flag:
    analytics: false       # Assuming default values
    badwords_insertion: false
    dukun: false
    under_attack: true
    reminder: false
home_group_id: 0       # Assuming default value
admin_ids: [ ]
# Optional sentry.io DSN, you can track project errors & performance there
sentry_dsn: ""
database:
    # Example value: postgres://username:password@host:port/database?sslmode=disable
    postgres_url: ""
    # Example value: mongodb://username:password@host:port/database
    mongo_url: ""
http_server:
    listening_host: ""
    listening_port: "8080"  # Assuming default value
under_attack:
    # Available options: "postgres", "memory"
    datastore_provider: "memory"  # Assuming default value
```

```json5
{
    // Assuming default value
    "environment": "production",
    // Required field
    "bot_token": "",
    "feature_flag": {
        "analytics": false,
        "badwords_insertion": false,
        "dukun": false,
        "under_attack": true,
        "reminder": false
    },
    "home_group_id": 0,
    "admin_ids": [],
    "sentry_dsn": "",
    "database": {
        "postgres_url": "",
        "mongo_url": ""
    },
    "http_server": {
        "listening_host": "",
        // Assuming default value
        "listening_port": "8080"
    },
    "under_attack": {
        // Assuming default value
        "datastore_provider": "memory"
    }
}
```

##### Environment Variables

Required:

* BOT_TOKEN: (No default value provided)

Optional:

* ENVIRONMENT: (Default: "production")
* FEATURE_FLAG_ANALYTICS: (Default: "false")
* FEATURE_FLAG_BADWORDS_INSERTION: (Default: "false")
* FEATURE_FLAG_DUKUN: (Default: "false")
* FEATURE_FLAG_UNDER_ATTACK: (Default: "true")
* FEATURE_FLAG_REMINDER: (Default: "false")
* HOME_GROUP_ID: (No default value provided)
* ADMIN_IDS: (No default value provided, comma-separated string)
* SENTRY_DSN: (No default value provided)
* POSTGRES_URL: (No default value provided)
* MONGO_URL: (No default value provided)
* HTTP_HOST: (No default value provided)
* HTTP_PORT: (Default: "8080")
* UNDER_ATTACK__DATASTORE_PROVIDER: (Default: "memory")

### Docker

```bash
docker run -e BOT_TOKEN='your telegram bot token' ghcr.io/teknologi-umum/captcha:latest
```

If you want to use version from the `master` branch, use `:edge` tag instead.

```bash
docker run -e BOT_TOKEN='your telegram bot token' ghcr.io/teknologi-umum/captcha:edge
```

### Docker Compose

This is the way that Teknologi Umum deploy the hosted version as seen on
the [infrastructure repository](https://github.com/teknologi-umum/infrastructure/blob/master/captcha/docker-compose.yml).

```yaml
services:
    application:
        # Change the tag to `:latest` for more stable releases
        image: ghcr.io/teknologi-umum/captcha:edge
        environment:
            BOT_TOKEN:
            TZ: UTC
        platform: linux/amd64
        healthcheck:
            test: curl -f http://localhost:8080 || exit 1
            interval: 15s
            timeout: 10s
            retries: 5
        deploy:
            mode: replicated
            replicas: 1
            restart_policy:
                condition: unless-stopped
                delay: 30s
                window: 120s
            resources:
                limits:
                    memory: 500MB
                    cpus: '1'
                reservations:
                    memory: 25MB
                    cpus: '0.10'
        logging:
            driver: json-file
            options:
                max-size: 20M
                max-file: 3
```

### Build from source

We don't provide precompiled binary, as it's more secure for everyone to build it from source themself.

1. Assuming you have Go already on your machine. If not, [download Go](https://go.dev/dl).
2. `git clone https://github.com/teknologi-umum/captcha.git`
3. `cd captcha`
4. For Mac and Linux: `go build -o captcha-bot -ldflags="-X main.version=$(git rev-parse HEAD) -s -w" ./cmd/captcha`,
   for Windows: `go build -o captcha-bot.exe -ldflags="-X main.version=$(git rev-parse HEAD) -s -w" ./cmd/captcha`
5. Provide `BOT_TOKEN` on environment variable. For Mac and Linux, it's `export BOT_TOKEN='your bot token'`. For
   Windows, it's `$env:BOT_TOKEN="your bot token"`.
6. Run `captcha-bot` (or `captcha-bot.exe` on Windows) binary.

## Contributing

Your contributions to this project are highly welcomed! Here's how you can help:

* **Report issues**: If you encounter any bugs or problems, please open an issue on GitHub.
* **Suggest features**: Have ideas for new features or improvements? Share them in the issue tracker or discuss them in
  the community.
* **Submit code changes**: If you're comfortable with coding, feel free to submit pull requests with your proposed
  changes.
* **Spread the word**: Help others discover this project by sharing it on social media or with your friends and
  colleagues.

Let's work together to make Telegram a safer and more enjoyable experience for everyone!

## License

```
Teknologi Umum Captcha Bot
Copyright (C) 2023 Teknologi Umum <opensource@teknologiumum.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
```

See [LICENSE](./LICENSE)
