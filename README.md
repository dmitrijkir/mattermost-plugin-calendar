# Mattermost Calendar

![calendar_screen](https://user-images.githubusercontent.com/22306239/216053931-14c285d0-5bfb-4337-a085-d7bca89b67d7.png)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

---

## How it works

1. You are creating a new event or someone is adding to the event.
2. When your event starts, you will receive a message to a group channel or a direct message from the bot.

## Installation

1. Download the latest version from the [release page](https://github.com/dmitrijkir/mattermost-plugin-calendar/releases).
2. Upload the file through **System Console > Plugins > Plugin Management**, or manually upload it to the Mattermost server under plugin directory.
3. Enable the plugin.


## Development

> **_Note_**
>
> Building the plugin requires the following:
> - Golang: version >= **1.18**
> - NodeJS: version **14.x**
> - NPM: version **6.x**

Use ```make dist``` to build this plugin.

Use `make deploy` to deploy the plugin to your local server.

For more details on how to develop a plugin refer to the official [documentation](https://developers.mattermost.com/extend/plugins/).

## License

This project is licensed under the [Apache-2.0 License](LICENSE).
