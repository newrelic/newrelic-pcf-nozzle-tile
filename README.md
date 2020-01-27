# **New Relic PCF Nozzle Tile**

This application is a Firehose nozzle that forwards metrics from the [PCF Loggregator][a] in [Pivotal Cloud Foundry][b] into [New Relic Insights][c] for visualization.

This code can either be pushed as a regular PCF application with `cf push`, or you can use the tile version, and install it in Ops Manager.

## **Changes From V1 to V2**

The V2 release includes several additional features as well as a few **breaking changes**. Deployment configurations, alerts, and dashboards might require updates. [Additional details for these updates are available](/V2.md).

### **Updates**

- Reverse Log Proxy Gateway and V2 Envelope Format
- Event type adjustments
- Event attribute modifications
- Event aggregation - metric type events
- Multi-account event routing
- Caching and rate limiting - PCF API calls
- Configuration variable changes
- Log message filters
- Metric type filters removed
- Graceful shutdown

## **Push as an application**

When you push the app as an application, you need to have a [manifest][d] with the following properties. The process:
1. Download the manifest file and the release from the repo. 
2. Unzip the release, and place the manifest file in the `dist` directory. 
3. Modify the manifest file to match your environment, and then deploy using `cf push -f <manifest file>`. 

Some of these properties are automatically set when you deploy as a tile.

>	---
	applications:
	- name: newrelic-firehose-nozzle
	  memory: 512M
      disk_quota: 256M
	  instances: 2
	  health-check-type: http
    health-check-http-endpoint: /health
	  host: cf-firehose-nozzle-${random-word}
          buildpacks:
          - binary_buildpack
          path: dist
          command: ./nr-fh-nozzle
    env:
        NRF_CF_CLIENT_ID: "Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Opentsdb Nozzle Credentials -> Link to Credential -> identity"
        NRF_CF_CLIENT_SECRET: "Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Opentsdb Nozzle Credentials -> Link to Credential -> password"
        NRF_CF_API_UAA_URL: "run cf curl /v2/info to get the url"
        NRF_CF_API_URL: "run cf curl /v2/info to get the url"
        NRF_FIREHOSE_ID: newrelic.firehose
        NRF_CF_SKIP_SSL: true
        NRF_ENABLED_ENVELOPE_TYPES: ValueMetric,CounterEvent,LogMessage,ContainerMetric,HttpStartStop

        NRF_CF_API_USERNAME: "Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Admin Credentials -> Link to Credential -> identity"
        NRF_CF_API_PASSWORD: "Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Admin Credentials -> Link to Credential -> password"
        NRF_NEWRELIC_ACCOUNT_ID: New Relic Account ID
        NRF_NEWRELIC_INSERT_KEY: New Relic Insights Insert Key
        
        # # Optional Settings (with their default values listed).  Uncomment the setting to change.
        # # New Relic account region.  Choose EU if RPM URL includes .eu.
        # NRF_NEWRELIC_ACCOUNT_REGION: US or EU
        # # How often accumulated metric events are sent.  Recommended: 29s, 59s, 89s, or 129s
        # NRF_NEWRELIC_DRAIN_INTERVAL: 59s
        # # Number of minutes before the HTTP connection to the RLP Gateway is considered hung and restarted. The RLP Gateway should force a new connection every 14 minutes. This is only applicable if the connection hangs.
        # NRF_FIREHOSE_HTTP_TIMEOUT_MINS: 16
        # # Number of consecutive seconds with no messages before the nozzle is automatically restarted. Set per environment based on normal message load.
        # NRF_FIREHOSE_RESTART_THRESH_SECS: 15
        # # Number of messages the nozzle buffer can hold while processing. Also the number of messages that will be dropped if the buffer fills. Recommended minimum is 6000.
        # NRF_FIREHOSE_DIODE_BUFFER: 8192
        # # Log level (INFO or DEBUG)
        # NRF_LOG_LEVEL: INFO
        # # Trace level logging (extremely verbose)
        # NRF_TRACER: false

        # # LogMessage filters (| separated values)
        # NRF_LOGMESSAGE_SOURCE_INCLUDE: ""
        # NRF_LOGMESSAGE_SOURCE_EXCLUDE: ""
        # NRF_LOGMESSAGE_MESSAGE_INCLUDE: ""
        # NRF_LOGMESSAGE_MESSAGE_EXCLUDE: ""

        # # if proxy used in your environment
        # http_proxy: <proxy server address:port>
        # no_proxy:  <comma separated list of servers to bypass proxy>


**Note:** In order to automate the `cf push` deployment process as much as possible, the project contains a Cloud Foundry [manifest][d] file. Update the manifest as required for your environment. Make sure to assign proper values to all required environment variables. Any property values within angle brackets need to be changed to the correct value for your environment.

