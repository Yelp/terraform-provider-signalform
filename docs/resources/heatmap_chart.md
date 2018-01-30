# Heatmap Chart

This chart type displays the specified plot in a heatmap fashion. This format is similar to the [Infrastructure Navigator](https://signalfx-product-docs.readthedocs-hosted.com/en/latest/built-in-content/infra-nav.html#infra), with squares representing each source for the selected metric, and the color of each square representing the value range of the metric.

![Heatmap Chart](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/resources/heatmap_chart.png)


## Example Usage

```terraform
resource "signalform_heatmap_chart" "myheatmapchart0" {
    name = "CPU Total Idle - Heatmap"

    program_text = <<-EOF
        myfilters = filter("cluster_name", "prod") and filter("role", "search")
        data("cpu.total.idle", filter=myfilters).publish()
        EOF

    description = "Very cool Heatmap"

    disable_sampling = true
    sort_by = "+host"
    group_by = ["hostname", "host"]
    hide_timestamp = true
}
```


## Argument Reference

The following arguments are supported in the resource block:

* `name` - (Required) Name of the chart.
* `program_text` - (Required) Signalflow program text for the chart. More info at <https://developers.signalfx.com/docs/signalflow-overview>.
* `description` - (Optional) Description of the chart.
* `unit_prefix` - (Optional) Must be `"Metric"` or `"Binary`". `"Metric"` by default.
* `minimum_resolution` - (Optional) The minimum resolution (in seconds) to use for computing the underlying program.
* `max_delay - (Optional) How long (in seconds) to wait for late datapoints.
* `disable_sampling` - (Optional) If `false`, samples a subset of the output MTS, which improves UI performance. `false` by default.
* `group_by` - (Optional) Properties to group by in the heatmap (in nesting order).
* `sort_by` - (Optional) The property to use when sorting the elements. Must be prepended with `+` for ascending or `-` for descending (e.g. `-foo`).
* `hide_timestamp` - (Optional) Whether to show the timestamp in the chart. `false` by default.
* `color_range` - (Optional. Conflict with color_scale) Values and color for the color range. Example: `color_range : { min : 0, max : 100, color : blue }`. Look at this [link](https://docs.signalfx.com/en/latest/charts/chart-options-tab.html).
    * `min_value` - (Optional) The minimum value within the coloring range.
    * `max_value` - (Optional) The maximum value within the coloring range.
    * `color` - (Required) The color range to use. Must be either gray, blue, navy, orange, yellow, magenta, purple, violet, lilac, green, aquamarine. ![Colors](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/resources/colors.png)
* `color_scale` - (Optional. Conflict with `color_range`) Single color range including both the color to display for that range and the borders of the range. Example: `[{ gt : 60, color : blue }, { lte : 60, color : yellow }]`. Look at this [link](https://docs.signalfx.com/en/latest/charts/chart-options-tab.html).
    * `gt` - (Optional) Indicates the lower threshold non-inclusive value for this range.
    * `gte` - (Optional) Indicates the lower threshold inclusive value for this range.
    * `lt` - (Optional) Indicates the upper threshold non-inculsive value for this range.
    * `lte` - (Optional) Indicates the upper threshold inclusive value for this range.
    * `color` - (Required) The color range to use. Must be either gray, blue, navy, orange, yellow, magenta, purple, violet, lilac, green, aquamarine. ![Colors](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/resources/colors.png)
* `synced` - (Optional) Whether the resource in SignalForm and SignalFx are identical or not. Used internally for syncing, you do not need to specify it. Whenever you see a change to this field in the plan, it means that your resource has been changed from the UI and Terraform is now going to re-sync it back to what is in your configuration.
