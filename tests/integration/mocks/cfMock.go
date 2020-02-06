package mocks

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
)

type MockCF struct {
	Server      *httptest.Server
	tokenString string
}

func NewMockCF(tokenType string, accessToken string) *MockCF {
	return &MockCF{
		tokenString: fmt.Sprintf(`{"token_type": "%s","access_token": "%s"}`, tokenType, accessToken),
	}
}

func (mCF *MockCF) Start() {
	mCF.Server = httptest.NewUnstartedServer(mCF)
	mCF.Server.Start()
}

func (mCF *MockCF) Stop() {
	mCF.Server.Close()
}

func (mCF *MockCF) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	log.Printf("API visited: %s", r.RequestURI)

	switch r.URL.Path {
	case "/v3/organizations":
		rw.Write([]byte(fmt.Sprintf(`{
		   "pagination": {
			  "total_results": 2,
			  "total_pages": 1,
			  "first": {
				 "href": "https://api.dev.cfdev.sh/v3/organizations?page=1&per_page=50"
			  },
			  "last": {
				 "href": "https://api.dev.cfdev.sh/v3/organizations?page=1&per_page=50"
			  },
			  "next": null,
			  "previous": null
		   },
		   "resources": [
			  {
				 "guid": "8614896d-6c98-4b59-9a19-2f6f15ae5fec",
				 "created_at": "2020-01-29T11:37:57Z",
				 "updated_at": "2020-01-29T11:46:22Z",
				 "name": "system",
				 "links": {
					"self": {
					   "href": "https://api.dev.cfdev.sh/v3/organizations/8614896d-6c98-4b59-9a19-2f6f15ae5fec"
					}
				 }
			  },
			  {
				 "guid": "d782be8d-add2-4f8b-82ea-a91ae60875e0",
				 "created_at": "2020-01-29T11:46:11Z",
				 "updated_at": "2020-01-29T11:46:11Z",
				 "name": "cfdev-org",
				 "links": {
					"self": {
					   "href": "https://api.dev.cfdev.sh/v3/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0"
					}
				 }
			  }
		   ]
			}`)))
	case "/oauth/token":
		rw.Write([]byte(mCF.tokenString))
	case "/v2/read":
		rw.Write([]byte(fmt.Sprintf(`
			{
			   "description": "Unknown request",
			   "error_code": "CF-NotFound",
			   "code": 10000
			}`)))
	case "/v2/info":
		rw.Write([]byte(fmt.Sprintf(`
		{
			"name": "Small Footprint PAS",
			"build": "2.4.7-build.16",
			"support": "https://support.pivotal.io",
			"version": 0,
			"description": "https://docs.pivotal.io/pivotalcf/2-3/pcf-release-notes/runtime-rn.html",
			"authorization_endpoint": "%s",
			"token_endpoint": "%s",
			"min_cli_version": "6.23.0",
			"min_recommended_cli_version": "6.23.0",
			"api_version": "2.125.0",
			"osbapi_version": "2.14",
			"routing_endpoint": "%s/routing"
		  }
		`, mCF.Server.URL, mCF.Server.URL, mCF.Server.URL)))
	case "/v2/apps":
		rw.Write([]byte(fmt.Sprintf(`
	{
   "total_results": 3,
   "total_pages": 1,
   "prev_url": null,
   "next_url": null,
   "resources": [
      {
         "metadata": {
            "guid": "6fb688fd-4eca-4e91-bc42-0beee1f40eb2",
            "url": "/v2/apps/6fb688fd-4eca-4e91-bc42-0beee1f40eb2",
            "created_at": "2020-01-29T11:46:26Z",
            "updated_at": "2020-01-29T11:48:16Z"
         },
         "entity": {
            "name": "p-invitations-green",
            "production": false,
            "space_guid": "e218f479-d4bc-4b17-ae88-e319758eca48",
            "stack_guid": "f9b82bf4-2344-4fb5-bff0-873452c815b9",
            "buildpack": null,
            "detected_buildpack": "nodejs",
            "detected_buildpack_guid": "a9651b4b-58cf-4cac-b1c4-9fa52cd042d2",
            "environment_json": {
               "CLOUD_CONTROLLER_URL": "https://api.dev.cfdev.sh",
               "COMPANY_NAME": "Pivotal",
               "INVITATIONS_CLIENT_ID": "invitations",
               "INVITATIONS_CLIENT_SECRET": "jGQmGMSDCf-A2VR_iC2fVDihVhwxzn7V",
               "NODE_TLS_REJECT_UNAUTHORIZED": "0",
               "NOTIFICATIONS_URL": "https://notifications.dev.cfdev.sh",
               "PRODUCT_NAME": "Apps Manager",
               "SUCCESS_CALLBACK_URL": "",
               "UAA_URL": "https://login.dev.cfdev.sh"
            },
            "memory": 256,
            "instances": 0,
            "disk_quota": 1024,
            "state": "STARTED",
            "version": "a35270c5-ddeb-486d-b68b-f8ab1b13cb3f",
            "command": null,
            "console": false,
            "debug": null,
            "staging_task_id": "a6e6d193-00dc-41c6-8b0d-06cc5865c6b1",
            "package_state": "STAGED",
            "health_check_type": "port",
            "health_check_timeout": null,
            "health_check_http_endpoint": null,
            "staging_failed_reason": null,
            "staging_failed_description": null,
            "diego": true,
            "docker_image": null,
            "docker_credentials": {
               "username": null,
               "password": null
            },
            "package_updated_at": "2020-01-29T11:46:27Z",
            "detected_start_command": "npm start",
            "enable_ssh": true,
            "ports": [
               8080
            ],
            "space_url": "/v2/spaces/e218f479-d4bc-4b17-ae88-e319758eca48",
            "stack_url": "/v2/stacks/f9b82bf4-2344-4fb5-bff0-873452c815b9",
            "routes_url": "/v2/apps/6fb688fd-4eca-4e91-bc42-0beee1f40eb2/routes",
            "events_url": "/v2/apps/6fb688fd-4eca-4e91-bc42-0beee1f40eb2/events",
            "service_bindings_url": "/v2/apps/6fb688fd-4eca-4e91-bc42-0beee1f40eb2/service_bindings",
            "route_mappings_url": "/v2/apps/6fb688fd-4eca-4e91-bc42-0beee1f40eb2/route_mappings"
         }
      },
      {
         "metadata": {
            "guid": "323dca53-b9a8-41e4-b8c1-15d48ba1f281",
            "url": "/v2/apps/323dca53-b9a8-41e4-b8c1-15d48ba1f281",
            "created_at": "2020-01-29T11:46:27Z",
            "updated_at": "2020-01-29T11:48:03Z"
         },
         "entity": {
            "name": "apps-manager-js-green",
            "production": false,
            "space_guid": "e218f479-d4bc-4b17-ae88-e319758eca48",
            "stack_guid": "f9b82bf4-2344-4fb5-bff0-873452c815b9",
            "buildpack": "staticfile_buildpack",
            "detected_buildpack": "staticfile",
            "detected_buildpack_guid": "f26cc511-6631-4791-ad2e-08958d140275",
            "environment_json": {
               "ACCENT_COLOR": "#00A79D",
               "ACCOUNT_URL": "https://login.dev.cfdev.sh/profile",
               "AMJS_ENV_VAR_KEYS": "ACCENT_COLOR ACCOUNT_URL APP_POLL_INTERVAL APPS_DOMAIN APPS_MANAGER_UAA_CLIENT_ID APPS_MANAGER_URL BILLING_ENABLED CLOUD_CONTROLLER_URL COMPANY_NAME CREATE_UPS_ENABLED CURRENCY_LOOKUP DISABLE_HTTP_FOR_LINKS DISPLAY_PLAN_PRICES DOCS_BASE_URL ENABLE_INVITING_USERS FEEDBACK_SERVICE_URL FOOTER_LINKS FOOTER_TEXT GLOBAL_WRAPPER_BG_COLOR GLOBAL_WRAPPER_FOOTER_CONTENT GLOBAL_WRAPPER_HEADER_CONTENT GLOBAL_WRAPPER_TEXT_COLOR HEAP_ANALYTICS_PROJECT_ID HIDE_APP_SEARCH HIDE_BETA_FEATURES INVITATIONS_SERVICE_ENABLED INVITATIONS_SERVICE_URL LOGGREGATOR_URL LOGGREGATOR_WEBSOCKET_URL LOGOUT_URL MAJOR_VERSION MARKETPLACE_NAME METRICS_URL MINOR_VERSION POLL_INTERVAL PRODUCT_NAME PUSH_APP_DOCS_URL SERVICE_PARAMS_DOCS_URL SIDEBAR_LINKS SYSTEM_DOMAIN UAA_URL USAGE_SERVICE_URL",
               "APPS_DOMAIN": "dev.cfdev.sh",
               "APPS_MANAGER_UAA_CLIENT_ID": "apps_manager_js",
               "APPS_MANAGER_URL": "https://apps.dev.cfdev.sh",
               "APP_POLL_INTERVAL": "10",
               "BILLING_ENABLED": "false",
               "CLOUD_CONTROLLER_URL": "https://api.dev.cfdev.sh",
               "COMPANY_NAME": "Pivotal",
               "CREATE_UPS_ENABLED": "true",
               "CURRENCY_LOOKUP": "{ \"usd\": \"$\", \"eur\": \"€\" }",
               "DISABLE_HTTP_FOR_LINKS": "false",
               "DISPLAY_PLAN_PRICES": "false",
               "DOCS_BASE_URL": "https://docs.pivotal.io/pivotalcf",
               "ENABLE_INVITING_USERS": "false",
               "FEEDBACK_SERVICE_URL": "",
               "FOOTER_LINKS": "[]",
               "FOOTER_TEXT": "©2017 Pivotal Software, Inc. All Rights Reserved.",
               "GLOBAL_WRAPPER_BG_COLOR": "#D6D6D6",
               "GLOBAL_WRAPPER_FOOTER_CONTENT": "",
               "GLOBAL_WRAPPER_HEADER_CONTENT": "",
               "GLOBAL_WRAPPER_TEXT_COLOR": "#333",
               "HEAP_ANALYTICS_PROJECT_ID": "",
               "HIDE_APP_SEARCH": "",
               "HIDE_BETA_FEATURES": "",
               "INVITATIONS_SERVICE_ENABLED": "true",
               "INVITATIONS_SERVICE_URL": "https://p-invitations.dev.cfdev.sh",
               "LOGGREGATOR_URL": "https://doppler.dev.cfdev.sh",
               "LOGGREGATOR_WEBSOCKET_URL": "wss://doppler.dev.cfdev.sh:443",
               "LOGOUT_URL": "https://login.dev.cfdev.sh/logout.do",
               "MAJOR_VERSION": "2",
               "MARKETPLACE_NAME": "Marketplace",
               "METRICS_URL": "https://metrics.dev.cfdev.sh",
               "MINOR_VERSION": "2",
               "POLL_INTERVAL": "30",
               "PRODUCT_NAME": "Apps Manager",
               "PUSH_APP_DOCS_URL": "http://docs.pivotal.io/pivotalcf/devguide/deploy-apps/deploy-app.html#push",
               "SERVICE_PARAMS_DOCS_URL": "http://docs.pivotal.io/pivotalcf/devguide/services/managing-services.html#arbitrary-params-create",
               "SIDEBAR_LINKS": "[{\"guid\":\"28b26833-a67f-4d5f-aaf3-1b4966606fd9\",\"href\":\"/marketplace\",\"name\":\"Marketplace\"},{\"guid\":\"e9962811-3c3d-4340-9e8e-b0643c30e6b5\",\"href\":\"https://docs.pivotal.io/pivotalcf/2-4/pas/intro.html\",\"name\":\"Docs\"},{\"guid\":\"3e1645a5-b5e4-4455-94e7-22bd0583656b\",\"href\":\"/tools\",\"name\":\"Tools\"}]",
               "SYSTEM_DOMAIN": "dev.cfdev.sh",
               "UAA_URL": "https://login.dev.cfdev.sh",
               "USAGE_SERVICE_URL": "https://app-usage.dev.cfdev.sh"
            },
            "memory": 128,
            "instances": 1,
            "disk_quota": 1024,
            "state": "STARTED",
            "version": "44d34fd3-68bc-4a75-bffd-7dc5875f7479",
            "command": "ruby -e 'require \"json\"; env = {}; ENV[\"AMJS_ENV_VAR_KEYS\"].split(\" \").each do |k|; if ENV.key?(k); env[k] = ENV[k].gsub(\"\\r\", \"\\\\r\").gsub(\"\\n\", \"\\\\n\"); end; end; File.open(\"public/config.js\", \"w\") do |file|; file.write(\"window.AMJS_ENV = \" + JSON.generate(env)); end;' && $HOME/boot.sh",
            "console": false,
            "debug": null,
            "staging_task_id": "a0f09aa9-c9b9-45ba-b20a-c7339bfe1fb6",
            "package_state": "STAGED",
            "health_check_type": "port",
            "health_check_timeout": null,
            "health_check_http_endpoint": null,
            "staging_failed_reason": null,
            "staging_failed_description": null,
            "diego": true,
            "docker_image": null,
            "docker_credentials": {
               "username": null,
               "password": null
            },
            "package_updated_at": "2020-01-29T11:46:27Z",
            "detected_start_command": "$HOME/boot.sh",
            "enable_ssh": true,
            "ports": [
               8080
            ],
            "space_url": "/v2/spaces/e218f479-d4bc-4b17-ae88-e319758eca48",
            "stack_url": "/v2/stacks/f9b82bf4-2344-4fb5-bff0-873452c815b9",
            "routes_url": "/v2/apps/323dca53-b9a8-41e4-b8c1-15d48ba1f281/routes",
            "events_url": "/v2/apps/323dca53-b9a8-41e4-b8c1-15d48ba1f281/events",
            "service_bindings_url": "/v2/apps/323dca53-b9a8-41e4-b8c1-15d48ba1f281/service_bindings",
            "route_mappings_url": "/v2/apps/323dca53-b9a8-41e4-b8c1-15d48ba1f281/route_mappings"
         }
      },
      {
         "metadata": {
            "guid": "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
            "url": "/v2/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
            "created_at": "2020-01-31T02:57:32Z",
            "updated_at": "2020-01-31T02:57:44Z"
         },
         "entity": {
            "name": "newrelic-firehose-nozzle",
            "production": false,
            "space_guid": "4b73ec0b-2a4b-49bb-9909-174043238763",
            "stack_guid": "f9b82bf4-2344-4fb5-bff0-873452c815b9",
            "buildpack": "binary_buildpack",
            "detected_buildpack": "binary",
            "detected_buildpack_guid": "60834ee0-2690-46f4-ba24-aac2dc1c6604",
            "environment_json": {
               "NRF_CF_API_PASSWORD": "XXXXXXXXXXXX",
               "NRF_CF_API_UAA_URL": "https://uaa.YOUR-PCF-DOMAIN",
               "NRF_CF_API_URL": "https://api.YOUR-PCF-DOMAIN",
               "NRF_CF_API_USERNAME": "admin",
               "NRF_CF_CLIENT_ID": "firehose-to-newrelic",
               "NRF_CF_CLIENT_SECRET": "XXXXXXXXXXX",
               "NRF_CF_SKIP_SSL": "true",
               "NRF_FIREHOSE_ID": "newrelic.firehose",
               "NRF_NEWRELIC_ACCOUNT_ID": "",
               "NRF_NEWRELIC_ACCOUNT_REGION": "US",
               "NRF_NEWRELIC_INSERT_KEY": ""
            },
            "memory": 512,
            "instances": 2,
            "disk_quota": 256,
            "state": "STARTED",
            "version": "ec7e3437-e1ca-4db0-8c4e-9bda11a4efd5",
            "command": "./nr-fh-nozzle",
            "console": false,
            "debug": null,
            "staging_task_id": "3a9816f8-f406-4a55-8da5-7f76843e3bde",
            "package_state": "STAGED",
            "health_check_type": "http",
            "health_check_timeout": null,
            "health_check_http_endpoint": "/health",
            "staging_failed_reason": null,
            "staging_failed_description": null,
            "diego": true,
            "docker_image": null,
            "docker_credentials": {
               "username": null,
               "password": null
            },
            "package_updated_at": "2020-01-31T02:57:34Z",
            "detected_start_command": ">&2 echo Error: no start command specified during staging or launch && exit 1",
            "enable_ssh": true,
            "ports": [
               8080
            ],
            "space_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763",
            "stack_url": "/v2/stacks/f9b82bf4-2344-4fb5-bff0-873452c815b9",
            "routes_url": "/v2/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed/routes",
            "events_url": "/v2/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed/events",
            "service_bindings_url": "/v2/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed/service_bindings",
            "route_mappings_url": "/v2/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed/route_mappings"
         }
      }
   ]
}
	`)))
	case "/v2/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed":
		rw.Write([]byte(fmt.Sprintf(`
			{
   "metadata": {
      "guid": "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
      "url": "/v2/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
      "created_at": "2020-01-31T02:57:32Z",
      "updated_at": "2020-01-31T02:57:44Z"
   },
   "entity": {
      "name": "newrelic-firehose-nozzle",
      "production": false,
      "space_guid": "4b73ec0b-2a4b-49bb-9909-174043238763",
      "stack_guid": "f9b82bf4-2344-4fb5-bff0-873452c815b9",
      "buildpack": "binary_buildpack",
      "detected_buildpack": "binary",
      "detected_buildpack_guid": "60834ee0-2690-46f4-ba24-aac2dc1c6604",
      "environment_json": {
         "NRF_CF_API_PASSWORD": "XXXXXXXXXXXX",
         "NRF_CF_API_UAA_URL": "https://uaa.YOUR-PCF-DOMAIN",
         "NRF_CF_API_URL": "https://api.YOUR-PCF-DOMAIN",
         "NRF_CF_API_USERNAME": "admin",
         "NRF_CF_CLIENT_ID": "firehose-to-newrelic",
         "NRF_CF_CLIENT_SECRET": "XXXXXXXXXXX",
         "NRF_CF_SKIP_SSL": "true",
         "NRF_FIREHOSE_ID": "newrelic.firehose",
         "NRF_NEWRELIC_ACCOUNT_ID": "",
         "NRF_NEWRELIC_ACCOUNT_REGION": "US",
         "NRF_NEWRELIC_INSERT_KEY": ""
      },
      "memory": 512,
      "instances": 2,
      "disk_quota": 256,
      "state": "STARTED",
      "version": "ec7e3437-e1ca-4db0-8c4e-9bda11a4efd5",
      "command": "./nr-fh-nozzle",
      "console": false,
      "debug": null,
      "staging_task_id": "3a9816f8-f406-4a55-8da5-7f76843e3bde",
      "package_state": "STAGED",
      "health_check_type": "http",
      "health_check_timeout": null,
      "health_check_http_endpoint": "/health",
      "staging_failed_reason": null,
      "staging_failed_description": null,
      "diego": true,
      "docker_image": null,
      "docker_credentials": {
         "username": null,
         "password": null
      },
      "package_updated_at": "2020-01-31T02:57:34Z",
      "detected_start_command": ">&2 echo Error: no start command specified during staging or launch && exit 1",
      "enable_ssh": true,
      "ports": [
         8080
      ],
      "space_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763",
      "space": {
         "metadata": {
            "guid": "4b73ec0b-2a4b-49bb-9909-174043238763",
            "url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763",
            "created_at": "2020-01-29T11:46:11Z",
            "updated_at": "2020-01-29T11:46:11Z"
         },
         "entity": {
            "name": "cfdev-space",
            "organization_guid": "d782be8d-add2-4f8b-82ea-a91ae60875e0",
            "space_quota_definition_guid": null,
            "isolation_segment_guid": null,
            "allow_ssh": true,
            "organization_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0",
            "organization": {
               "metadata": {
                  "guid": "d782be8d-add2-4f8b-82ea-a91ae60875e0",
                  "url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0",
                  "created_at": "2020-01-29T11:46:11Z",
                  "updated_at": "2020-01-29T11:46:11Z"
               },
               "entity": {
                  "name": "cfdev-org",
                  "billing_enabled": false,
                  "quota_definition_guid": "d37935da-6eee-419b-ae2d-07016a828d21",
                  "status": "active",
                  "default_isolation_segment_guid": null,
                  "quota_definition_url": "/v2/quota_definitions/d37935da-6eee-419b-ae2d-07016a828d21",
                  "spaces_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/spaces",
                  "domains_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/domains",
                  "private_domains_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/private_domains",
                  "users_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/users",
                  "managers_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/managers",
                  "billing_managers_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/billing_managers",
                  "auditors_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/auditors",
                  "app_events_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/app_events",
                  "space_quota_definitions_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/space_quota_definitions"
               }
            },
            "developers_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/developers",
            "developers": [
               {
                  "metadata": {
                     "guid": "1a0fd9cc-1969-46ba-af20-ec5d3be421b4",
                     "url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4",
                     "created_at": "2020-01-29T11:46:08Z",
                     "updated_at": "2020-01-29T11:46:08Z"
                  },
                  "entity": {
                     "admin": false,
                     "active": true,
                     "default_space_guid": null,
                     "spaces_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/spaces",
                     "organizations_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/organizations",
                     "managed_organizations_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/managed_organizations",
                     "billing_managed_organizations_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/billing_managed_organizations",
                     "audited_organizations_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/audited_organizations",
                     "managed_spaces_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/managed_spaces",
                     "audited_spaces_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/audited_spaces"
                  }
               },
               {
                  "metadata": {
                     "guid": "f25f5af9-dfbf-4a46-a3d7-7bcb77268039",
                     "url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039",
                     "created_at": "2020-01-29T11:46:10Z",
                     "updated_at": "2020-01-29T11:46:10Z"
                  },
                  "entity": {
                     "admin": false,
                     "active": false,
                     "default_space_guid": null,
                     "spaces_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/spaces",
                     "organizations_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/organizations",
                     "managed_organizations_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/managed_organizations",
                     "billing_managed_organizations_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/billing_managed_organizations",
                     "audited_organizations_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/audited_organizations",
                     "managed_spaces_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/managed_spaces",
                     "audited_spaces_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/audited_spaces"
                  }
               }
            ],
            "managers_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/managers",
            "managers": [
               {
                  "metadata": {
                     "guid": "1a0fd9cc-1969-46ba-af20-ec5d3be421b4",
                     "url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4",
                     "created_at": "2020-01-29T11:46:08Z",
                     "updated_at": "2020-01-29T11:46:08Z"
                  },
                  "entity": {
                     "admin": false,
                     "active": true,
                     "default_space_guid": null,
                     "spaces_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/spaces",
                     "organizations_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/organizations",
                     "managed_organizations_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/managed_organizations",
                     "billing_managed_organizations_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/billing_managed_organizations",
                     "audited_organizations_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/audited_organizations",
                     "managed_spaces_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/managed_spaces",
                     "audited_spaces_url": "/v2/users/1a0fd9cc-1969-46ba-af20-ec5d3be421b4/audited_spaces"
                  }
               },
               {
                  "metadata": {
                     "guid": "f25f5af9-dfbf-4a46-a3d7-7bcb77268039",
                     "url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039",
                     "created_at": "2020-01-29T11:46:10Z",
                     "updated_at": "2020-01-29T11:46:10Z"
                  },
                  "entity": {
                     "admin": false,
                     "active": false,
                     "default_space_guid": null,
                     "spaces_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/spaces",
                     "organizations_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/organizations",
                     "managed_organizations_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/managed_organizations",
                     "billing_managed_organizations_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/billing_managed_organizations",
                     "audited_organizations_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/audited_organizations",
                     "managed_spaces_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/managed_spaces",
                     "audited_spaces_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/audited_spaces"
                  }
               }
            ],
            "auditors_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/auditors",
            "auditors": [
               {
                  "metadata": {
                     "guid": "f25f5af9-dfbf-4a46-a3d7-7bcb77268039",
                     "url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039",
                     "created_at": "2020-01-29T11:46:10Z",
                     "updated_at": "2020-01-29T11:46:10Z"
                  },
                  "entity": {
                     "admin": false,
                     "active": false,
                     "default_space_guid": null,
                     "spaces_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/spaces",
                     "organizations_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/organizations",
                     "managed_organizations_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/managed_organizations",
                     "billing_managed_organizations_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/billing_managed_organizations",
                     "audited_organizations_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/audited_organizations",
                     "managed_spaces_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/managed_spaces",
                     "audited_spaces_url": "/v2/users/f25f5af9-dfbf-4a46-a3d7-7bcb77268039/audited_spaces"
                  }
               }
            ],
            "apps_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/apps",
            "routes_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/routes",
            "routes": [
               {
                  "metadata": {
                     "guid": "d6fdb0b2-1244-43dd-8faa-78faacb6b03f",
                     "url": "/v2/routes/d6fdb0b2-1244-43dd-8faa-78faacb6b03f",
                     "created_at": "2020-01-31T02:57:32Z",
                     "updated_at": "2020-01-31T02:57:32Z"
                  },
                  "entity": {
                     "host": "newrelic-firehose-nozzle",
                     "path": "",
                     "domain_guid": "ba51b2ca-e351-48b1-8d83-460e4fd50dc0",
                     "space_guid": "4b73ec0b-2a4b-49bb-9909-174043238763",
                     "service_instance_guid": null,
                     "port": null,
                     "domain_url": "/v2/shared_domains/ba51b2ca-e351-48b1-8d83-460e4fd50dc0",
                     "space_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763",
                     "apps_url": "/v2/routes/d6fdb0b2-1244-43dd-8faa-78faacb6b03f/apps",
                     "route_mappings_url": "/v2/routes/d6fdb0b2-1244-43dd-8faa-78faacb6b03f/route_mappings"
                  }
               }
            ],
            "domains_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/domains",
            "domains": [
               {
                  "metadata": {
                     "guid": "ba51b2ca-e351-48b1-8d83-460e4fd50dc0",
                     "url": "/v2/shared_domains/ba51b2ca-e351-48b1-8d83-460e4fd50dc0",
                     "created_at": "2020-01-29T11:37:57Z",
                     "updated_at": "2020-01-29T11:37:57Z"
                  },
                  "entity": {
                     "name": "dev.cfdev.sh",
                     "internal": false,
                     "router_group_guid": null,
                     "router_group_type": null
                  }
               },
               {
                  "metadata": {
                     "guid": "8092e976-507f-495e-b599-674aa92c12bb",
                     "url": "/v2/shared_domains/8092e976-507f-495e-b599-674aa92c12bb",
                     "created_at": "2020-01-29T11:37:57Z",
                     "updated_at": "2020-01-29T11:37:57Z"
                  },
                  "entity": {
                     "name": "apps.internal",
                     "internal": true,
                     "router_group_guid": null,
                     "router_group_type": null
                  }
               }
            ],
            "service_instances_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/service_instances",
            "service_instances": [],
            "app_events_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/app_events",
            "events_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/events",
            "security_groups_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/security_groups",
            "security_groups": [
               {
                  "metadata": {
                     "guid": "14e1bfce-70f0-42b7-80f2-fb420fed009d",
                     "url": "/v2/security_groups/14e1bfce-70f0-42b7-80f2-fb420fed009d",
                     "created_at": "2020-01-29T11:37:57Z",
                     "updated_at": "2020-01-29T11:37:57Z"
                  },
                  "entity": {
                     "name": "default_security_group",
                     "rules": [
                        {
                           "destination": "0.0.0.0-169.253.255.255",
                           "protocol": "all"
                        },
                        {
                           "destination": "169.255.0.0-255.255.255.255",
                           "protocol": "all"
                        }
                     ],
                     "running_default": true,
                     "staging_default": true,
                     "spaces_url": "/v2/security_groups/14e1bfce-70f0-42b7-80f2-fb420fed009d/spaces",
                     "staging_spaces_url": "/v2/security_groups/14e1bfce-70f0-42b7-80f2-fb420fed009d/staging_spaces"
                  }
               },
               {
                  "metadata": {
                     "guid": "8cf82c06-c793-428f-bebf-8d252226a193",
                     "url": "/v2/security_groups/8cf82c06-c793-428f-bebf-8d252226a193",
                     "created_at": "2020-01-29T11:37:57Z",
                     "updated_at": "2020-01-29T11:37:57Z"
                  },
                  "entity": {
                     "name": "all_access",
                     "rules": [
                        {
                           "destination": "0.0.0.0/0",
                           "protocol": "all"
                        }
                     ],
                     "running_default": true,
                     "staging_default": true,
                     "spaces_url": "/v2/security_groups/8cf82c06-c793-428f-bebf-8d252226a193/spaces",
                     "staging_spaces_url": "/v2/security_groups/8cf82c06-c793-428f-bebf-8d252226a193/staging_spaces"
                  }
               }
            ],
            "staging_security_groups_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/staging_security_groups",
            "staging_security_groups": [
               {
                  "metadata": {
                     "guid": "14e1bfce-70f0-42b7-80f2-fb420fed009d",
                     "url": "/v2/security_groups/14e1bfce-70f0-42b7-80f2-fb420fed009d",
                     "created_at": "2020-01-29T11:37:57Z",
                     "updated_at": "2020-01-29T11:37:57Z"
                  },
                  "entity": {
                     "name": "default_security_group",
                     "rules": [
                        {
                           "destination": "0.0.0.0-169.253.255.255",
                           "protocol": "all"
                        },
                        {
                           "destination": "169.255.0.0-255.255.255.255",
                           "protocol": "all"
                        }
                     ],
                     "running_default": true,
                     "staging_default": true,
                     "spaces_url": "/v2/security_groups/14e1bfce-70f0-42b7-80f2-fb420fed009d/spaces",
                     "staging_spaces_url": "/v2/security_groups/14e1bfce-70f0-42b7-80f2-fb420fed009d/staging_spaces"
                  }
               },
               {
                  "metadata": {
                     "guid": "8cf82c06-c793-428f-bebf-8d252226a193",
                     "url": "/v2/security_groups/8cf82c06-c793-428f-bebf-8d252226a193",
                     "created_at": "2020-01-29T11:37:57Z",
                     "updated_at": "2020-01-29T11:37:57Z"
                  },
                  "entity": {
                     "name": "all_access",
                     "rules": [
                        {
                           "destination": "0.0.0.0/0",
                           "protocol": "all"
                        }
                     ],
                     "running_default": true,
                     "staging_default": true,
                     "spaces_url": "/v2/security_groups/8cf82c06-c793-428f-bebf-8d252226a193/spaces",
                     "staging_spaces_url": "/v2/security_groups/8cf82c06-c793-428f-bebf-8d252226a193/staging_spaces"
                  }
               }
            ]
         }
      },
      "stack_url": "/v2/stacks/f9b82bf4-2344-4fb5-bff0-873452c815b9",
      "stack": {
         "metadata": {
            "guid": "f9b82bf4-2344-4fb5-bff0-873452c815b9",
            "url": "/v2/stacks/f9b82bf4-2344-4fb5-bff0-873452c815b9",
            "created_at": "2020-01-29T11:37:57Z",
            "updated_at": "2020-01-29T11:37:57Z"
         },
         "entity": {
            "name": "cflinuxfs3",
            "description": "Cloud Foundry Linux-based filesystem - Ubuntu Bionic 18.04 LTS"
         }
      },
      "routes_url": "/v2/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed/routes",
      "routes": [
         {
            "metadata": {
               "guid": "d6fdb0b2-1244-43dd-8faa-78faacb6b03f",
               "url": "/v2/routes/d6fdb0b2-1244-43dd-8faa-78faacb6b03f",
               "created_at": "2020-01-31T02:57:32Z",
               "updated_at": "2020-01-31T02:57:32Z"
            },
            "entity": {
               "host": "newrelic-firehose-nozzle",
               "path": "",
               "domain_guid": "ba51b2ca-e351-48b1-8d83-460e4fd50dc0",
               "space_guid": "4b73ec0b-2a4b-49bb-9909-174043238763",
               "service_instance_guid": null,
               "port": null,
               "domain_url": "/v2/shared_domains/ba51b2ca-e351-48b1-8d83-460e4fd50dc0",
               "domain": {
                  "metadata": {
                     "guid": "ba51b2ca-e351-48b1-8d83-460e4fd50dc0",
                     "url": "/v2/shared_domains/ba51b2ca-e351-48b1-8d83-460e4fd50dc0",
                     "created_at": "2020-01-29T11:37:57Z",
                     "updated_at": "2020-01-29T11:37:57Z"
                  },
                  "entity": {
                     "name": "dev.cfdev.sh",
                     "internal": false,
                     "router_group_guid": null,
                     "router_group_type": null
                  }
               },
               "space_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763",
               "space": {
                  "metadata": {
                     "guid": "4b73ec0b-2a4b-49bb-9909-174043238763",
                     "url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763",
                     "created_at": "2020-01-29T11:46:11Z",
                     "updated_at": "2020-01-29T11:46:11Z"
                  },
                  "entity": {
                     "name": "cfdev-space",
                     "organization_guid": "d782be8d-add2-4f8b-82ea-a91ae60875e0",
                     "space_quota_definition_guid": null,
                     "isolation_segment_guid": null,
                     "allow_ssh": true,
                     "organization_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0",
                     "developers_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/developers",
                     "managers_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/managers",
                     "auditors_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/auditors",
                     "apps_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/apps",
                     "routes_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/routes",
                     "domains_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/domains",
                     "service_instances_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/service_instances",
                     "app_events_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/app_events",
                     "events_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/events",
                     "security_groups_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/security_groups",
                     "staging_security_groups_url": "/v2/spaces/4b73ec0b-2a4b-49bb-9909-174043238763/staging_security_groups"
                  }
               },
               "apps_url": "/v2/routes/d6fdb0b2-1244-43dd-8faa-78faacb6b03f/apps",
               "route_mappings_url": "/v2/routes/d6fdb0b2-1244-43dd-8faa-78faacb6b03f/route_mappings"
            }
         }
      ],
      "events_url": "/v2/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed/events",
      "service_bindings_url": "/v2/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed/service_bindings",
      "service_bindings": [],
      "route_mappings_url": "/v2/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed/route_mappings"
   }
}		`)))
	case "/v2/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed/instances":
		rw.Write([]byte(fmt.Sprintf(`{
			"0": {
			"state": "CRASHED",
				"uptime": 384,
				"since": 1580440574
		},
			"1": {
			"state": "CRASHED",
				"uptime": 384,
				"since": 1580440574
		}
		}`)))
	case "/v2/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed/env":
		rw.Write([]byte(fmt.Sprintf(`
			{
   "staging_env_json": {},
   "running_env_json": {},
   "environment_json": {
      "NRF_CF_API_PASSWORD": "XXXXXXXXXXXX",
      "NRF_CF_API_UAA_URL": "https://uaa.YOUR-PCF-DOMAIN",
      "NRF_CF_API_URL": "https://api.YOUR-PCF-DOMAIN",
      "NRF_CF_API_USERNAME": "admin",
      "NRF_CF_CLIENT_ID": "firehose-to-newrelic",
      "NRF_CF_CLIENT_SECRET": "XXXXXXXXXXX",
      "NRF_CF_SKIP_SSL": "true",
      "NRF_FIREHOSE_ID": "newrelic.firehose",
      "NRF_NEWRELIC_ACCOUNT_ID": "",
      "NRF_NEWRELIC_ACCOUNT_REGION": "US",
      "NRF_NEWRELIC_INSERT_KEY": ""
   },
   "system_env_json": {
      "VCAP_SERVICES": {}
   },
   "application_env_json": {
      "VCAP_APPLICATION": {
         "cf_api": "https://api.dev.cfdev.sh",
         "limits": {
            "fds": 16384,
            "mem": 512,
            "disk": 256
         },
         "application_name": "newrelic-firehose-nozzle",
         "application_uris": [
            "newrelic-firehose-nozzle.dev.cfdev.sh"
         ],
         "name": "newrelic-firehose-nozzle",
         "space_name": "cfdev-space",
         "space_id": "4b73ec0b-2a4b-49bb-9909-174043238763",
         "uris": [
            "newrelic-firehose-nozzle.dev.cfdev.sh"
         ],
         "users": null,
         "application_id": "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
         "version": "ec7e3437-e1ca-4db0-8c4e-9bda11a4efd5",
         "application_version": "ec7e3437-e1ca-4db0-8c4e-9bda11a4efd5"
      }
   }
}
		`)))
	case "/v3/processes":
		switch r.URL.Query().Get("page") {
		case "", "1":
			rw.Write([]byte(fmt.Sprintf(`
			{
   "pagination": {
      "total_results": 3,
      "total_pages": 1,
      "first": {
         "href": "https://api.dev.cfdev.sh/v3/processes?page=1&per_page=50"
      },
      "last": {
         "href": "https://api.dev.cfdev.sh/v3/processes?page=1&per_page=50"
      },
      "next": null,
      "previous": null
   },
   "resources": [
      {
         "guid": "6fb688fd-4eca-4e91-bc42-0beee1f40eb2",
         "type": "web",
         "command": "[PRIVATE DATA HIDDEN IN LISTS]",
         "instances": 0,
         "memory_in_mb": 256,
         "disk_in_mb": 1024,
         "health_check": {
            "type": "port",
            "data": {
               "timeout": null,
               "invocation_timeout": null
            }
         },
         "created_at": "2020-01-29T11:46:26Z",
         "updated_at": "2020-01-29T11:48:16Z",
         "links": {
            "self": {
               "href": "https://api.dev.cfdev.sh/v3/processes/6fb688fd-4eca-4e91-bc42-0beee1f40eb2"
            },
            "scale": {
               "href": "https://api.dev.cfdev.sh/v3/processes/6fb688fd-4eca-4e91-bc42-0beee1f40eb2/actions/scale",
               "method": "POST"
            },
            "app": {
               "href": "https://api.dev.cfdev.sh/v3/apps/6fb688fd-4eca-4e91-bc42-0beee1f40eb2"
            },
            "space": {
               "href": "https://api.dev.cfdev.sh/v3/spaces/e218f479-d4bc-4b17-ae88-e319758eca48"
            },
            "stats": {
               "href": "https://api.dev.cfdev.sh/v3/processes/6fb688fd-4eca-4e91-bc42-0beee1f40eb2/stats"
            }
         }
      },
      {
         "guid": "323dca53-b9a8-41e4-b8c1-15d48ba1f281",
         "type": "web",
         "command": "[PRIVATE DATA HIDDEN IN LISTS]",
         "instances": 1,
         "memory_in_mb": 128,
         "disk_in_mb": 1024,
         "health_check": {
            "type": "port",
            "data": {
               "timeout": null,
               "invocation_timeout": null
            }
         },
         "created_at": "2020-01-29T11:46:27Z",
         "updated_at": "2020-01-29T11:48:03Z",
         "links": {
            "self": {
               "href": "https://api.dev.cfdev.sh/v3/processes/323dca53-b9a8-41e4-b8c1-15d48ba1f281"
            },
            "scale": {
               "href": "https://api.dev.cfdev.sh/v3/processes/323dca53-b9a8-41e4-b8c1-15d48ba1f281/actions/scale",
               "method": "POST"
            },
            "app": {
               "href": "https://api.dev.cfdev.sh/v3/apps/323dca53-b9a8-41e4-b8c1-15d48ba1f281"
            },
            "space": {
               "href": "https://api.dev.cfdev.sh/v3/spaces/e218f479-d4bc-4b17-ae88-e319758eca48"
            },
            "stats": {
               "href": "https://api.dev.cfdev.sh/v3/processes/323dca53-b9a8-41e4-b8c1-15d48ba1f281/stats"
            }
         }
      },
      {
         "guid": "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
         "type": "web",
         "command": "[PRIVATE DATA HIDDEN IN LISTS]",
         "instances": 2,
         "memory_in_mb": 512,
         "disk_in_mb": 256,
         "health_check": {
            "type": "http",
            "data": {
               "timeout": null,
               "invocation_timeout": null,
               "endpoint": "/health"
            }
         },
         "created_at": "2020-01-31T02:57:32Z",
         "updated_at": "2020-01-31T02:57:44Z",
         "links": {
            "self": {
               "href": "https://api.dev.cfdev.sh/v3/processes/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed"
            },
            "scale": {
               "href": "https://api.dev.cfdev.sh/v3/processes/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed/actions/scale",
               "method": "POST"
            },
            "app": {
               "href": "https://api.dev.cfdev.sh/v3/apps/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed"
            },
            "space": {
               "href": "https://api.dev.cfdev.sh/v3/spaces/4b73ec0b-2a4b-49bb-9909-174043238763"
            },
            "stats": {
               "href": "https://api.dev.cfdev.sh/v3/processes/c70684e2-4443-4ed5-8dc8-28b7cf7d97ed/stats"
            }
         }
      }
   ]
}`)))
		case "2":
			rw.Write([]byte(fmt.Sprintf(`
			{
   "pagination": {
      "total_results": 3,
      "total_pages": 1,
      "first": {
         "href": "https://api.dev.cfdev.sh/v3/processes?page=1&per_page=50"
      },
      "last": {
         "href": "https://api.dev.cfdev.sh/v3/processes?page=1&per_page=50"
      },
      "next": null,
      "previous": {
         "href": "https://api.dev.cfdev.sh/v3/processes?page=1&per_page=50"
      }
   },
   "resources": []
}`)))
		}
	case "/v3/spaces":
		rw.Write([]byte(fmt.Sprintf(`
			{
   "pagination": {
      "total_results": 2,
      "total_pages": 1,
      "first": {
         "href": "https://api.dev.cfdev.sh/v3/spaces?page=1&per_page=50"
      },
      "last": {
         "href": "https://api.dev.cfdev.sh/v3/spaces?page=1&per_page=50"
      },
      "next": null,
      "previous": null
   },
   "resources": [
      {
         "guid": "4b73ec0b-2a4b-49bb-9909-174043238763",
         "created_at": "2020-01-29T11:46:11Z",
         "updated_at": "2020-01-29T11:46:11Z",
         "name": "cfdev-space",
         "relationships": {
            "organization": {
               "data": {
                  "guid": "d782be8d-add2-4f8b-82ea-a91ae60875e0"
               }
            }
         },
         "links": {
            "self": {
               "href": "https://api.dev.cfdev.sh/v3/spaces/4b73ec0b-2a4b-49bb-9909-174043238763"
            },
            "organization": {
               "href": "https://api.dev.cfdev.sh/v3/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0"
            }
         }
      },
      {
         "guid": "e218f479-d4bc-4b17-ae88-e319758eca48",
         "created_at": "2020-01-29T11:46:24Z",
         "updated_at": "2020-01-29T11:46:24Z",
         "name": "system",
         "relationships": {
            "organization": {
               "data": {
                  "guid": "8614896d-6c98-4b59-9a19-2f6f15ae5fec"
               }
            }
         },
         "links": {
            "self": {
               "href": "https://api.dev.cfdev.sh/v3/spaces/e218f479-d4bc-4b17-ae88-e319758eca48"
            },
            "organization": {
               "href": "https://api.dev.cfdev.sh/v3/organizations/8614896d-6c98-4b59-9a19-2f6f15ae5fec"
            }
         }
      }
   ]
}`)))
	case "/v2/quota_definitions":
		rw.Write([]byte(fmt.Sprintf(`{
   "total_results": 2,
   "total_pages": 1,
   "prev_url": null,
   "next_url": null,
   "resources": [
      {
         "metadata": {
            "guid": "d37935da-6eee-419b-ae2d-07016a828d21",
            "url": "/v2/quota_definitions/d37935da-6eee-419b-ae2d-07016a828d21",
            "created_at": "2020-01-29T11:37:57Z",
            "updated_at": "2020-01-29T11:46:14Z"
         },
         "entity": {
            "name": "default",
            "non_basic_services_allowed": true,
            "total_services": 100,
            "total_routes": 1000,
            "total_private_domains": -1,
            "memory_limit": 10240,
            "trial_db_allowed": false,
            "instance_memory_limit": -1,
            "app_instance_limit": -1,
            "app_task_limit": -1,
            "total_service_keys": -1,
            "total_reserved_route_ports": 100
         }
      },
      {
         "metadata": {
            "guid": "3b8a4d39-7707-42aa-88eb-f64cc9da3991",
            "url": "/v2/quota_definitions/3b8a4d39-7707-42aa-88eb-f64cc9da3991",
            "created_at": "2020-01-29T11:37:57Z",
            "updated_at": "2020-01-29T11:37:57Z"
         },
         "entity": {
            "name": "runaway",
            "non_basic_services_allowed": true,
            "total_services": -1,
            "total_routes": 1000,
            "total_private_domains": -1,
            "memory_limit": 102400,
            "trial_db_allowed": false,
            "instance_memory_limit": -1,
            "app_instance_limit": -1,
            "app_task_limit": -1,
            "total_service_keys": -1,
            "total_reserved_route_ports": 0
         }
      }
   ]
}`)))
	case "/v2/organizations":
		rw.Write([]byte(fmt.Sprintf(`{
   "total_results": 2,
   "total_pages": 1,
   "prev_url": null,
   "next_url": null,
   "resources": [
      {
         "metadata": {
            "guid": "8614896d-6c98-4b59-9a19-2f6f15ae5fec",
            "url": "/v2/organizations/8614896d-6c98-4b59-9a19-2f6f15ae5fec",
            "created_at": "2020-01-29T11:37:57Z",
            "updated_at": "2020-01-29T11:46:22Z"
         },
         "entity": {
            "name": "system",
            "billing_enabled": false,
            "quota_definition_guid": "3b8a4d39-7707-42aa-88eb-f64cc9da3991",
            "status": "active",
            "default_isolation_segment_guid": null,
            "quota_definition_url": "/v2/quota_definitions/3b8a4d39-7707-42aa-88eb-f64cc9da3991",
            "spaces_url": "/v2/organizations/8614896d-6c98-4b59-9a19-2f6f15ae5fec/spaces",
            "domains_url": "/v2/organizations/8614896d-6c98-4b59-9a19-2f6f15ae5fec/domains",
            "private_domains_url": "/v2/organizations/8614896d-6c98-4b59-9a19-2f6f15ae5fec/private_domains",
            "users_url": "/v2/organizations/8614896d-6c98-4b59-9a19-2f6f15ae5fec/users",
            "managers_url": "/v2/organizations/8614896d-6c98-4b59-9a19-2f6f15ae5fec/managers",
            "billing_managers_url": "/v2/organizations/8614896d-6c98-4b59-9a19-2f6f15ae5fec/billing_managers",
            "auditors_url": "/v2/organizations/8614896d-6c98-4b59-9a19-2f6f15ae5fec/auditors",
            "app_events_url": "/v2/organizations/8614896d-6c98-4b59-9a19-2f6f15ae5fec/app_events",
            "space_quota_definitions_url": "/v2/organizations/8614896d-6c98-4b59-9a19-2f6f15ae5fec/space_quota_definitions"
         }
      },
      {
         "metadata": {
            "guid": "d782be8d-add2-4f8b-82ea-a91ae60875e0",
            "url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0",
            "created_at": "2020-01-29T11:46:11Z",
            "updated_at": "2020-01-29T11:46:11Z"
         },
         "entity": {
            "name": "cfdev-org",
            "billing_enabled": false,
            "quota_definition_guid": "d37935da-6eee-419b-ae2d-07016a828d21",
            "status": "active",
            "default_isolation_segment_guid": null,
            "quota_definition_url": "/v2/quota_definitions/d37935da-6eee-419b-ae2d-07016a828d21",
            "spaces_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/spaces",
            "domains_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/domains",
            "private_domains_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/private_domains",
            "users_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/users",
            "managers_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/managers",
            "billing_managers_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/billing_managers",
            "auditors_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/auditors",
            "app_events_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/app_events",
            "space_quota_definitions_url": "/v2/organizations/d782be8d-add2-4f8b-82ea-a91ae60875e0/space_quota_definitions"
         }
      }
   ]
}`)))
	default:
		log.Printf("Mock did not cover the API visited: %s", r.RequestURI)
	}

}
