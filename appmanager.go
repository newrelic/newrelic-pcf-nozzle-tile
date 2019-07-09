package main

import (
	"time"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

//AppInfo is application data looked up via pcf API
type AppInfo struct {
	timestamp   int64
	name        string
	guid        string
	createdTime string
	lastUpdated string
	instances   int
	stackGUID   string
	state       string
	diego       bool
	sshEnabled  bool
	spaceName   string
	spaceGUID   string
	orgName     string
	orgGUID     string
}

//AppManager manages the application details that are looked up via pcf api
type AppManager struct {
	appData           map[string]*AppInfo
	readChannel       chan readRequest
	closeChannel      chan bool
	updateChannel     chan map[string]*AppInfo
	client            *cfclient.Client
	appUpdateInterval int
}

type readRequest struct {
	appGUID      string
	responseChan chan AppInfo
}

//NewAppManager create and initialize an AppManager
func NewAppManager(cfClient *cfclient.Client, updateInterval int) *AppManager {
	instance := &AppManager{}
	instance.client = cfClient
	instance.appUpdateInterval = updateInterval
	instance.appData = make(map[string]*AppInfo, 0)
	instance.readChannel = make(chan readRequest)
	instance.closeChannel = make(chan bool)
	instance.updateChannel = make(chan map[string]*AppInfo)
	return instance
}

//Start starts the app manager
//periodically updates application data and provides
//synchronized accesas to application data
func (am *AppManager) Start() {
	logger.Printf("Starting Goroutine to refresh applications data every %d minute(s)\n", am.appUpdateInterval)
	//get the data as soon as possible
	go am.refreshAppData()
	ticker := time.NewTicker(time.Duration(int64(am.appUpdateInterval)) * time.Minute)

	go func() {
		for {
			select {
			case <-ticker.C:
				go am.refreshAppData()

			case tempAppInfo := <-am.updateChannel:
				logger.Printf("App Update....received %d app details", len(tempAppInfo))
				am.appData = tempAppInfo

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

func (am *AppManager) refreshAppData() {
	logger.Println("Refreshing application data...")
	apps, err := client.ListApps()
	if err != nil {
		// error in cf-clinet library -- failed to get updated applist - will try next cycle
		logger.Printf("Warning: cf-client failed to return applications list - will try again in %d minute(s)...\n", am.appUpdateInterval)
	} else {
		eventCount := len(apps)
		logger.Printf("App Count: %3d\n", eventCount)

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
		//write updated app data to channel to avoid ReplaceAllString
		am.updateChannel <- tempAppInfo
	}
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
	ai.name = "awaiting update"
	return *ai
}

//IsEmpty checks whether the struct is initialized with data
func (ai *AppInfo) IsEmpty() bool {
	if ai.timestamp == 0 {
		return true
	}
	return false
}