**Note:** When you're pushing the nozzle as an app, the `product` and `release` folders are not required. Make sure to remove these folders from the directory where you run `cf push` to reduce the size of the upload, or use `.cfignore` file.



## **Import as a tile in Ops Manager**

Import the tile from [releases][f] to Ops Mgr. Once it's imported, install the tile, and follow the steps below to configure the tile.

When it's installed as a tile in Ops Mgr, click the firehose nozzle tile to access the setup, and enter the following properties in the tile settings:

Under **New Relic Firehose Nozzle tile -> Settings -> Assign AZs and Networks:**

    select your desired networks.

Under **New Relic Firehose Nozzle tile -> Settings -> New Relic Firehose Nozzle** set the following fields:

    New Relic RPM Account Id: <New Relic RPM Account>
    New Relic RPM Account Region: <US or EU. Choose EU if your RPM URL contains .eu>
    New Relic Insights Insert Key: <New Relic Insights Insert Key>
    Firehose Subscription Id: Unique Subscription Identifier (i.e. newrelic.firehose)
    Nozzle Instances: You could run 1 to 30 instances of the nozzle in any environment

Under **New Relic Firehose Nozzle tile -> Settings -> Advanced Settings** set the following fields:

    Proxy Server Address and Port: <proxy server address:port or leave blank>
    Proxy Bypass: <Comma separated list of servers to bypass proxy>
    Skip SSL Verification (True/false): Whether to verify SSL connection
    Selected Events: Comma-separated List of event types
    Drain Interval: How often aggregated metric type events should be sent.
    Firehose -> Reverse Log Proxy Gateway Timeout (minutes): Number of minutes before the HTTP connection to the RLP Gateway is considered hung and restarted. The RLP Gateway should force a new connection every 14 minutes.  
    Firehose No Traffic Restart Threshold (seconds): Number of consecutive seconds with no messages before the nozzle is automatically restarted. Set per environment based on normal message load.
    Firehose Queue Buffer Size (messages): Number of messages the nozzle buffer can hold while processing. Also the number of messages that will be dropped if the buffer fills. Recommended minimum is 6000.
    Log Level: Verbosity of log files (INFO or DEBUG)
    Enable Tracer (True/false): Whether to include trace level logging (extremely verbose)

Under **New Relic Firehose Nozzle tile -> Settings -> LogMessage Filters** set the following fields:

    LogMessage Source Include Filter: Only send PCFLogMessage events from sources in this list (| separated)
    LogMessage Source Exclude Filter: Ignore PCFLogMessage events from sources in this list (| separated)
    LogMessage Message Content Include Filter: Only send PCFLogMessage events if the message content contains the items in this list (| separated)
    LogMessage Message Content Exclude Filter: Ignore PCFLogMessage events if the message content contains the items in this list (| separated)


Once all this information is entered, go back to **Installation Dashboard**, and click the **Apply Changes** button on the top right.

## **Where to obtain configuration values**

