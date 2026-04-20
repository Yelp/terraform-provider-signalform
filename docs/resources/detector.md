# Detector

A detector is a collection of rules that define who should be notified when certain detect functions within SignalFlow fire alerts. Each rule maps a detect label to a severity and a list of notifications. When the conditions within the given detect function are fulfilled, notifications will be sent to the destinations defined in the rule for that detect function.

![Detector](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/resources/detector.png)


## Example Usage

```terraform
resource "signalform_detector" "application_delay" {
    provider = "signalform"
    count = "${length(var.clusters)}"
    name = " max average delay - ${var.clusters[count.index]}"
    description = "your application is slow - ${var.clusters[count.index]}"
    max_delay = 30
    program_text = <<-EOF
        signal = data('app.delay', filter('cluster','${var.clusters[count.index]}'), extrapolation='last_value', maxExtrapolations=5).max()
        detect(when(signal > 60, '5m')).publish('Processing old messages 5m')
        detect(when(signal > 60, '30m')).publish('Processing old messages 30m')
    EOF
    rule {
        description = "maximum > 60 for 5m"
        severity = "Warning"
        detect_label = "Processing old messages 5m"
        notifications = ["Email,foo-alerts@bar.com"]
    }
    rule {
        description = "maximum > 60 for 30m"
        severity = "Critical"
        detect_label = "Processing old messages 30m"
        notifications = ["Email,foo-alerts@bar.com"]
    }
}

provider "signalform" {}

variable "clusters" {
    default = ["clusterA", "clusterB"]
}
```

## Argument Reference

* `name` - (Required) Name of the detector.
* `program_text` - (Required) Signalflow program text for the detector. More info at <https://developers.signalfx.com/docs/signalflow-overview>.
* `description` - (Optional) Description of the detector.
* `max_delay` - (Optional) How long (in seconds) to wait for late datapoints. See <https://signalfx-product-docs.readthedocs-hosted.com/en/latest/charts/chart-builder.html#delayed-datapoints> for more info. Max value is `900` seconds (15 minutes).
* `show_data_markers` - (Optional) When `true`, markers will be drawn for each datapoint within the visualization. `false` by default.
* `time_range` - (Optional) From when to display data. SignalFx time syntax (e.g. `"-5m"`, `"-1h"`). Conflicts with `start_time` and `end_time`.
* `start_time` - (Optional) Seconds since epoch. Used for visualization. Conflicts with `time_range`.
* `end_time` - (Optional) Seconds since epoch. Used for visualization. Conflicts with `time_range`.
* `tags` - (Optional) Tags associated with the detector.
* `teams` - (Optional) Team IDs to associcate the detector to.
* `rule` - (Required) Set of rules used for alerting.
    * `detect_label` - (Required) A detect label which matches a detect label within `program_text`.
    * `severity` - (Required) The severity of the rule, must be one of: `"Critical"`, `"Major"`, `"Minor"`, `"Warning"`, `"Info"`.
    * `disabled` - (Optional) When true, notifications and events will not be generated for the detect label. `false` by default.
    * `notifications` - (Optional) List of strings specifying where notifications will be sent when an incident occurs. See <https://developers.signalfx.com/v2/reference#section-notifications> for more info.
    * `parameterized_body` - (Optional) Custom notification message body when an alert is triggered. See <https://developers.signalfx.com/v2/reference#section-custom-notification-messages> for more info.
    * `parameterized_subject` - (Optional) Custom notification message subject when an alert is triggered. See <https://developers.signalfx.com/v2/reference#section-custom-notification-messages> for more info.
    * `runbook_url` - (Optional) URL of page to consult when an alert is triggered. This can be used with custom notification messages.
    * `tip` - (Optional) Plain text suggested first course of action, such as a command line to execute. This can be used with custom notification messages.

**Notes**

It is highly recommended that you use both `max_delay` in your detector configuration and an `extrapolation` policy in your program text to reduce false positives/negatives.

`max_delay` allows SignalFx to continue with computation if there is a lag in receiving data points.

`extrapolation` allows you to specify how to handle missing data. An extrapolation policy can be added to individual signals by updating the data block in your `program_text`.

See <https://signalfx-product-docs.readthedocs-hosted.com/en/latest/charts/chart-builder.html#delayed-datapoints> for more info.
