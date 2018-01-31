# Single Value Chart

This chart type displays a single number in a large font, representing the current value of a single metric on a plot line.

If the time period is in the past, the number represents the value of the metric near the end of the time period.

![Single Value Chart](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/resources/single_value_chart.png)

## Example Usage

```terraform
resource "signalform_single_value_chart" "mysvchart0" {
    name = "CPU Total Idle - Single Value"

    program_text = <<-EOF
        myfilters = filter("cluster_name", "prod") and filter("role", "search")
        data("cpu.total.idle", filter=myfilters).publish()
        EOF

    description = "Very cool Single Value Chart"

    color_by = "Dimension"

    max_delay = 2
    refresh_interval = 1
    max_precision = 2
    is_timestamp_hidden = true
    show_spark_line = true
}
```


## Argument Reference

The following arguments are supported in the resource block:

* `name` - (Required) Name of the chart.
* `program_text` - (Required) Signalflow program text for the chart. More info at <https://developers.signalfx.com/docs/signalflow-overview>.
* `description` - (Optional) Description of the chart.
* `color_by` - (Optional) Must be `"Dimension"` or `"Metric"`. `"Dimension"` by default.
* `color_scale` - (Optional. `color_by` must be `"Scale"`) Single color range including both the color to display for that range and the borders of the range. Example: `[{ gt : 60, color : blue }, { lte : 60, color : yellow }]`. Look at this [link](https://docs.signalfx.com/en/latest/charts/chart-options-tab.html).
    * `gt` - (Optional) Indicates the lower threshold non-inclusive value for this range.
    * `gte` - (Optional) Indicates the lower threshold inclusive value for this range.
    * `lt` - (Optional) Indicates the upper threshold non-inculsive value for this range.
    * `lte` - (Optional) Indicates the upper threshold inclusive value for this range.
    * `color` - (Required) The color range to use. Must be either gray, blue, navy, orange, yellow, magenta, purple, violet, lilac, green, aquamarine. ![Colors](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/resources/colors.png)
* `unit_prefix` - (Optional) Must be `"Metric"` or `"Binary"`. `"Metric"` by default.
* `max_delay - (Optional) How long (in seconds) to wait for late datapoints
* `refresh_interval` - (Optional) How often (in seconds) to refresh the value.
* `max_precision` - (Optional) The maximum precision to for value displayed.
* `is_timestamp_hidden` - (Optional) Whether to hide the timestamp in the chart. `false` by default.
* `show_spark_line` - (Optional) Whether to show a trend line below the current value. `false` by default.
* `synced` - (Optional) Whether the resource in SignalForm and SignalFx are identical or not. Used internally for syncing, you do not need to specify it. Whenever you see a change to this field in the plan, it means that your resource has been changed from the UI and Terraform is now going to re-sync it back to what is in your configuration.
