# Time Chart

Time charts display datpoints over a period of time.

![Time Chart](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/resources/time_chart.png)

The first four icons in the chart’s title bar represent four visualization options for time charts: line chart, area chart, column chart, and histogram chart (user the `plot_type` to choose your favorite chart type).

![Time Chart Types](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/resources/time_chart_types.jpg)


## Example Usage

```terraform
resource "signalform_time_chart" "mychart0" {
    name = "CPU Total Idle"

    program_text = <<-EOF
        myfilters = filter("shc_name", "prod") and filter("role", "splunk_searchhead")
        data("cpu.total.idle", filter=myfilters).publish(label="CPU Idle")
        EOF

    time_range = "-15m"

    plot_type = "LineChart"
    show_data_markers = true

    legend_fields_to_hide = ["collector", "prefix", "hostname"]
    viz_options {
        label = "CPU Idle"
        axis = "left"
        color = "orange"
    }

    axis_left {
        label = "CPU Total Idle"
        low_watermark = 1000
    }
}
```


## Argument Reference

The following arguments are supported in the resource block:

* `name` - (Required) Name of the chart.
* `program_text` - (Required) Signalflow program text for the chart. More info at <https://developers.signalfx.com/docs/signalflow-overview>.
* `plot_type` - (Optional) The default plot display style for the visualization. Must be `"LineChart"`, `"AreaChart"`, `"ColumnChart"`, or `"Histogram"`. Default: `"LineChart"`.
* `description` - (Optional) Description of the chart.
* `unit_prefix` - (Optional) Must be `"Metric"` or `"Binary`". `"Metric"` by default.
* `color_by` - (Optional) Must be `"Dimension"` or `"Metric"`. `"Dimension"` by default.
* `minimum_resolution` - (Optional) The minimum resolution (in seconds) to use for computing the underlying program.
* `max_delay - (Optional) How long (in seconds) to wait for late datapoints.
* `disable_sampling` - (Optional) If `false`, samples a subset of the output MTS, which improves UI performance. `false` by default
* `time_range` - (Optional) From when to display data. SignalFx time syntax (e.g. `"-5m"`, `"-1h"`). Conflicts with `start_time` and `end_time`.
* `start_time` - (Optional) Seconds since epoch. Used for visualization. Conflicts with `time_range`.
* `end_time` - (Optional) Seconds since epoch. Used for visualization. Conflicts with `time_range`.
* `axes_include_zero` - (Optional) Force the chart to display zero on the y-axes, even if none of the data is near zero.
* `axis_left` - (Optional) Set of axis options.
    * `label` - (Optional) Label of the left axis.
    * `min_value` - (Optional) The minimum value for the left axis.
    * `max_value` - (Optional) The maximum value for the left axis.
    * `high_watermark` - (Optional) A line to draw as a high watermark.
    * `high_watermark_label` - (Optional) A label to attach to the high watermark line.
    * `low_watermark`  - (Optional) A line to draw as a low watermark.
    * `low_watermark_label` - (Optional) A label to attach to the low watermark line.
* `axis_right` - (Optional) Set of axis options.
    * `label` - (Optional) Label of the right axis.
    * `min_value` - (Optional) The minimum value for the right axis.
    * `max_value` - (Optional) The maximum value for the right axis.
    * `high_watermark` - (Optional) A line to draw as a high watermark.
    * `high_watermark_label` - (Optional) A label to attach to the high watermark line.
    * `low_watermark`  - (Optional) A line to draw as a low watermark.
    * `low_watermark_label` - (Optional) A label to attach to the low watermark line.
* `viz_options` - (Optional) Plot-level customization options, associated with a publish statement.
    * `label` - (Required) Label used in the publish statement that displays the plot (metric time series data) you want to customize.
    * `color` - (Optional) Color to use : gray, blue, azure, navy, brown, orange, yellow, iris, magenta, pink, purple, violet, lilac, emerald, green, aquamarine. ![Colors](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/resources/colors.png)
    * `axis` - (Optional) Y-axis associated with values for this plot. Must be either `right` or `left`.
    * `plot_type` - (Optional) The visualization style to use. Must be `"LineChart"`, `"AreaChart"`, `"ColumnChart"`, or `"Histogram"`. Chart level `plot_type` by default.
    * `value_unit` - (Optional) A unit to attach to this plot. Units support automatic scaling (eg thousands of bytes will be displayed as kilobytes).
    * `value_prefix`, `value_suffix` - (Optional) Arbitrary prefix/suffix to display with the value of this plot.
* `legend_fields_to_hide` - (Optional) List of properties that should not be displayed in the chart legend (i.e. dimension names). All the properties are visible by default.
* `on_chart_legend_dimension` - (Optional) Dimensions to show in the on-chart legend. On-chart legend is off unless a dimension is specified. Allowed: `"metric"`, `"plot_label"` and any dimension.
* `show_event_lines` - (Optional) Whether vertical highlight lines should be drawn in the visualizations at times when events occurred. `false` by default.
* `show_data_markers` - (Optional) Show markers (circles) for each datapoint used to draw line or area charts. `false` by default.
* `stacked` - (Optional) Whether area and bar charts in the visualization should be stacked. `false` by default.
* `synced` - (Optional) Whether the resource in SignalForm and SignalFx are identical or not. Used internally for syncing, you do not need to specify it. Whenever you see a change to this field in the plan, it means that your resource has been changed from the UI and Terraform is now going to re-sync it back to what is in your configuration.
