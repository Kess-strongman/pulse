package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func GetAuthTokenFromRequest(r *http.Request) string {
	var tmptoken string
	tmptoken = r.Header.Get("X-API-KEY")
	if tmptoken != "" {
		return tmptoken
	}
	tmptoken = r.URL.Query().Get("authtoken")
	if tmptoken != "" {
		return tmptoken
	}

	tmptoken = r.Header.Get("wf-tkn")
	if tmptoken != "" {
		return tmptoken
	}
	tmptoken = r.URL.Query().Get("wf_tkn")
	if tmptoken != "" {
		return tmptoken
	}
	tmptoken = r.URL.Query().Get("auth_tkn")
	if tmptoken != "" {
		return tmptoken
	}
	return tmptoken
}

type UserPartnerRights struct {
	PartnerID   string `json:"partnerid" validate:"required,max=30"`
	PartnerName string `json:"partnername" validate:"required,max=30"`
	Rights      []string
}

type APIResponse struct {
	Status string              `json:"status"`
	Data   []UserPartnerRights `json:"data"`
	Count  int                 `json:"count"`
	Token  string              `json:"token"`
}

func (up *UserPartnerRights) HasRight(rightName string) bool {
	for _, v := range up.Rights {
		if rightName == v {
			return true
		}

	}
	return false
}

func GetRights(tkn string) ([]UserPartnerRights, error) {
	// testtoken
	var resp APIResponse

	aresp, err := http.Get("https://proddev.airsensa.io/prod/V01/admin/rights?wf_tkn=" + tkn)
	if err != nil {
		return resp.Data, errors.New("Could not validate token")
	}
	defer aresp.Body.Close()
	body, err := ioutil.ReadAll(aresp.Body)

	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Println("ERROR - Can't unmarshal body from token login ", err)
		return resp.Data, err
	}

	return resp.Data, nil

}

func CanAdminPulse(pr []UserPartnerRights) bool {
	for _, pr := range pr {
		if pr.PartnerID == "AIRSENSA" {
			if pr.HasRight("PulseAdmin") {
				return true
			}
		}
	}
	return false
}

func TokenLogin(w http.ResponseWriter, r *http.Request) {

}
