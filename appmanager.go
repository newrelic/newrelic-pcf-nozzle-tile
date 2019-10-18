package main

import (
	"encoding/json"
	"fmt"
	"time"
	"errors"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/go-redis/redis"
	"github.com/cloudfoundry-community/go-cfenv"
)

const (
	redisChannelName = "NewRelicNozzleApplicationData"
)

//AppInfo is application data looked up via pcf API
type AppInfo struct {
	Timestamp   int64
	Name        string
	Guid        string
	CreatedTime string
	LastUpdated string
	Instances   int
	StackGUID   string
	State       string
	Diego       bool
	SshEnabled  bool
	SpaceName   string
	SpaceGUID   string
	OrgName     string
	OrgGUID     string
}

type RedisDbOptions struct {
	RedisDb                       string `json:"value"`
	SELECTED_OPTION  struct {
		RedisServiceName          string `json:"nozzle_redis_service_name"`
		RedisServicePlan          string `json:"nozzle_redis_service_plan"`
	}
}

//AppManager manages the application details that are looked up via pcf api
type AppManager struct {
	appData           map[string]*AppInfo
	readChannel       chan readRequest
	closeChannel      chan bool
	updateChannel     chan map[string]*AppInfo
	client            *cfclient.Client
	appUpdateInterval int
	pcfExtendedConfig *PcfExtConfig
	lastUpdate        time.Time

	redisClient       *redis.Client
	redisDbEnabled    bool
	redisServiceName  string
	redisServicePlan  string

	redisHost         string
	redisPort         string
	redisPassword     string
}

type readRequest struct {
	appGUID      string
	responseChan chan AppInfo
}

//NewAppManager create and initialize an AppManager
func NewAppManager(cfClient *cfclient.Client, updateInterval int, config *PcfExtConfig) *AppManager {
	instance := &AppManager{}
	instance.client = cfClient
	instance.appUpdateInterval = updateInterval
	instance.appData = make(map[string]*AppInfo, 0)
	instance.readChannel = make(chan readRequest)
	instance.closeChannel = make(chan bool)
	instance.updateChannel = make(chan map[string]*AppInfo)
	instance.pcfExtendedConfig = config
	instance.lastUpdate = time.Now()
	instance.parseRedisDbOptions()
	if (instance.redisDbEnabled) {
		instance.initRedisClient()
	}
	return instance
}

func (am *AppManager) parseRedisDbOptions() {
	if (am.pcfExtendedConfig.REDIS_DB != "") {
		var options *RedisDbOptions
		err := json.Unmarshal([]byte(am.pcfExtendedConfig.REDIS_DB), &options)
		if err != nil {
			logger.Printf("Failed to unmarshal app json from redis: %s\n", err.Error())
		}
		am.redisDbEnabled = (options.RedisDb == "Enable")
		am.redisServiceName = options.SELECTED_OPTION.RedisServiceName
		am.redisServicePlan = options.SELECTED_OPTION.RedisServicePlan
	} else {
		am.redisDbEnabled = false
	}
}

func (am *AppManager) initRedisClient() {
	logger.Printf("Creating Redis Client...")
	am.getDbCredentialsFromVcapServices()
	addr := fmt.Sprintf("%s:%s", am.redisHost, am.redisPort)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: am.redisPassword,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		logger.Printf("Redis Client-pong failed: %s\n", err.Error())
		logger.Println("Could not connect to Redis. Ignoring database.")
		am.redisDbEnabled = false
		return
	}
	logger.Printf("pong result: %s\n", pong)
	am.redisClient = client
	if nozzleInstanceId != "0" {
		am.subscribeToAppUpdates()
	}
}

func (am *AppManager) getDbCredentialsFromVcapServices() {
	// parse VCAP_SERVICES to find redis host, port, pw
	appEnv, err := cfenv.Current()
	if (err != nil) {
		// error - "unable to get app environment"
		panic(err)
	}
	redisService, err := appEnv.Services.WithLabel(am.redisServiceName)
	if (err != nil) { // this can happen if the nozzle was pushed manually and redis DB not installed in PCF foundation
		logger.Printf("\tERROR! You have selected to use redis database service %s, but it is not available", 
			am.redisServiceName)
		logger.Println("\tInstall redis on this PCF foundation, and restart the nozzle")
		panic(err)
	}

	redisServiceFound := false
	for rs := range redisService { // find the selected plan if for any weird reason there are multiple service plans

		if (redisService[rs].Plan == am.redisServicePlan) {
	    	redisServiceFound = true
		    am.redisHost = fmt.Sprintf("%v", redisService[rs].Credentials["host"])
		    am.redisPort = fmt.Sprintf("%v", redisService[rs].Credentials["port"])
		    am.redisPassword = fmt.Sprintf("%v", redisService[rs].Credentials["password"])
		    break
	    }
    }
    if (!redisServiceFound) { // this can happen if the nozzle was pushed manually and redis instance not created
		logger.Printf("\tERROR! You have selected to use redis database, but service instance for plan %s is not available\n", 
			am.redisServicePlan)
		logger.Println("\tCreate the redis service instance and restart the nozzle")
    	panic(errors.New("ERROR: specified service instance not available!"))
    }
}

