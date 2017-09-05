# List Chart

This chart type displays current data values in a list format.

![List Chart](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/resources/list_chart.png)

The name of each value in the chart reflects the name of the plot and any associated dimensions. We recommend you click the Pencil icon and give the plot a meaningful name, as in plot B below. Otherwise, just the raw metric name will be displayed on the chart, as in plot A below.


## Example Usage

```terraform
resource "signalform_list_chart" "mylistchart0" {
    name = "CPU Total Idle - List"

    program_text = <<-EOF
    myfilters = filter("cluster_name", "prod") and filter("role", "search")
    data("cpu.total.idle", filter=myfilters).publish()
    EOF

    description = "Very cool List Chart"

    color_by = "Metric"
    max_delay = 2
    disable_sampling = true
    refresh_interval = 1
    legend_fields_to_hide = ["collector", "host"]
    max_precision = 2
    sort_by = "-value"
 }
```

## Argument Reference

The following arguments are supported in the resource block:

* `name` - (Required) Name of the chart.
* `program_text` - (Required) Signalflow program text for the chart. More info at <https://developers.signalfx.com/docs/signalflow-overview>.
* `description` - (Optional) Description of the chart.
* `unit_prefix` - (Optional) Must be `"Metric"` or `"Binary`". `"Metric"` by default.
* `color_by` - (Optional) Must be `"Dimension"` or `"Metric"`. `"Dimension"` by default.
* `max_delay - (Optional) How long (in seconds) to wait for late datapoints.
* `disable_sampling` - (Optional) If `false`, samples a subset of the output MTS, which improves UI performance. `false` by default.
* `refresh_interval` - (Optional) How often (in seconds) to refresh the values of the list.
* `legend_fields_to_hide` - (Optional) List of properties that should not be displayed in the chart legend (i.e. dimension names). All the properties are visible by default.
* `max_precision` - (Optional) Maximum number of digits to display when rounding values up or down.
* `sort_by` - (Optional) The property to use when sorting the elements. Use `value` if you want to sort by value. Must be prepended with `+` for ascending or `-` for descending (e.g. `-foo`).
* `synced` - (Optional) Whether the resource in SignalForm and SignalFx are identical or not. Used internally for syncing, you do not need to specify it. Whenever you see a change to this field in the plan, it means that your resource has been changed from the UI and Terraform is now going to re-sync it back to what is in your configuration.
