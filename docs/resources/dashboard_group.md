# Dashboard Group

In the SignalFx web UI, a [dashboard group](https://developers.signalfx.com/v2/docs/dashboard-group-model) is a collection of dashboards.

**NOTE:** Dashboard groups cannot be accessed directly, but just via a dashboard contained in them. This is the reason why make show won't show any of yours dashboard groups.


## Example Usage

```terraform
resource "signalform_dashboard_group" "mydashboardgroup0" {
    name = "My team dashboard group"
    description = "Cool dashboard group"
}
```

## Argument Reference

The following arguments are supported in the resource block:

* `name` - (Required) Name of the dashboard group.
* `description` - (Required) Description of the dashboard group.
* `teams` - (Optional) Team IDs to associate the dashboard group to.
* `synced` - (Optional) Whether the resource in SignalForm and SignalFx are identical or not. Used internally for syncing, you don not need to specify it. Whenever you see a change to this field in the plan, it means that your resource has been changed from the UI and Terraform is now going to re-sync it back to what is in your configuration.
