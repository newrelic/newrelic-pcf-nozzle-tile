# **New Relic VMware Tanzu Nozzle Tile**
​
The New Relic VMware Tanzu (ex PCF) Nozzle Tile is a Firehose nozzle that forwards metrics from [VMware Tanzu Loggregator][a] in [VMware Tanzu][b] to [New Relic][c] for visualization.
​
This code can be either pushed as a regular VMware Tanzu application with `cf push` or installed in Ops Manager using the tile version.
​
The tile is also available in the [Pivotal Network][f] alongside the [documentation][h] describing how to configure and use the nozzle.
​
## **Compatibility**
​
The New Relic VMware Tanzu Nozzle Tile is compatible with VMware Tanzu **2.4** and higher.
​
## **Changes From V1 to V2**
​
The V2 release includes several additional features as well as **breaking changes**. Deployment configurations, alerts, and dashboards might require updates. [Full details are available in the V2 notes](/V2.md).
​
### **Main updates**
​
- Reverse Log Proxy Gateway and V2 Envelope Format
- Event type adjustments
- Event attribute modifications
- Event aggregation - metric type events
- Multi-account event routing
- Caching and rate limiting - VMware Tanzu API calls
- Configuration variable changes
- Log message filters
- Metric type filters removed
- Graceful shutdown
​
## **Application build and deploy**
​
The application is [prebuilt][f] and can be [pushed as an application][j] or [imported to Ops Manager][k]. If you make any changes to the code, you can rebuild both the binary and the tile.
​
### Build the binary
​
1. Get `dep` to manage dependencies:
```bash
$ go get -u github.com/golang/dep/cmd/dep
```
2. Generate the `nr-fh-nozzle` binary inside `./dist`:
```bash
$ make build-linux
```
You can then deploy the application with `cf push` using the newly generated files, as described in the [push as an application][j] section.
​
### Generate the tile
​
1. Install [bosh-cli](https://github.com/cloudfoundry/bosh-cli/releases) and [tile-generator](https://github.com/cf-platform-eng/tile-generator/releases/).
​
2. Generate the tile under `./product`.
```bash
$ make release
```
You can  use the generated tile right away and [import it to Ops Manager][k]. 
​
## **Local testing**
​
### Requirements
​
- Access to a VMware Tanzu environment
- VMware Tanzu API credentials with admin rights
- VMware Tanzu [UAA authorized client][l]
​
### Setup
​
1. Set your environment variables as per the `manifest.yml.sample` file.
2. Run `go run main.go`
​
To run tests and compile locally:
```bash
$ make build
```
### Generate UAAC Client
​
You can create a new `doppler.firehose` enabled client instead of retrieving the default client:
        
```bash
$ uaac target https://uaa.[your cf system domain]
$ uaac token client get admin -s [your admin-secret]
$ uaac client add firehose-to-newrelic \
    --name firehose-to-newrelic \
    --secret [your_client_secret] \
    --authorized_grant_types client_credentials,refresh_token \
    --authorities doppler.firehose,cloud_controller.admin_read_only \
    --scope doppler.firehose
```
​
* `firehose-to-newrelic`: your `NRF_CF_CLIENT_ID` env variable.
* `--secret`: your `NRF_CF_CLIENT_SECRET` env variable.
​
## **Push as an application**
​
When you push the app as an application, you must edit `manifest.yml` first
1. Download the `manifest.yml.sample` file and the [release][f] from the repo. 
2. Unzip the release, rename `manifest.yml.sample` to `manifest.yml` and place the file in the `dist` directory. 
3. Modify the manifest file to match your environment.
4. Deploy: 
```bash
cf push -f <manifest file>
``` 
Make sure to assign proper values to all required environment variables. Any property values within angle brackets need to be changed to the correct value for your environment.
​
>**Note:** When you're pushing the nozzle as an app, the `product` and `release` folders are not required. Make sure to remove both folders to reduce the size of the upload, or use the `.cfignore` file.
​
​
## Import as a tile in Ops Manager
​
Import the tile from [releases][f] to Ops Mgr. Once imported, install the tile, and follow the steps detailed in the [Pivotal Partner Docs][i].
​
## Import dashboard

A VMware Tanzu dashboard could be manually imported to New Relic One Dashboards. You can follow this steps:
1. Modify the [dashboard.json](/dashboard/dashboard.json) use your user account id to replace `"dashboard_account_id"=<NRF_NEWRELIC_ACCOUNT_ID>` 
2. Go to [New Relic One Dashboards](https://one.newrelic.com/launcher/dashboards.launcher) and use the `import a dashboard` function on the top right.
3. Copy and Paste the modified [dashboard.json](/dashboard/dashboard.json) and import the dashboard.

>**Note:** Only Administrator user can import dashboards.


## Support

You can find more detailed documentation [on our website](http://newrelic.com/docs),and specifically in the [Infrastructure category](https://docs.newrelic.com/docs/infrastructure).

If you can't find what you're looking for there, reach out to us on our [support site](http://support.newrelic.com/) or our [community forum](http://forum.newrelic.com) and we'll be happy to help you.

Find a bug? Contact us via [support.newrelic.com](http://support.newrelic.com/), or email support@newrelic.com.

New Relic, Inc.

​
### Community
​
New Relic hosts and moderates an online forum where customers can interact with New Relic employees as well as other customers to get help and share best practices. Like all official New Relic open source projects, there's a related Community topic in the New Relic Explorers Hub. You can find this project's topic/threads here:
​
https://discuss.newrelic.com/t/new-relic-firehose-nozzle-for-pcf/92709
​
### Issues / Enhancement Requests
​
Issues and enhancement requests can be submitted in the [Issues tab of this repository](../../issues). Please search for and review the existing open issues before submitting a new issue.
​
## **License**
​
The project is released under version 2.0 of the [Apache License][e].

[a]: https://docs.cloudfoundry.org/loggregator/architecture.html
[b]: https://pivotal.io/platform
[c]: http://newrelic.com/insights
[d]: manifest.yml
[e]: http://www.apache.org/licenses/LICENSE-2.0
[f]: https://github.com/newrelic/newrelic-pcf-nozzle-tile/releases
[g]: https://network.pivotal.io/products/nr-firehose-nozzle/
[h]: https://docs.pivotal.io/partners/new-relic-nozzle/index.html
[i]: https://docs.pivotal.io/partners/new-relic-nozzle/installing.html
[j]: #push-as-an-application
[k]: #import-as-a-tile-in-ops-manager
[l]: #generate#uaac#client