The following properties can be obained either from **Ops Mgr Elastic Runtime** or from **Insights**:
<pre>
    * UAA Url: https://uaa.<your-pcf-domain>  --  "cf curl /v2/info"
    * API Url: https://api.<your-pcf-domain>  --  "cf curl /v2/info"
    * Firehose Subscription Id: A unique Id (i.e. newrelic.firehose)
    * Skip SSL: If SSL is disabled this is value should be set to "true"
    * Selected Events: A comma-separated list of any of the following firehose event types:
    	- ValueMetric
    	- CounterEvent
    	- ContainerMetric
    	- HttpStartStop
    	- LogMessage
    * New Relic Account Id: The first number that you find in your RPM Url (i.e. https://insights.newrelic.com/accounts/<rpm-id>/...)
    * Insights Insert Key: An "Insert Key" from https://insights.newrelic.com/accounts/<rpm-id>/manage/api_keys. In the UI you can go to "New Relic Insights -> Manage Data -> Api Keys" to create an "Insert Key" if one does not exist already, or if you'd like to create a fresh insert key specifically for this purpose.
    * API Username: "Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Admin Credentials -> Link to Credential -> identity"
    * API Password: "Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Admin Credentials -> Link to Credential -> password"
    * Client Id: "Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Opentsdb Nozzle Credentials -> Link to Credential -> identity"
    * Client Secret: "Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Opentsdb Nozzle Credentials -> Link to Credential -> password"
    * A new doppler.firehose enabled client can be created instead of retrieving the default client described above:
        ```
        uaac target https://uaa.[your cf system domain]
        uaac token client get admin -s [your admin-secret]
        uaac client add firehose-to-newrelic \
          --name firehose-to-newrelic \
          --secret [your_client_secret] \
          --authorized_grant_types client_credentials,refresh_token \
          --authorities doppler.firehose,cloud_controller.admin_read_only \
          --scope doppler.firehose
        ```
        `firehose-to-newrelic` is your `NRF_CF_CLIENT_ID` env variable
        the `--secret` you chose is your `NRF_CF_CLIENT_SECRET` env variable
</pre>



## **Sample Insights queries**

Multiple event types are used for the nozzle, each of which start with PCF. The following are some NRQL strings you can use to extract events and metrics.

```
SELECT count(*) FROM PCFCapacity, PCFContainerMetric, PCFCounterEvent, PCFHttpStartStop, PCFLogMessage, PCFValueMetric SINCE 1 day ago FACET pcf.envelope.type

SELECT count(*) FROM PCFValueMetric SINCE 1 day ago FACET pcf.job TIMESERIES

SELECT count(*) FROM PCFValueMetric WHERE pcf.job = 'diego_cell' SINCE 1 day ago FACET origin TIMESERIES

SELECT average(metric.sum/metric.samples.count) FROM PCFContainerMetric WHERE metric.name = 'app.cpu' FACET app.name TIMESERIES

SELECT count(*) from PCFHttpStartStop facet http.status
```

Events from all PCF deployments end up in **`PCFCapacity`, `PCFContainerMetric`, `PCFCounterEvent`, `PCFHttpStartStop`, `PCFLogMessage`, and `PCFValueMetric`**. If you collect events from multiple PCF environments, you can use **`pcf.domain`** and **`pcf.ip`** attributes to distunguish between events from different PCF deploments (either in a **`WHERE`** clause or by a **`FACET`** of the events by **`pcf.domain`**).

**Note:** Please contact New Relic to obtain the pre-built dashboards for the nozzle.


## **Insights dashboards**

Please contact your New Relic representative to import pre-built nozzle dashboards to your New Relic account.


## **Using proxy**

If you need to use a proxy server in your environment, please use the following two environment variables:

    * **`http_proxy`**
    * **`no_proxy`**

If you use the tile, during the setup of the tile in Ops Mgr you can specify values for these properties. If you use the app version of the nozzle (running by `cf push`) then uncomment the last two environment variables at the end of `manifest.yml` in the **`env`** section.

**Notes**   

    * These proxy environment variables must be in lower case.
    * You must set **`http_proxy`** to your proxy server address and port (i.e. http://my_proxyserver:my_proxy_port)
    * You must set **`no_proxy`** to any address that you need to bypass. In order for the nozzle to work with proxies, you must bypass the doppler server (i.e. `doppler.my_pcf_domain.com`). Make sure you do not include the protocol and the port to `no_proxy`, just add the server name.


## **Compatibility**

This project has been tested and is compatible with PCF **2.4** and higher.



## **Application build and deploy**

The application is already built and ready to run on PCF linux. If you make any changes to the code, or would like to run on other OSs, you can rebuild the binary.

The project uses `dep` to manage the dependencies. To pull the necessary packages into the vendor folder run: ```dep ensure```

<pre>
dep init
dep ensure
env GOOS=&lt;OS-name&gt; GOARCH=amd64 go build -o dist/nr-nozzle
cf push
</pre>



## **License**

The project is released under version 2.0 of the [Apache License][e].


[a]: https://docs.cloudfoundry.org/loggregator/architecture.html
[b]: https://pivotal.io/platform
[c]: http://newrelic.com/insights
[d]: manifest.yml
[e]: http://www.apache.org/licenses/LICENSE-2.0
[f]: https://github.com/newrelic/newrelic-pcf-nozzle-tile/releases

## **Build**

1. Clone the repo.
2. Run `dep ensure`.
3. Run `./release.sh`.

This creates a `nr-fh-nozzle` binary specific for the PCF Linux OS, in `./dist` along with a `tar.gz` for convenient shipping.

## **Local testing**

Requirements:
- access to a PCF environment
- PCF API credentials with admin rights
- PCF UAA credentials with following rights:

```bash
--authorized_grant_types client_credentials,refresh_token
--authorities doppler.firehose,cloud_controller.admin_read_only
```

1. Set your environment variables per the `manifest.yml` file located in the main directory.
2. Run: `go run main.go`

## Support

New Relic has open-sourced this project. This project is provided AS-IS WITHOUT WARRANTY OR DEDICATED SUPPORT. Issues and contributions should be reported to the project here on GitHub.

We encourage you to bring your experiences and questions to the [Explorers Hub](https://discuss.newrelic.com) where our community members collaborate on solutions and new ideas.

### Community

New Relic hosts and moderates an online forum where customers can interact with New Relic employees as well as other customers to get help and share best practices. Like all official New Relic open source projects, there's a related Community topic in the New Relic Explorers Hub. You can find this project's topic/threads here:

https://discuss.newrelic.com/t/new-relic-firehose-nozzle-for-pcf/92709

### Issues / Enhancement Requests

Issues and enhancement requests can be submitted in the [Issues tab of this repository](../../issues). Please search for and review the existing open issues before submitting a new issue.