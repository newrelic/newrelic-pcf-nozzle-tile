# **New Relic PCF Nozzle Tile**

This application is a Firehose nozzle which forwards metrics from the [PCF Loggregator][a] in [Pivotal Cloud Foundry][b] into [New Relic Insights][c] for visualization.

This code could either be pushed as a regular PCF application with **"cf push"**, or you could use the tile version of it and install it in Ops Mgr.



## **Push as an application**

When pushed as an application, you need to have a [manifest][d] with the following properties:

>	---
	applications:
	- name: newrelic-firehose-nozzle
	  memory: 512M
	  instances: 2
	  health-check-type: none
	  host: cf-firehose-nozzle-${random-word}
	  env:
	    NOZZLE_USERNAME: <nozzle user>
	    NOZZLE_PASSWORD: <nozzle password>
	    NOZZLE_UAA_URL: https://uaa.<your-pcf-domain>
	    NOZZLE_TRAFFIC_CONTROLLER_URL: wss://doppler.<your-pcf-domain>:<ssl-port>
	    NOZZLE_FIREHOSE_SUBSCRIPTION_ID: newrelic.firehose
	    NOZZLE_SKIP_SSL: true/false
	    NOZZLE_SELECTED_EVENTS: Comma-separated list of firehose event types (optional)
        NOZZLE_EXCLUDED_DEPLOYMENTS: Comma-separated list of deployments to exclude (optional)
        NOZZLE_EXCLUDED_ORIGINS: Comma-separated list of origins to exclude (optional)
        NOZZLE_EXCLUDED_JOBS: Comma-separated list of jobs to exclude
        NOZZLE_ADMIN_USER: <admin-user> with admin privileges to obtain all application details in all orgs/spaces
        NOZZLE_ADMIN_PASSWORD: <admin-password> password for the user with admin privileges
        NOZZLE_APP_DETAIL_INTERVAL: interval for querying application details (defaults to 1 minute)
	    NEWRELIC_INSIGHTS_BASE_URL: https://insights-collector.newrelic.com/v1
	    NEWRELIC_INSIGHTS_RPM_ID: <newrelic-rpm-account-id>
	    NEWRELIC_INSIGHTS_INSERT_KEY: <insights-insert-key>

        # http_proxy: <proxy server address:port>
        # no_proxy:  <comma separated list of servers to bypass proxy>


**Note:**	In order to automate the **"cf push"** deployment process as much as possible, the project contains a Cloud Foundry [manifest][d] file. Update the manifest as required for your environment. Make sure to assign proper values to all required environment variables. Any property values within angle brackets needs to be changed to the correct value for your environment.

**Note:**	When you're pushing the nozzle as an app, the **"product"** and **"release"** folders are not required. Make sure to remove these folders from the directory where you run **"cf push"** to reduce the size of the upload, or use **.cfignore** file.



## **Import as a tile in Ops Mgr**

Import the tile from [releases][f] to Ops Mgr. Once imported, install the tile and follow the steps below to configure the tile.

When installed as a tile in Ops Mgr, **"click on the firehose nozzle tile"** to access the setup, and enter the following properties in the tile settings:

Under **New Relic Firehose Nozzle tile -> Settings -> Assign AZs and Networks:**

    select your desired networks.

Under **New Relic Firehose Nozzle tile -> Settings -> New Relic Firehose Nozzle** set the following fields:

    New RelicInsights Base Url: https://insights-collector.newrelic.com/v1
    New Relic RPM Account Id: <New Relic RPM Account>
    New Relic Insights Insert Key: <New Relic Insights Insert Key>
    UAA Url: UAA Url of your PCF deployment
    Nozzle Instances: You could run 1 to 6 instances of the nozzle in any environment
    Skip SSL Verification (True/false): Whether to verify SSL connection
    UAA API User Account Name: User name for UAA
    UAA API User Account Password: Password for UAA
    Traffic Controller Url: Traffic Controller Url of your PCF deployment
    Firehose Subscription Id: Unique Subscription Identifier (i.e. newrelic.firehose)
    Selected Events: Comma-separated List of event types
    Excluded Deployments: Comma-separated list of deployments to exclude (optional)
    Excluded Origins: Comma-separated list of origins to exclude (optional)
    Excluded Jobs: Comma-separated list of jobs to exclude (optional)
    Admin User: <admin-user> with admin privileges to obtain all application details in all orgs/spaces
    Admin Password: <admin-password> password for the user with admin user
    App Detail Collection Interval: interval for querying application details (defaults to 1 minute)

    if proxy is used in your environment:
    http_proxy: <proxy server address:port>
    no_proxy: <comma separated list of servers to bypass proxy>


