# Mattermost Calendar Plugin


![calendar_screen](https://github.com/dmitrijkir/mattermost-plugin-calendar/assets/22306239/6c59b90f-9798-4ab4-9de7-e0b9d90ec6bc)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/v/release/dmitrijkir/mattermost-plugin-calendar?include_prereleases)](https://github.com/dmitrijkir/mattermost-plugin-calendar/releases/v0.1.0-alpha)
---

The Mattermost Calendar Plugin is a powerful tool to help you schedule and manage team meetings and events directly within Mattermost.


## Features

- **Event Scheduling:** Easily create, schedule, and manage team meetings and events from within Mattermost.
- **Event Notifications:** Receive reminders and notifications for upcoming events to keep your team organized.
- **User-Friendly Interface:** Intuitive user interface for creating and managing events, making it easy for team members to use.
- **Customization:** Configure event settings, such as time slots, attendees, and descriptions, to suit your team's needs.


## Installation

To install the Mattermost Calendar Plugin, follow these steps:

1. Download the latest release from the [Releases](https://github.com/dmitrijkir/mattermost-plugin-calendar/releases) page.
2. Upload the plugin to your Mattermost server.
3. Enable the plugin in your Mattermost settings.

> **_Note_**
> 
> *Make sure that I set &parseTime=true for MySQL connection string.*


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


## Contribution

We welcome contributions to the Mattermost Calendar Plugin!


## Support

If you encounter any issues or have questions, please create a [GitHub Issue](https://github.com/dmitrijkir/mattermost-plugin-calendar/issues).


## License

This project is licensed under the [Apache-2.0 License](LICENSE).
