package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	getUserEndpoint = "customers/search?searchCriteria[filter_groups][0][filters][0][field]=email&searchCriteria[filter_groups][0][filters][0][value]="
)

type Attribute struct {
	Code  string `json:"attribute_code"`
	Value string `json:"value"`
}

type Region struct {
	RegionCode string    `json:"region_code"`
	Region     string `json:"region"`
	RegionId   int `json:"region_id"`
}

type Address struct {
	Id        int      		`json:"id,omitempty"`
	Region    Region   		`json:"region"`
	CountryId string   		`json:"country_id"`
	Street    []string 		`json:"street"`
	Telephone string   		`json:"telephone"`
	Postcode  string   		`json:"postcode"`
	Firstname string   		`json:"firstname"`
	Lastname  string   		`json:"lastname"`
	City      string   		`json:"city"`
	Attributes []Attribute `json:"custom_attributes"`
}

type MagentoUser struct {
	Id        		int    	   `json:"id,omitempty"`
	Email     		string 	   `json:"email"`
	Firstname 		string 	   `json:"firstname"`
	Lastname  		string 	   `json:"lastname"`
	GroupId   		int    	   `json:"group_id"`
	DefaultShipping int    	   `json:"default_shipping,string,omitempty"`
	Addresses 		*[]Address `json:"addresses,omitempty"`
	Hash      		string     `json:"hash"`
}

type MagentoResults struct {
	Items []MagentoUser `json:"items"`
	Total int           `json:"total_count"`
}

func GetMagentoUser(email string) (MagentoResults, error) {
	magentoResults := MagentoResults{}

	response, err := consumeMagentoEndpoint(getUserEndpoint + email)

	if err != nil {
		fmt.Println("Error returned by consumeMagentoEndpoint function: ", err.Error())
		return magentoResults, err
	}

	json.Unmarshal(response, &magentoResults)

	return magentoResults, nil
}

func consumeMagentoEndpoint(request string) ([]byte, error) {
	url := os.Getenv("magentoUrl") + request
	var bearer = "Bearer " + os.Getenv("magentoBearer")

	req, err := http.NewRequest("GET", url, nil) // Create a new request using http
	if err != nil {
		fmt.Println("Error create http object ("+url+") function consumeMagentoEndpoint: ", err.Error())
		return nil, err
	}
	req.Header.Add("Authorization", bearer) // Add authorization header to the req

	client := &http.Client{}
	resp, err := client.Do(req)

	if resp.Status != "200 OK" {
		return nil, errors.New("magento endpoint (" + url + ") returned a non 200 status, reurned: " + resp.Status)
	}

	if err != nil {
		fmt.Println("Error on request of endpoint ("+url+"): ", err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading resp.Body function consumeMagentoEndpoint: ", err.Error())
		return nil, err
	}
	return body, nil
}
