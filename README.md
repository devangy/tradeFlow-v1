# TradeFlow : Prediction Markets Whale Tracker 🐋

TradeFlow is a Telegram bot built in Golang with Telego library. It monitors real-time activities across major prediction markets like latest events and tries to track whale wallets or smart traders with good winrate % from past 50 closed trades. This works by long polling multiple APIs from **Polymarket** and **Kalshi** using Golang goroutines and channels for aggregation. When a trader with good winrate is identified the bot sends a message to users instantly via Telegram.


### What are Prediction Markets?

Prediction markets are platforms where people trade on the outcomes of future events such as elections, economic indicators, sports, or global events.

Market prices reflect the collective belief about the likelihood of an outcome.

**Market Price ≈ Probability**

Example: 

Event: Will xyz polictical party win the elections in 2026?
- YES @ $0.62 → 62% chance
- NO @ $0.38 → 38% chance


Links: 
- [Wikipedia Prediction Markets?](https://en.wikipedia.org/wiki/Prediction_market)
- [Polymarket](https://polymarket.com) 
- [Kalshi Prediction Market](https://kalshi.com)

##  Features

-   **Multi-Market Intelligence**: Aggregate data from 4 distinct APIs covering Kalshi and Polymarket.
-   **Whale Tracking**: Automated filtering and analysis of past 50 trades to identify "smart traders" movement.
-   **Polling Engine**: Efficient concurrent polling built on Go's goroutines for low-latency updates every 300  ms.
-   **Telego**: Uses the high-performance `telego` library for seamless Telegram Bot API interaction.
-   **Infrastructure as Code**: Includes Terraform configurations for cloud deployment on AWS EC2 Instance.
- **Proxy Server Support:**  Rotating proxy server support for restricted regions for polling API's.  


## Tech Stack

-   **Language**: Golang
-   **Telegram Library**: [Telego](https://github.com/mymmrac/telego)
-   **Containerization**: Docker & Docker Compose
-   **Infrastructure**: Terraform (AWS Provider)
-   **Data Sources**: [Kalshi API](https://docs.kalshi.com/) , [Polymarket API](https://docs.polymarket.com/)



## Prerequisites

-   [Go 1.21+](https://go.dev/dl/) (if running locally)
-   [Docker](https://www.docker.com/get-started), [Docker Compose](https://docs.docker.com/compose/)
-   Telegram Bot Token (from [@BotFather](https://t.me/botfather))
-   API Credentials for Kalshi/Polymarket
- Proxy server credentials IF REQUIRED

## Installation

### 1. Clone the repository
```bash
git clone git@github.com:devangy/tradeFlow-v1.git

cd tradeFlow-v1
```


### 2. Create a .env file in the root directory:
```env
# proxy server details only if restrictions in country
username=
password=
country=
entryPoint=

# api endpoints
kalshi_events_API=
poly_events_API=
poly_trades_API=
poly_walletProfile_API=
kalshi_trades_API=

#tg bot token
BOT_TOKEN=
```