Once all this information is entered, go back to **"Installation Dashboard"** and click the big blue button on top right to **"Apply Changes"**.



## **Where to obtain Configuration Values**

Following properties can be obained either from Ops Mgr Elastic Runtime or from Insights:
<pre>
    * User Name: "Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Opentsdb Nozzle Credentials -> Link to Credential -> identity"
    * Password: "Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Opentsdb Nozzle Credentials -> Link to Credential -> password"
    * UAA Url: https://uaa.<your-pcf-domain>  --  "cf curl /v2/info"
    * Traffic Controller Url: wss://doppler.<pcf-domain>:<ssl-port>  --  "cf curl /v2/info"
    * Firehose Subscription Id: A unique Id (i.e. newrelic.firehose)
    * Skip SSL: If SSL is disabled this is value should be set to "true"
    * Selected Events: A comma-separated list of any of the following firehose event types:
    	- ValueMetric
    	- CounterEvent
    	- ContainerMetric
    	- HttpStartStop
    	- LogMessage
    * Insights Base Url: https://insights-collector.newrelic.com/<API-Version> (API version is currently v1)
    * Insights RPM Id: The first number that you find in your RPM Url (i.e. https://insights.newrelic.com/accounts/<rpm-id>/...)
    * Insights Insert Key: An "Insert Key" from https://insights.newrelic.com/accounts/<rpm-id>/manage/api_keys. In the UI you can go to "New Relic Insights -> Manage Data -> Api Keys" to create an "Insert Key" if one does not exist already, or if you'd like to create a fresh insert key specifically for this purpose.
    * Admin User Name: "Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Admin Credentials -> Link to Credential -> identity"
    * Admin Password: "Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Admin Credentials -> Link to Credential -> password"
</pre>



## **Sample Insights Queries**

The **"Insights Event type"** is called **"PcfFirehoseEvent"**. Following are some NRQL strings you could use to extract events and metrics.

```
select count(*) from PcfFirehoseEvent since 1 day ago facet FirehoseEventType

select count(*) from PcfFirehoseEvent since 1 day ago facet job timeseries

select count(*) from PcfFirehoseEvent where job = 'diego_cell' since 1 day ago facet origin timeseries

select average(containerMetricCpuPercentage) from PcfFirehoseEvent facet containerMetricApplicationId timeseries

select count(*) from  PcfFirehoseEvent where FirehoseEventType = 'HttpStartStop' facet httpStatusCode
```

Events from all PCF deployments end up in **"PcfFirehoseEvent"**. If you collect events from multiple PCF environments you can use **pcfDomain** and **pcfInstanceIp** metrics to distunguish between events from different PCF deploments (either in a **WHERE** clause or by **FACET**ing the events by **pcfDomain**).

**Note:**	Please contact New Relic to obtain the pre-built dashboards for the nozzle.


## **Insights Dashboards**

Please contact your New Relic representative to import pre-built nozzle dashboards to your New Relic account.


## **Using Proxy**

If you need to use proxy server in your environment, please use the following 2 environment variables:
    * **http_proxy**
    * **no_proxy**

If you use the tile, during the setup of the tile in Ops Mgr you can specify values for these properties. If you use the app version of the nozzle (running by cf push) then uncomment the last 2 environment variables at the end of manifest.yml in the **env** section.

**Notes**   
    * These proxy environment variables must be in lower case.
    * You need to set **http_proxy** to your proxy server address and port (i.e. http://my_proxyserver:my_proxy_port)
    * You need to set **no_proxy** to any address that you need to bypass. In order for the nozzle to work with proxies, you must bypass the doppler server (i.e. doppler.my_pcf_domain.com). Make sure you do not include the protocol and the port to no_proxy, just add the server name.


## **Compatibility**

This project has been tested and is compatible with PCF **1.8**, **1.9**, **1.10**, **1.11**, **1.12**, and **2.0**.



## **Application Build & Deploy**

The application is already built and ready to run on PCF linux. If you make any changes to the code, or would like to run on other OS's, you can rebuild the binary.

<pre>
env GOOS=&lt;OS-name&gt; GOARCH=amd64 go build -o nr-nozzle
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
