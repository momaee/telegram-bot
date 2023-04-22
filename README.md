# Telegram Bot

This repository contains code for a Telegram Bot that can interact with users on the Telegram platform. The bot is built using Golang and connects to the Telegram API via the `telegram-bot-api` library. This bot is also connected to ChatGPT, a large language model trained by OpenAI, to enable it to respond to a wider range of user inputs.

## How does it work?

There are two deployed bots, based on the same codebase. The first one is for dev environments, and the second one is for production. The telegram id of the first bot is [`@momaeedevbot`](https://t.me/momaeedevbot), and the second one is [`@momaeebot`](https://t.me/momaeebot). The dev bot is used for testing and development and is not fully stable, and the production bot is used for actual interactions with users.

There are 4 branches for this repo:

- `dev`: the dev branch, used for development and testing, everything is pushed to this branch first.
- `release-cloud-run`: the release branch for the production bot, this branch is deployed to Google Cloud Run.
- `release-app-engine`: the release branch for the production bot, this branch is deployed to Google App Engine.
- `main`: the main branch, which dev merges into after testing and automatic deployment to Google Cloud Run.
<!-- ## Getting started

To get started with this bot, you will need to follow these steps:

1. Clone this repository to your local machine using `git clone https://github.com/momaee/telegram-bot.git`.
2. Create a new Telegram bot using the [BotFather](https://telegram.me/BotFather) bot on Telegram. Follow the instructions to create a new bot and obtain a bot token.
3. Open the `config.go` file in the project and add the bot token obtained from the BotFather in the `Token` field.
4. Install the required dependencies using `go get`.
5. Run the bot using `go run main.go`. -->

<!-- ## Features

This bot currently has the following features:

- Sends a welcome message when a user starts a chat with the bot.
- Sends a random quote from a list of predefined quotes when the user sends the command `/quote`.
- Sends a random joke from a list of predefined jokes when the user sends the command `/joke`.
- Allows users to send feedback using the command `/feedback`. The feedback is sent to the bot owner via email. -->

## Contributions

Contributions to this project are welcome. If you find any bugs or would like to add a new feature, please feel free to submit a pull request.
