# Lark Notification

Docker image / CI plugin to send pipeline notifications to Lark (Feishu) using Interactive Message Cards. Designed to work with Woodpecker CI, but can be used with any CI system that supports Docker images.

## Features

- Customizable notification with:
  - Pipeline status (success/failure)
  - Author information
  - Commit/build details
  - Optional variables section
  - Customizable action buttons
- Support for both text messages and interactive cards
- Message signature verification
- Debug mode

## Configuration

### Basic Configuration

```yaml
steps:
  - name: notify-lark
    image: 7a6163/ci-lark-notification
    settings:
      webhook_url: https://open.feishu.cn/open-apis/bot/v2/hook/...
    when:
      - status: [success, failure]
        event: [manual, push, tag]
```

### Environment Variables

The plugin uses environment variables:

- `CI_REPO` - Repository name
- `CI_REPO_URL` - Repository URL
- ~~`CI_PIPELINE_STATUS`~~ `CI_PREV_PIPELINE_STATUS` - Pipeline status
  * `CI_PIPELINE_STATUS` is missing in 3.1.0 :(
- `CI_PIPELINE_URL` - Pipeline URL
- `CI_PIPELINE_FORGE_URL` - Forge commit URL
- `CI_COMMIT_SHA` - Commit SHA (shortened to 7 characters)
- `CI_COMMIT_TAG` - Release tag (if available)
- `CI_COMMIT_MESSAGE` - Commit message
- `CI_COMMIT_AUTHOR` - Commit author
- `CI_COMMIT_AUTHOR_AVATAR` - Author's avatar URL



### Plugin Settings

- `webhook_url` (required) - Lark webhook URL
- `secret` (optional) - Secret for signature verification
- `use_card` (optional) - Use interactive card instead of text message (default: true)
- `debug` (optional) - Enable debug output of the message JSON
- `buttons` (optional) - Comma-separated list of buttons to display:
  - `pipeline` - Link to pipeline
  - `commit` - Link to commit (for non-tag builds)
  - `release` - Link to release (for tag builds)
  - Default: all buttons are shown
- `variables` (optional) - Comma-separated list of environment variables to display

### Example Configuration

```yaml
steps:
  - name: notify-lark
    image: 7a6163/ci-lark-notification
    settings:
      webhook_url:
        from_secret: lark_webhook_url
      secret:
        from_secret: lark_secret
      use_card: true
      buttons: pipeline,commit
      variables: MY_VAR1,MY_VAR2
      debug: true
    when:
      - status: [success, failure]
        event: [manual, push, tag]
```

## Development

The plugin is written in Go and uses [Lark Interactive Message Cards](https://open.feishu.cn/document/ukTMukTMukTM/uYTNwUjL2UDM14iN1ATN) for rich notifications. It supports customization through environment variables and plugin settings.

### Lark Webhook Setup

1. Create a custom bot in your Lark workspace
2. Configure the webhook URL and copy it
3. Optionally, enable signature verification and set a secret

For enhanced security, it's recommended to use the signature verification feature by providing a secret.

## Text Message vs Interactive Card

The plugin supports two message formats:

1. **Text Message** - Simple text-based notification with emoji and formatting
2. **Interactive Card** - Rich card with colored header, formatted text, and action buttons

You can choose the format using the `use_card` setting.

Inspired by [woodpecker-teams-notify-plugin](https://github.com/GECO-IT/woodpecker-plugin-teams-notify) and [ci-teams-notification](https://github.com/mobydeck/ci-teams-notification).
