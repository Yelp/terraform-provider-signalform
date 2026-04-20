# Integration

SignalFx supports integrations to ingest metrics from other monitoring systems, connect to Single Sign-On providers, and to report notifications for messaging and incident management. Note that your SignalForm API key must have admin permissions to use the SignalFx integration API.

## Example Usage

```terraform
resource "signalform_integration" "pagerduty_myteam" {
    provider = "signalform"
    name = "PD - My Team"
    enabled = true
    type = "PagerDuty"
    api_key = "1234567890"
}
```

## Argument Reference

* `name` - (Required) Name of the integration.
* `enabled` - (Required) Whether the integration is enabled.
* `type` - (Required) Type of the integration. See the full list at <https://developers.signalfx.com/reference#integrations-overview>.
* `api_key` - (Required for `PagerDuty`) PagerDuty API key.
* `webhook_url` - (Required for `Slack`) Slack incoming webhook URL.

**Notes**

This resource does not support all known types of integration. Contributions are welcome to implement more types.
