# terraform-provider-signalform

[![GitHub version](https://badge.fury.io/gh/Yelp%2Fterraform-provider-signalform.svg)](https://badge.fury.io/gh/Yelp%2Fterraform-provider-signalform)
[![Build Status](https://travis-ci.org/Yelp/terraform-provider-signalform.svg?branch=master)](https://travis-ci.org/Yelp/terraform-provider-signalform)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

SignalForm is a terraform provider to codify [SignalFx](http://signalfx.com) detectors, charts and dashboards, thereby making it easier to create, manage and version control them.

Signalform is **not** an official SignalFx product, so we do not guarantee a 1:1 mapping between the SignalFx API and the functionalities offered by the provider.

Please note that this provider is tested against Terraform **0.10**. We do not guarantee that the provider works correctly with any other version of Terraform, even though it might.

Documentation is available [here](https://yelp.github.io/terraform-provider-signalform/).

Changelog is available [here](https://github.com/Yelp/terraform-provider-signalform/blob/master/build/changelog).


* Resources
    * [Detector](https://yelp.github.io/terraform-provider-signalform/resources/detector.html)
    * [Chart](https://yelp.github.io/terraform-provider-signalform/resources/chart.html)
        * [Time Chart](https://yelp.github.io/terraform-provider-signalform/resources/time_chart.html)
        * [List Chart](https://yelp.github.io/terraform-provider-signalform/resources/list_chart.html)
        * [Single Value Chart](https://yelp.github.io/terraform-provider-signalform/resources/single_value_chart.html)
        * [Heatmap Chart](https://yelp.github.io/terraform-provider-signalform/resources/heatmap_chart.html)
        * [Text Note](https://yelp.github.io/terraform-provider-signalform/resources/text_note.html)
    * [Dashboard](https://yelp.github.io/terraform-provider-signalform/resources/dashboard.html)
    * [Dashboard Group](https://yelp.github.io/terraform-provider-signalform/resources/dashboard_group.html)
    * [Integration](https://yelp.github.io/terraform-provider-signalform/resources/integration.html)
* [Build And Install](#build-and-install)
    * [Build binary from source](#build-binary-from-source)
    * [Build debian package from source](#build-debian-package-from-source)
* [Release](#release)
* [Contributing](#contributing)
* [FAQ](#faq)


## Build And Install

### Build binary from source

To build the go binary from source:

```bash
make build
```

The output binary will be placed in the `bin/` folder.

The default platform target is `linux-amd64`. If you want to customize your target platform set the `GOOS` and `GOARCH` environment variables; e.g.:
```bash
GOOS=darwin GOARCH=amd64 make build
```

Once you have built the binary, place it in the same folder of your terraform installation for it to be available everywhere.

### Build debian package from source

If you want to build your package targeting at the same time Ubuntu Trusty and Xenial, then run:
```shell
make package
```

The output package will be placed in the `dist/` folder (e.g. `dist/terraform-provider-signalform-0.9_2.2.5-default0_amd64.deb`)

If you want to target a specific Ubuntu version, use the `itest_*` command as below:
```shell
make itest_trusty
```

You can set environament variables to customize your build:

* `TF_PATH`: Path of your terraform installation. Default: `/nail/opt/terraform-$(TF_VERSION)`, with `TF_VERSION` being the one supported by the provider.
* `ORG`: Organization name to be used for your package name. Default: `default` (e.g. `terraform-provider-signalform-0.9_2.2.5-$(ORG)0_amd64.deb`)
* `BUILD_NUMBER`: This variable will be used as iteration number. If not set, the default will be `0` (e.g. `terraform-provider-signalform-0.9_2.2.5-default$(BUILD_NUMBER)_amd64.deb`)
* `upstream_build_number`: If `BUILD_NUMBER` is not set and if your CI pipeline (e.g. Jenkins) defines this variable, then the job id will be used as iteration number for your package (e.g. `terraform-provider-signalform-0.9_2.2.5-default$(upstream_build_number)_amd64.deb`)
* `GOOS` and `GOARCH`: see the above [section](#build-binary-from-source). Remember to run `make clean` between builds with different target platform, so you start from a clean environment.

Once you built the package, you can just install like:
```shell
sudo dpkg -i dist/terraform-provider-signalform.deb
```

## Release

To make a new release:

1. bump `VERSION` up in `build/Makefile` (use [semantic versioning](http://semver.org/))
1. run `make changelog` and edit the changelog file
1. `git commit`
1. `git tag v<VERSION>`
1. `git push origin master && git push origin --tags`


## Contributing
Everyone is encouraged to contribute to `terraform-provider-signalform`. You can contribute by forking the GitHub repo and making a pull request or opening an issue.

## Running tests

To run the tests, run `make test`

To pass options to the test, e.g to target a specific test, pass the TEST_OPTS
option to the make task - e.g

```
make test TEST_OPTS='-test.run TestSendRequestSuccess'
```

To make the tests run faster, you can run

```
make test TEST_OPTS='-i'
```

Subsequent make test commands should be quicker

## FAQ

**Why not calling it terraform-provider-signalfx?**

Signalform is *not* an official SignalFx product, being owned and maintained by Yelp. For this reason, we decided not to call this provider terraform-provider-signalfx in case SignalFx decides to publish an official one.

**SignalFlow is hard!**

It is a bit hard, indeed. You might find useful to read the [SignalFlow Overview](https://developers.signalfx.com/docs/signalflow-overview).

Also remember that given a chart or detector created from the UI, you can see its representation in Signalflow from the Actions menu:

![Show SignalFlow](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/show_signalflow.png)
![Signalflow](https://github.com/Yelp/terraform-provider-signalform/raw/master/docs/signalflow.png)
