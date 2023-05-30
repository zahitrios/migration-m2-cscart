package services

import (
	"bytes"
	"encoding/base64"
	"strings"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

const (
	getUserByEmailEndpoint = "api/users?email="
	userEndpoint           = "api/users"
	profilesEndpoint       = "api/profiles"
)

type GamaResult struct {
	Users  []GamaUser       `json:"users"`
	Params GamaResultParams `json:"params"`
}

type GamaResultParams struct {
	TotalItems string `json:"total_items"`
}

type GamaUser struct {
	Id        string `json:"user_id"`
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	CompanyId string `json:"company_id,omitempty"`
	Status    string `json:"status"`
	UserType  string `json:"user_type"`
	Created   string `json:"created,omitempty"`
	Hash      string `json:"password,omitempty"`
}

type Profile struct {
	ProfileId   int    `json:"profile_id,omitempty"`
	ProfileName int    `json:"profile_name"`
	Sfirstname  string `json:"s_firstname"`
	Slastname   string `json:"s_lastname"`
	Saddress    string `json:"s_address"`
	Saddress2   string `json:"s_address_2"`
	Scity       string `json:"s_city"`
	Scountry    string `json:"s_country"`
	Sstate      int    `json:"s_state"`
	Szipcode    string `json:"s_zipcode"`
	Sphone      string `json:"s_phone"`
	Bfirstname  string `json:"b_firstname"`
	Blastname   string `json:"b_lastname"`
	Baddress    string `json:"b_address"`
	Baddress2   string `json:"b_address_2"`
	Bcity       string `json:"b_city"`
	Bcountry    string `json:"b_country"`
	Bstate      int    `json:"b_state"`
	Bzipcode    string `json:"b_zipcode"`
	Bphone      string `json:"b_phone"`
	Fields      Fields `json:"fields"`
}

type Fields struct {
	NumExt	   string `json:"59"`
	NumInt	   string `json:"61"`
	Reference  string `json:"63"`
	Suburb	   string `json:"65"`
	BNumExt	   string `json:"58"`
	BNumInt	   string `json:"60"`
	BReference string `json:"62"`
	BSuburb	   string `json:"64"`
}

type GamaProfileRequest struct {
	Email    string    `json:"email"`
	Profiles []Profile `json:"profiles"`
}

type GamaProfileResponse struct {
	Profiles map[int]int `json:"profiles"`
}

type GamaUpdateProfileResponse struct {
	Profiles map[int]bool `json:"profiles"`
}

type GamaUserResponse struct {
	ProfileId int `json:"profile_id,string"`
}

func GamaImportUser(magentoUser MagentoUser, force bool) (responseCode int, err error) {
	gamaResult, err := getGamaUserByEmail(magentoUser.Email)

	if err != nil {
		return 3, err
	}

	totalItems, _ := strconv.ParseInt(gamaResult.Params.TotalItems, 10, 0)

	if totalItems > 1 {
		return 3, errors.New("more of one user found in GAMA with this email")
	} else if totalItems == 1 && !force {
		return 3, errors.New("user already exists on GAMA, try send force param equals to 'true' (string)")
	} else if totalItems == 1 && force || totalItems == 0 {
		if totalItems == 1 {
			_, err = sentToGama(magentoUser, gamaResult.Users[0].Id, "update")
			if err != nil {
				return 3, err
			}
			err = sendGamaAddresses(magentoUser)
			if err != nil {
				return 3, err
			}
			return 2, err
		} else if totalItems == 0 {
			gamaUserResponse, err := sentToGama(magentoUser, "", "insert")
			if err != nil {
				return 3, err
			}
			savePrincipalProfileId(magentoUser, gamaUserResponse)
			err = sendGamaAddresses(magentoUser)
			if err != nil {
				return 3, err
			}
			return 1, err
		}
	}

	return 1, nil
}

func sentToGama(magentoUser MagentoUser, gamaUserId string, mode string) (gamaUserResponse GamaUserResponse, err error) {
	gamaUser, userHash := translateUserInformation(magentoUser, mode)
	methodRequest := http.MethodGet // to prevent not allowed actions
	url := os.Getenv("gamaUrl") + userEndpoint

	if mode == "update" {
		methodRequest = http.MethodPut
		url += "/" + gamaUserId + "&" + gamaParam
	} else if mode == "insert" {
		methodRequest = http.MethodPost
		url += "&" + gamaParam
	}

	payload, err := json.Marshal(gamaUser)
	if err != nil {
		return gamaUserResponse, errors.New("error marshaling gamaUser in order to create the payload")
	}

	client := &http.Client{}

	request, err := http.NewRequest(methodRequest, url, bytes.NewBuffer(payload))
	if err != nil {
		return gamaUserResponse, err
	}
	request.SetBasicAuth(os.Getenv("gamaUser"), os.Getenv("gamaPassword"))
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return gamaUserResponse, err
	}

	defer response.Body.Close()

	if response.StatusCode != 201 && response.StatusCode != 200 {
		return gamaUserResponse, errors.New("gama endpoint (" + url + ") returned a non 2xx status, reurned: " + response.Status)
	}
	
	saveUserHash(userHash)
	
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading resp.Body function consumeMagentoEndpoint: ", err.Error())
		return gamaUserResponse, err
	}

	json.Unmarshal(body, &gamaUserResponse)

	return gamaUserResponse, nil
}

