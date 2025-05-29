---
name: Microsoft Teams Notification
author: mobydeck
icon: https://raw.githubusercontent.com/mobydeck/ci-teams-notification/refs/heads/main/assets/ms-teams-logo.png
description: Plugin to send pipeline notifications to Microsoft Teams using Adaptive Cards
tags: [notifications, chat]
containerImage: mobydeck/ci-teams-notification
containerImageUrl: https://hub.docker.com/r/mobydeck/ci-teams-notification
url: https://github.com/mobydeck/ci-teams-notification
---

# Teams Notification Plugin

CI plugin to send pipeline notifications to Microsoft Teams using Adaptive Cards.

For creating a Teams Webhook, follow [this guide](https://learn.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook).

## Settings

| Setting Name  | Description                                                        | Default |
|---------------|--------------------------------------------------------------------|---------|
| `webhook_url` | Teams Webhook URL (required)                                       | _none_  |
| `facts`       | Comma-separated list of facts to display (project,message,version) | _all_   |
| `buttons`     | Comma-separated list of buttons (pipeline,commit,release)          | _all_   |
| `variables`   | Comma-separated list of environment variables to display           | _none_  |
| `debug`       | Enable debug output of the card JSON                               | _false_ |

## Basic Usage

```yaml
steps:
  - name: notify-lark
    image: 7a6163/ci-lark-notification
    settings:
      webhook_url: https://outlook.office.com/webhook/...
    when:
      - status: [success, failure]
        event: [manual, push, tag]
```

## Advanced Usage

```yaml
steps:
  - name: notify-lark
    image: 7a6163/ci-lark-notification
    settings:
      webhook_url:
        from_secret: lark_webhook_url
      facts: project,version
      buttons: pipeline,commit
      variables: MY_VAR1,MY_VAR2
      debug: true
    when:
      - status: [success, failure]
        event: [manual, push, tag]
```

## Features

- Customizable notification card with:
  - Pipeline status (success/failure)
  - Author information with avatar
  - Commit/build details
  - Configurable fact sections
  - Optional variables table
  - Customizable action buttons
- Base64-encoded avatar images

## Notification Examples

### Success Notification

![success](https://raw.githubusercontent.com/mobydeck/ci-teams-notification/refs/heads/main/assets/ci-pipeline-succeeded.png)

### Failure Notification

![failure](https://raw.githubusercontent.com/mobydeck/ci-teams-notification/refs/heads/main/assets/ci-pipeline-failed.png)