//Start starts the app manager
//periodically updates application data and provides
//synchronized accesas to application data
func (am *AppManager) Start() {
	logger.Printf("Starting Goroutine to refresh applications data every %d minute(s)\n", am.appUpdateInterval)
	//get the data as soon as possible
	am.refreshAppData()
	ticker := time.NewTicker(time.Duration(int64(am.appUpdateInterval)) * time.Minute)

	go func() {
		if am.redisClient != nil {
			defer am.redisClient.Close()
		}
		for {
			select {
			case <-ticker.C:
				am.refreshAppData()

			case tempAppInfo := <-am.updateChannel:
				logger.Printf("App Update....received %d app details", len(tempAppInfo))
				am.appData = tempAppInfo
				am.lastUpdate = time.Now()

			case rr := <-am.readChannel:
				ad := am.getAppData(rr.appGUID)
				rr.responseChan <- ad

			case <-am.closeChannel:
				logger.Print("quit \r\n")
				ticker.Stop()
			}
		}
	}()
}

func (am *AppManager) subscribeToAppUpdates() {
	pubsub := am.redisClient.Subscribe(redisChannelName)
	appCh := pubsub.Channel()
	go func() {
		for appMsg := range appCh {
			var tempAppMap map[string]*AppInfo
			err := json.Unmarshal([]byte(appMsg.Payload), &tempAppMap)
			if err != nil {
				logger.Printf("Failed to unmarshal app json from redis: %s\n", err.Error())
			}
			am.updateChannel <- tempAppMap
		}
	}()
}

func (am *AppManager) refreshAppData() {
	if am.redisClient != nil {
		if nozzleInstanceId == "0" {
			go am.getAndPublishAppData()
			return
		}
		if time.Now().Sub(am.lastUpdate) > time.Duration(int64(am.appUpdateInterval))*time.Minute*3 {
			logger.Println("Failed to receive application updates from redis for 3 intervals. Calling CF myself")
		} else {
			//we are getting updates from subscribe so do nothing
			return
		}
	}
	go am.getAppDataFromCf()
}

func (am *AppManager) getAndPublishAppData() {
	tempAppInfo := am.getAppDataFromCf()
	if tempAppInfo == nil {
		//error calling cf API-we don't have any data to publish
		return
	}
	aJSON, err := json.Marshal(tempAppInfo)
	if err != nil {
		logger.Printf("Error marshalling app data: %s", err.Error())
		return
	}
	//logger.Printf("app json: %s\n", string(aJSON))
	err = am.redisClient.Publish(redisChannelName, aJSON).Err()
	if err != nil {
		logger.Printf("Error publishing application data: %s\n", err.Error())
	}
}

func (am *AppManager) getAppDataFromCf() map[string]*AppInfo {
	apps, err := client.ListApps()
	if err != nil {
		// error in cf-clinet library -- failed to get updated applist - will try next cycle
		logger.Printf("Warning: cf-client ListApps failed - will try again in %d minute(s). Error: %s\n",
			am.appUpdateInterval, err.Error())
		return nil
	}
	tempAppInfo := map[string]*AppInfo{}
	for _, app := range apps {

		tempAppInfo[app.Guid] = &AppInfo{
			time.Now().UnixNano() / 1000000,
			app.Name,
			app.Guid,
			app.CreatedAt,
			app.UpdatedAt,
			app.Instances,
			app.StackGuid,
			app.State,
			app.Diego,
			app.EnableSSH,
			app.SpaceData.Entity.Name,
			app.SpaceData.Entity.Guid,
			app.SpaceData.Entity.OrgData.Entity.Name,
			app.SpaceData.Entity.OrgData.Entity.Guid,
		}
	}
	am.updateChannel <- tempAppInfo
	return tempAppInfo
}

//GetAppData will look in the cache for the appGuid
func (am *AppManager) GetAppData(appGUID string) AppInfo {
	//logger.Printf("Searching for %s\n", appGUID)
	req := readRequest{appGUID, make(chan AppInfo)}
	am.readChannel <- req
	ai := <-req.responseChan
	//logger.Printf("Recevied response for %s: %+v", appGUID, ai)
	return ai
}

func (am *AppManager) getAppData(appGUID string) AppInfo {
	//logger.Printf("\tSearching for %s in map with %d items\n", appGUID, len(am.appData))
	if ai, found := am.appData[appGUID]; found {
		//logger.Printf("\tFound %s: %+v\n", appGUID, ai)
		return *ai
	}
	//logger.Printf("\tCouldn't find %s\n", appGUID)
	ai := &AppInfo{}
	ai.Name = "awaiting update"
	return *ai
}

//IsEmpty checks whether the struct is initialized with data
func (ai *AppInfo) IsEmpty() bool {
	if ai.Timestamp == 0 {
		return true
	}
	return false
}
