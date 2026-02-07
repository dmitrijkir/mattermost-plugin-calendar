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
- **iCal/CalDAV Support:** Sync your calendar with external applications like Apple Calendar, Thunderbird, or Google Calendar.


## CalDAV Setup

CalDAV allows you to sync your Mattermost calendar with external calendar applications (Apple Calendar, Thunderbird, etc.) for two-way synchronization.

### Getting the CalDAV URL

1. Open the Calendar plugin in Mattermost
2. Click the **Settings** icon (gear) in the top right
3. In the settings panel, find the **iCal/CalDAV** section
4. Click **Generate Token** to create a new subscription token
5. Copy the **CalDAV URL**

### Configuring External Calendar Apps

**Important:** Authentication is done via the token in the URL. You can enter **any username and password** when prompted - they are not validated.

#### Apple Calendar (macOS/iOS)

1. Go to **System Settings** → **Internet Accounts** → **Add Account** → **Other** → **CalDAV Account**
2. Select **Manual** configuration
3. Enter:
   - **Server:** Your CalDAV URL (e.g., `https://your-mattermost.com/plugins/com.dmkir.calendar/caldav/{token}/`)
   - **Username:** Any value (e.g., `user`)
   - **Password:** Any value (e.g., `pass`)
4. Click **Sign In**

#### Thunderbird

1. Go to **Calendar** tab → Right-click → **New Calendar**
2. Select **On the Network**
3. Enter:
   - **Username:** Any value
   - **Location:** Your CalDAV URL
4. Click **Find Calendars**
5. Enter any password when prompted

#### Other CalDAV Clients

Use the CalDAV URL from the plugin settings. When prompted for credentials, enter any username and password.

### Security Notes

- The token in the URL provides full access to your calendar - keep it private
- You can revoke the token at any time from the plugin settings
- Revoking the token will immediately disconnect all synced calendar apps


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
