package dpm

import (
	"bytes"
	"encoding/json"
	"github.com/streamsets/datacollector-edge/container/common"
	"log"
	"net/http"
	"runtime"
)

const (
	REGISTRATION_URL_PATH = "/security/public-rest/v1/components/registration"
)

type Attributes struct {
	BaseHttpUrl     string `json:"baseHttpUrl"`
	Sdc2GoGoVersion string `json:"sdc2goGoVersion"`
	Sdc2GoGoOS      string `json:"sdc2goGoOS"`
	Sdc2GoGoArch    string `json:"sdc2goGoArch"`
	Sdc2GoBuildDate string `json:"sdc2goBuildDate"`
	Sdc2GoRepoSha   string `json:"sdc2goRepoSha"`
	Sdc2GoVersion   string `json:"sdc2goVersion"`
}

type RegistrationData struct {
	AuthToken   string     `json:"authToken"`
	ComponentId string     `json:"componentId"`
	Attributes  Attributes `json:"attributes"`
}

func RegisterWithDPM(
	dpmConfig Config,
	buildInfo *common.BuildInfo,
	runtimeInfo *common.RuntimeInfo,
) {
	if dpmConfig.Enabled && dpmConfig.AppAuthToken != "" {
		attributes := Attributes{
			BaseHttpUrl:     runtimeInfo.HttpUrl,
			Sdc2GoGoVersion: runtime.Version(),
			Sdc2GoGoOS:      runtime.GOOS,
			Sdc2GoGoArch:    runtime.GOARCH,
			Sdc2GoBuildDate: buildInfo.BuiltDate,
			Sdc2GoRepoSha:   buildInfo.BuiltRepoSha,
			Sdc2GoVersion:   buildInfo.Version,
		}

		registrationData := RegistrationData{
			AuthToken:   dpmConfig.AppAuthToken,
			ComponentId: runtimeInfo.ID,
			Attributes:  attributes,
		}

		jsonValue, err := json.Marshal(registrationData)
		if err != nil {
			log.Println(err)
		}

		var registrationUrl = dpmConfig.BaseUrl + REGISTRATION_URL_PATH

		req, err := http.NewRequest("POST", registrationUrl, bytes.NewBuffer(jsonValue))
		req.Header.Set(common.HEADER_X_REST_CALL, "SDC Edge")
		req.Header.Set(common.HEADER_CONTENT_TYPE, common.APPLICATION_JSON)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		log.Println("[INFO] DPM Registration Status:", resp.Status)
		if resp.StatusCode != 200 {
			panic("DPM Registration failed")
		}
		runtimeInfo.DPMEnabled = true
		runtimeInfo.AppAuthToken = dpmConfig.AppAuthToken
	} else {
		runtimeInfo.DPMEnabled = false
	}
}
