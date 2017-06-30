# **New Relic PCF Nozzle Tile**

This application is a nozzle which forwards metrics from the [PCF Loggregator][a] in [Pivotal Cloud Foundry][b] into [New Relic Insights][c] for visualization.

The application could either be pushed as a regular PCF application with **"cf push"**, or you could use the tile version of it and install it in Ops Mgr.



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
	    NOZZLE_API_URL: https://api.<your-pcf-domain>
	    NOZZLE_UAA_URL: https://uaa.<your-pcf-domain>
	    NOZZLE_TRAFFIC_CONTROLLER_URL: wss://doppler.<your-pcf-domain>:<ssl-port>
	    NOZZLE_FIREHOSE_SUBSCRIPTION_ID: newrelic.firehose
	    NOZZLE_SKIP_SSL: true/false
	    NOZZLE_SELECTED_EVENTS: Comma-separated list of event types
	    NEWRELIC_INSIGHTS_BASE_URL: https://insights-collector.newrelic.com/v1
	    NEWRELIC_INSIGHTS_RPM_ID: <newrelic-rpm-account-id>
	    NEWRELIC_INSIGHTS_INSERT_KEY: <insights-insert-key>

**Note:**	In order to automate the **"cf push"** deployment process as much as possible, the project contains a Cloud Foundry [manifest][d] file. Update the manifest as required for your environment. Make sure to assign proper values to all required environment variables. Any property values within angle brackets need to be changed to correct values for your environment.

**Note:**	When you're pushing the nozzle as an app, the **"product"** and **"release"** folders are not required. Make sure to remove these folders from the directory where you run **"cf push"** to reduce the size of the upload.



## **Import as a tile in Ops Mgr**

Import the tile from **"releases"** folder (i.e. **"releases/nr-firehose-nozzle-0.0.2.pivotal"**) to Ops Mgr. Once imported, install the tile and follow the steps below to configure the tile.

When installed as a tile in Ops Mgr, you need to setup the following properties in the tile settings:

Under **New Relic Firehose Nozzle tile -> Settings -> Assign AZs and Networks:**

    set Network to "ert"

Under **New Relic Firehose Nozzle tile -> Settings -> New Relic Firehose Nozzle** set the following fields:

    New RelicInsights Base Url: https://insights-collector.newrelic.com/v1
    New Relic RPM Account Id: <New Relic RPM Account>
    New Relic Insights Insert Key: <New Relic Insights Insert Key>
    UAA Url: UAA Url of your PCF deployment
    Skip SSL Verification (True/false): Whether to verify SSL connection
    UAA API User Account Name: User name for UAA
    UAA API User Account Password: Password for UAA
    Traffic Controller Url: Traffic Controller Url of your PCF deployment
    Firehose Subscription Id: Unique Subscription Identifier (i.e. newrelic.firehose)
    Selected Events: Comma-separated List of event types


Once all this information is entered, go back to **"Installation Dashboard"** and click the big blue button on top right to **"Apply Changes"**.



## **Where to obtain Configuration Values**

Following properties can be obained either from Ops Mgr Elastic Runtime or from Insights:
<pre>
    * User Name: Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Opentsdb Nozzle Credentials -> Link to Credential -> identity
    * Password: Ops Mgr -> Elastic Runtime -> Credentials -> Job -> UAA -> Opentsdb Nozzle Credentials -> Link to Credential -> password
    * UAA Url: https://uaa.<your-pcf-domain>
    * Traffic Controller Url: wss://doppler.<pcf-domain>:<ssl-port>
    * Firehose Subscription Id: A unique Id (i.e. newrelic.firehose)
    * Skip SSL: If SSL is disabled this is value should be set to "true"
    * Selected Events: A comma-separated list of any of the following event types:
    	- ValueMetric
    	- CounterEvent
    	- ContainerMetric
    	- HttpStopStart
    	- LogMessage
    * Insights Base Url: https://insights-collector.newrelic.com/<API-Version> (API version is currently v1)
    * Insights RPM Id: The first number that you find in your RPM Url (i.e. https://insights.newrelic.com/accounts/<rpm-id>/...)
    * Insights Insert Key: An "Insert Key" from https://insights.newrelic.com/accounts/<rpm-id>/manage/api_keys. You may need to create an "Isert Key" if one does not exist already, or is being used for other purposes.
</pre>



## **Sample Insights Queries**

The **"Insights Event type"** is called **"PcfFirehoseEvent"**. Following are some NRQL strings you could use to extract events and metrics.

```
select count(*) from PcfFirehoseEvent since 1 day ago facet FirehoseEventType

select count(*) from PcfFirehoseEvent since 1 day ago facet job timeseries

select count(*) from PcfFirehoseEvent where job = 'diego_cell' since 1 day ago facet origin  timeseries

select average(containerMetricCpuPercentage) from PcfFirehoseEvent facet containerMetricApplicationId timeseries

select count(*) from  PcfFirehoseEvent where FirehoseEventType = 'HttpStartStop' facet httpStatusCode
```

Events from all PCF deployments end up in **"PcfFirehoseEvent"**. If you collect events from multiple PCF environments you can use **pcfApiUrl** to distunguish between events from different PCF deploment (either in a **WHERE** clause or by **FACET**ing the events by **pcfApiUrl**).

**Note:**	Please contact your New Relic to obtain the pre-built dashboards for the nozzle.


## **Compatibility**

This project has been tested and is compatible with PCF **1.8**, **1.9**, and **1.10**.



## **Application Build & Deploy**

The application is already built and ready to run on PCF linux. If you make any changes to the code, or would like to run on other OS's, you can rebuild the binary.

```
env GOOS=linux GOARCH=amd64 go build -o nr-nozzle
cf push
```



## **License**

The project is released under version 2.0 of the [Apache License][e].






[a]: https://docs.cloudfoundry.org/loggregator/architecture.html
[b]: https://pivotal.io/platform
[c]: http://newrelic.com/insights
[d]: manifest.yml
[e]: http://www.apache.org/licenses/LICENSE-2.0

