[![New Relic Community Plus header](https://raw.githubusercontent.com/newrelic/open-source-office/master/examples/categories/images/Community_Plus.png)](https://opensource.newrelic.com/oss-category/#community-plus)

# New Relic VMware Tanzu Nozzle Tile

The New Relic VMware Tanzu (PCF) Nozzle Tile is a Firehose nozzle that forwards metrics from [VMware Tanzu Loggregator][a] in [VMware Tanzu][b] to [New Relic][c] for visualization.
​
This code can be either pushed as a regular VMware Tanzu application with `cf push` or installed in Ops Manager using the tile version.
​
The tile is also available in the [Pivotal Network][f] alongside the [documentation][h] describing how to configure and use the nozzle.
​
See our [documentation](https://docs.newrelic.com/docs/vmwaretanzu-integration-new-relic-infrastructure) for more details.

## Compatibility
​
The New Relic VMware Tanzu Nozzle Tile is compatible with VMware Tanzu **2.4** and higher.
​
## Changes From V1 to V2
​
The V2 release includes several additional features as well as **breaking changes**. Deployment configurations, alerts, and dashboards might require updates. [Full details are available in the V2 notes](/V2.md).
​
### Main updates
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
## Application build and deploy
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
## Testing
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
## Push as an application
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
>When you're pushing the nozzle as an app, the `product` and `release` folders are not required. Make sure to remove both folders to reduce the size of the upload, or use the `.cfignore` file.
​
​
## Import as a tile in Ops Manager
​
Import the tile from [releases][f] to Ops Mgr. Once imported, install the tile, and follow the steps detailed in the [Pivotal Partner Docs][i].
​
## Import dashboard

A VMware Tanzu dashboard could be manually imported to New Relic dashboards using the Dashboard API. Follow [this documentation](https://docs.newrelic.com/docs/insights/insights-api/manage-dashboards/insights-dashboard-api) to get detailed information about where to obtain the Admin user API key and to use the API explorer.

1. Go to [API Explorer](https://rpm.newrelic.com/api/explore/dashboards/create)
2. Use your [Admin user API key](https://docs.newrelic.com/docs/apis/get-started/intro-apis/types-new-relic-api-keys#admin-api)
3. Copy the content of [dashboard.json](/dashboard/dashboard.json) and paste to the dashboard parameter of the request.
4. Send the request

## Support

Should you need assistance with New Relic products, you are in good hands with several support diagnostic tools and support channels.

>This [troubleshooting framework](https://discuss.newrelic.com/t/troubleshooting-frameworks/108787) steps you through common troubleshooting questions.

>New Relic offers NRDiag, [a client-side diagnostic utility](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics) that automatically detects common problems with New Relic agents. If NRDiag detects a problem, it suggests troubleshooting steps. NRDiag can also automatically attach troubleshooting data to a New Relic Support ticket. Remove this section if it doesn't apply.

If the issue has been confirmed as a bug or is a feature request, file a GitHub issue.

**Support Channels**

* [New Relic Documentation](https://docs.newrelic.com): Comprehensive guidance for using our platform
* [New Relic Community](https://discuss.newrelic.com/t/new-relic-firehose-nozzle-for-pcf/92709): The best place to engage in troubleshooting questions
* [New Relic Developer](https://developer.newrelic.com/): Resources for building a custom observability applications
* [New Relic University](https://learn.newrelic.com/): A range of online training for New Relic users of every level
* [New Relic Technical Support](https://support.newrelic.com/) 24/7/365 ticketed support. Read more about our [Technical Support Offerings](https://docs.newrelic.com/docs/licenses/license-information/general-usage-licenses/support-plan).

## Privacy

At New Relic we take your privacy and the security of your information seriously, and are committed to protecting your information. We must emphasize the importance of not sharing personal data in public forums, and ask all users to scrub logs and diagnostic information for sensitive information, whether personal, proprietary, or otherwise.

We define “Personal Data” as any information relating to an identified or identifiable individual, including, for example, your name, phone number, post code or zip code, Device ID, IP address, and email address.

For more information, review [New Relic’s General Data Privacy Notice](https://newrelic.com/termsandconditions/privacy).

## Contribute

We encourage your contributions to improve this project! Keep in mind that when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.

If you have any questions, or to execute our corporate CLA (which is required if your contribution is on behalf of a company), drop us an email at opensource@newrelic.com.

**A note about vulnerabilities**

As noted in our [security policy](/security/policy), New Relic is committed to the privacy and security of our customers and their data. We believe that providing coordinated disclosure by security researchers and engaging with the security community are important means to achieve our security goals.

If you believe you have found a security vulnerability in this project or any of New Relic's products or websites, we welcome and greatly appreciate you reporting it to New Relic through [HackerOne](https://hackerone.com/newrelic).

If you would like to contribute to this project, review [these guidelines](./CONTRIBUTING.md).

To all contributors, we thank you!  Without your contribution, this project would not be what it is today.
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
