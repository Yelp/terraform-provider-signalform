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

Argument Reference

The following arguments are supported in the resource block:

* `name` - (Required) Name of the chart.
* `program_text` - (Required) Signalflow program text for the chart. More info at https://developers.signalfx.com/docs/signalflow-overview.

* `description` - (Optional) Description of the chart.
* `color_by` - (Optional) Must be `"Dimension"` or `"Metric"`. `"Dimension"` by default.
* `color_scale` - (Optional. `color_by` must be `"Scale"`) Values for each color in the range. Example: `{ thresholds : [ 80, 60, 40, 20, 0 ], inverted : true }`. Look at this [link](https://docs.signalfx.com/en/latest/charts/chart-options-tab.html).
    * `thresholds` - (Required) The thresholds to set for the color range being used. Values (at most 4) must be in descending order.
    * `inverted` - (Optional) If false or omitted, values are red if they are above the highest specified value. If `true`, values are red if they are below the lowest specified value. `false` by default.
* `unit_prefix` - (Optional) Must be `"Metric"` or `"Binary"`. `"Metric"` by default.
* `max_delay - (Optional) How long (in seconds) to wait for late datapoints
* `refresh_interval` - (Optional) How often (in seconds) to refresh the value.
* `max_precision` - (Optional) The maximum precision to for value displayed.
* `is_timestamp_hidden` - (Optional) Whether to hide the timestamp in the chart. `false` by default.
* `show_spark_line` - (Optional) Whether to show a trend line below the current value. `false` by default.
* `synced` - (Optional) Whether the resource in SignalForm and SignalFx are identical or not. Used internally for syncing, you do not need to specify it. Whenever you see a change to this field in the plan, it means that your resource has been changed from the UI and Terraform is now going to re-sync it back to what is in your configuration.