func translateUserInformation(magentoUser MagentoUser, mode string) (gamaUser GamaUser, userHash UserHash) {
	var err error

	if mode == "update" {
		magentoUser.Hash = ""
	}
	if magentoUser.Hash != "" {
		userHash, err = GetHashFromDb(magentoUser.Email)
		userHash.Email = magentoUser.Email
		rawDecodedText, _ := base64.StdEncoding.DecodeString(magentoUser.Hash)
		hash := strings.Split(string(rawDecodedText), ":")
		if err == nil && userHash.Hash != "" {
			rawDecodedText, _ := base64.StdEncoding.DecodeString(userHash.Hash)
			hashDB := strings.Split(string(rawDecodedText), ":")
			if hashDB[1] == hash[1] {
				hash[1] = ""
			} else {
				userHash.Hash = magentoUser.Hash
			}
		} else {
			userHash.Hash = magentoUser.Hash
		}
		magentoUser.Hash = hash[1]
	}

	var user = GamaUser{
		Status:    "A",
		Firstname: magentoUser.Firstname,
		Lastname:  magentoUser.Lastname,
		Email:     magentoUser.Email,
		Hash:	   magentoUser.Hash,
	}

	if mode == "insert" {
		user.CompanyId = "0"
		user.UserType = "C"
	}

	return user,userHash
}

func getGamaUserByEmail(email string) (GamaResult, error) {
	var gamaResult = GamaResult{}

	url := os.Getenv("gamaUrl") + getUserByEmailEndpoint + email + "&" + gamaParam

	req, err := http.NewRequest("GET", url, nil) // Create a new request using http
	if err != nil {
		fmt.Println("Error create http object ("+url+") function GetGamaUserByEmail: ", err.Error())
		return gamaResult, err
	}
	req.SetBasicAuth(os.Getenv("gamaUser"), os.Getenv("gamaPassword"))

	client := &http.Client{}
	resp, err := client.Do(req)

	if resp.StatusCode != 200 {
		return gamaResult, errors.New("gama endpoint (" + url + ") returned a non 200 status, reurned: " + resp.Status)
	}

	if err != nil {
		fmt.Println("Error on request of endpoint ("+url+"): ", err.Error())
		return gamaResult, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading resp.Body function getGamaUserByEmail: ", err.Error())
		return gamaResult, err
	}

	json.Unmarshal(body, &gamaResult)

	return gamaResult, nil
}

