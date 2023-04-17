# Mattermost Calendar

![calendar_screen](https://user-images.githubusercontent.com/22306239/232596968-2b7aa5ae-c9ec-46e2-aa8c-d65cf6214c45.png)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/v/release/dmitrijkir/mattermost-plugin-calendar?include_prereleases)](https://github.com/dmitrijkir/mattermost-plugin-calendar/releases/v0.1.0-alpha)
---


## How it works

- Create appointments and events
- Organize meetings

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

Check API documentation [here](/docs/README.md)

## License

This project is licensed under the [Apache-2.0 License](LICENSE).
