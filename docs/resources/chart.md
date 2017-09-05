# Chart

Charts enable you to visualize the metrics you are sending in to SignalFx.

![Chart](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/resources/chart.jpg)

## Choosing a chart type

SignalFx available chart types are the following:

* [Time Chart](resources/time_chart.md)
* [List Chart](resources/list_chart.md)
* [Single Value Chart](resources/single_value_chart.md)
* [Heatmap Chart](resources/heatmap_chart.md)
* [Text Note](resources/text_node.md)

Time chart is the only chart type that includes four different visualization options for SignalFx graphs (image below): Line Chart, Column Chart, Area Chart and Histogram Chart.

![Time Chart Types](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/resources/time_chart_types.png)

Just note that if you want to create Area Chart, you need to create a Time Chart Resource and set the property `plot_type = "AreaChart"` (more info [here](resources/time_chart.md)).