func sendGamaAddresses(magentoUser MagentoUser) (error){
	var gamaProfileResponse = GamaProfileResponse{}
	var gamaUpdateProfileResponse = GamaUpdateProfileResponse{}
	if magentoUser.Addresses != nil {
		addressesToCreate, addressToUpdate := translateProfileInformation(magentoUser.Addresses, magentoUser)
		body, err := gamaCreateProfile(magentoUser, addressesToCreate, "insert")
		if err != nil {
			return err
		}
		json.Unmarshal(body, &gamaProfileResponse)
		checkProfilesResponse(magentoUser.Email, gamaProfileResponse)
		body, err = gamaCreateProfile(magentoUser, addressToUpdate, "update")
		if err != nil {
			return err
		}
		json.Unmarshal(body, &gamaUpdateProfileResponse)
		err = checkUpdateProfilesResponse(magentoUser.Email, gamaUpdateProfileResponse, addressToUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

func gamaCreateProfile(magentoUser MagentoUser, addresses []Profile,  mode string) ([]byte, error) {
	var gamaResponse []byte
	methodRequest := http.MethodGet // to prevent not allowed actions
	url := os.Getenv("gamaUrl") + profilesEndpoint
	gamaRequestProfile := GamaProfileRequest{
		Email:	  magentoUser.Email,
		Profiles: addresses,
	}

	if mode == "update" {
		methodRequest = http.MethodPut
		url += "/1" + "&" + gamaParam
	} else if mode == "insert" {
		methodRequest = http.MethodPost
		url += "&" + gamaParam
	}

	payload, err := json.Marshal(gamaRequestProfile)
	if err != nil {
		return gamaResponse, errors.New("error marshaling gamaRequestProfile in order to create the payload")
	}

	client := &http.Client{}

	request, err := http.NewRequest(methodRequest, url, bytes.NewBuffer(payload))
	if err != nil {
		return gamaResponse, err
	}
	request.SetBasicAuth(os.Getenv("gamaUser"), os.Getenv("gamaPassword"))
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return gamaResponse, err
	}

	defer response.Body.Close()

	if response.StatusCode != 201 && response.StatusCode != 200 {
		return gamaResponse, errors.New("gama endpoint (" + url + ") returned a non 2xx status, reurned: " + response.Status)
	}

	gamaResponse, err = ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading resp.Body function consumeMagentoEndpoint: ", err.Error())
		return gamaResponse, err
	}

	return gamaResponse, nil
}

func translateProfileInformation(addresses *[]Address, magentoUser MagentoUser) ([]Profile, []Profile) {
	var states = GetMapStates()
	var profilesToCreate, profilesToUpdate []Profile
	for _, address := range *addresses {
		var attributes = make(map[string]string)
		for _, attribute := range address.Attributes {
			attributes[attribute.Code] = attribute.Value
		}
		fields := Fields{
			NumExt: attributes["external_number"],
			NumInt: attributes["internal_number"],
			Suburb: attributes["suburb"],
			Reference: attributes["receptor_details"],
			BNumExt: attributes["external_number"],
			BNumInt: attributes["internal_number"],
			BSuburb: attributes["suburb"],
			BReference: attributes["receptor_details"],
		}
		if attributes["township"] == "" {
			attributes["township"] = address.City
		}
		profile := Profile{
			ProfileName: address.Id,
			Sfirstname: address.Firstname,
			Slastname:  address.Lastname,
			Saddress:   address.Street[0],
			Saddress2:	attributes["township"],
			Scity:      address.City,
			Scountry:   address.CountryId,
			Sstate:		states[address.Region.RegionId],
			Szipcode:   address.Postcode,
			Sphone:     address.Telephone,
			Bfirstname: address.Firstname,
			Blastname:  address.Lastname,
			Baddress:   address.Street[0],
			Baddress2:	attributes["township"],
			Bcity:      address.City,
			Bcountry:   address.CountryId,
			Bstate:		states[address.Region.RegionId],
			Bzipcode:   address.Postcode,
			Bphone:     address.Telephone,
			Fields:		fields,
		}
		migratedProfile, err := GetAddressFromDb(magentoUser.Email + fmt.Sprint(address.Id))
		if err != nil || !migratedProfile.Result {
			profilesToCreate = append(profilesToCreate, profile)
		} else {
			profile.ProfileId = migratedProfile.GamaId
			profilesToUpdate = append(profilesToUpdate, profile)
		}
	}

	return profilesToCreate, profilesToUpdate
}

func checkProfilesResponse(email string, gamaProfileResponse GamaProfileResponse) {
	for magento_id, profile_id := range gamaProfileResponse.Profiles {
		var addressProfile = AddressProfile{
			MagentoId: magento_id,
			GamaId: profile_id,
			Email: email + fmt.Sprint(magento_id),
			Result: profile_id != 0,
		}
		SaveAddressToDb(addressProfile)
	}
}

func savePrincipalProfileId(magentoUser MagentoUser, gamaUserResponse GamaUserResponse) {
	if magentoUser.DefaultShipping != 0 {
		var addressProfile = AddressProfile{
			MagentoId: magentoUser.DefaultShipping,
			GamaId: gamaUserResponse.ProfileId,
			Email: magentoUser.Email + fmt.Sprint(magentoUser.DefaultShipping),
			Result: gamaUserResponse.ProfileId != 0,
		}
		SaveAddressToDb(addressProfile)
	}
}

func checkUpdateProfilesResponse(email string, gamaUpdateProfileResponse GamaUpdateProfileResponse, addressToUpdate []Profile) error{
	for _, address := range addressToUpdate {
		var addressProfile = AddressProfile{
			MagentoId: address.ProfileName,
			GamaId: address.ProfileId,
			Email: email + fmt.Sprint(address.ProfileName),
			Result: gamaUpdateProfileResponse.Profiles[address.ProfileId],
		}
		err := SaveAddressToDb(addressProfile)
		if err != nil {
			return err
		}
	}

	return nil
}

func saveUserHash(userHash UserHash) (err error){
	if userHash.Hash != "" {
		err = SaveHashToDb(userHash)
	}
	return err
}