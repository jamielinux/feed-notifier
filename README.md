# feed-notifier

Check RSS/Atom feeds for new articles and send notifications.

> [!NOTE]
> This was a quick weekend project to learn Golang! :grin:

## Features

- ⚡ Concurrent fetches.
- 🔔 Multiple notification methods:
    - Mattermost incoming webhook (with HTML to markdown conversion if needed)
    - Pushover API
    - More coming soon ...
- 🤝 Respectful when fetching:
    - Uses `max-age`, `etag` and `last-modified` if available.

### Coming soon

- 🚧 Binary release.
- 🚧 Support for include/exclude filters on articles.
- 🚧 Support for templating notifications.
- 🚧 Optional notifications when an article is updated.
- 🚧 Ability to send test notifications.
- 🚧 More notification methods.

## Usage

> [!CAUTION]
> This is pre-alpha software! It's possible even the database (sqlite3) may be swapped for something else.

Copy `config_example.yml` to anywhere:


```console
$ cp config_example.yml "$HOME/.config/feed-notifier/config.yml"
```

Modify the config:

```console
$ vim "$HOME/.config/feed-notifier/config.yml"
```

Run the program:

```console
$ feed-notifier "$HOME/.config/feed-notifier/config.yml"
```

### Example config

```yaml
# Path to database file. The default is: $HOME/.local/share/feed-notifier/db
database: "$HOME/.local/share/feed-notifier/db"

# Enable debug logging.
debug: false

# Global settings for fetching RSS/Atom feeds.
fetch:
  # The number of concurrent fetches (default=3). If 0, the default is used.
  # The maximum is 10.
  jobs: 3
  # The interval (in minutes) to wait before refreshing feeds (default=60).
  # This can be overridden in each feed. If 0, the default is used.
  interval: 60

# Define notification methods here.
#   `id` must be a unique string.
#   `type` must be one of: mattermost_webhook, pushover
notifiers:

  # The mattermost_webhook notifier must have `settings.webhook` defined.
  # Optionally, set `html_to_markdown: true` to convert the content of each
  # article from HTML to Markdown.
  - id: my-mattermost
    type: mattermost_webhook
    settings:
      webhook: "https://mattermost.example.com/hooks/bfwdg8tpyfdg..."
      html_to_markdown: true

  # The pushover notifier must have `settings.app_token` and `settings.user_key`
  # defined.
  - id: my-pushover
    type: pushover
    settings:
      app_token: "bjn4eqxb55xm..."
      user_key: "ayynt9ch8g5e..."

# Define the notifier to use by default when a feed doesn't explicitly specify
# a notifier. The `stdout` notifier is a built-in notifier that is always
# available and just prints JSON to standard output.
default_notifier: my-mattermost

# Define the feeds to fetch and the notifier to use.
# REQUIRED FIELDS
#   - `id` must be unique across all configured feeds. If you change it then
#     its article history will be reset.
#   - `url` is the URL to fetch the feed.
#   - `display_name` is the title of the feed to show in notifications.
# OPTIONAL FIELDS
#   - `interval` is the time (in minutes) between checks for new articles.
#     If not defined then the global `fetch.interval` setting is used.
#   - `notifier` is the notifier to use to send notifications for this feed.
#     If not defined then the `default_notifier` setting is used.
feeds:

  - id: hetzner
    url: "https://status.hetzner.com/en.atom"
    display_name: "Hetzner Status"
    interval: 10
    notifier: my-pushover

  - id: scaleway
    url: "https://status.scaleway.com/history.atom"
    display_name: "Scaleway Status"
```

## License

`feed-notifier` is distributed under the terms of the [Mozilla Public License 2.0](LICENSE).

```
Copyright (C) 2025 Jamie Nguyen <j@jamielinux.com>

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
```
