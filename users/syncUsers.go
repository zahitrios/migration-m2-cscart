package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	services "migration-m2-gama/services"
)

var stage string //this var is assigned from make file on build command

type BodyRequest struct {
	Force bool   `json:"force"`
	Users []User `json:"users"`
}

type User struct {
	Email string `json:"email"`
	Hash  string `json:"hash,omitempty"`
}

type ResponseCode struct {
	Code  int    `json:"code"`
	Label string `json:"label"`
}

type BodyResults struct {
	Results       []services.BodyResult   `json:"results"`
	ResponseCodes []ResponseCode `json:"response_codes"`
}

func SyncUsers(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var magentoResult services.MagentoResults
	var err error
	responsecodes := getResponseCodeLabels()
	bodyRequest := BodyRequest{}
	bodyResults := BodyResults{}

	for label, code := range responsecodes {
		bodyResults.ResponseCodes = append(bodyResults.ResponseCodes, ResponseCode{
			Code:  code,
			Label: label,
		})
	}

	err = json.Unmarshal([]byte(request.Body), &bodyRequest)
	if err != nil {
		fmt.Println("Error destructuring the body of the request on SyncUsers function : ", err.Error())
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusBadRequest}, nil
	}

	for _, user := range bodyRequest.Users {
		force := bodyRequest.Force // this variable is used to persit magento user information if in gama the user already exists

		fmt.Println("Updating: " + user.Email + " | force: " + strconv.FormatBool(force))
		migratedUser, err := services.GetMigratedUser(user.Email)
		
		if err != nil {
			fmt.Println("Error returned by getMigratedUser function: ", err.Error())
			return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusBadRequest}, nil
		}

		if (migratedUser.ResponseCode == 2 || migratedUser.ResponseCode == 1) && !force {
			bodyResults.Results = append(bodyResults.Results, migratedUser)
		} else {
			magentoResult, err = services.GetMagentoUser(user.Email)
			if err != nil {
				fmt.Println("Error returned by GetMagentoUser function: ", err.Error())
				return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusBadRequest}, nil
			}
			if magentoResult.Total <= 0 {
				bodyResult := services.BodyResult{
					Email:        user.Email,
					ResponseCode: 3,
					Reason:       "Not user found on magento databse",
				}
				bodyResults.Results = append(bodyResults.Results, bodyResult)
				services.SaveResultToDb(bodyResult)
			}
	
			// looping for each magento item
			for _, magentoUser := range magentoResult.Items {
				magentoUser.Hash = user.Hash
				responseCode, err := services.GamaImportUser(magentoUser, force)
				if err != nil {
					bodyResult := services.BodyResult{
						Email:        magentoUser.Email,
						ResponseCode: 3,
						Reason:       err.Error(),
					}
					bodyResults.Results = append(bodyResults.Results, bodyResult)
					services.SaveResultToDb(bodyResult)
				} else {
					bodyResult := services.BodyResult{
						Email:        magentoUser.Email,
						ResponseCode: responseCode,
					}
					bodyResults.Results = append(bodyResults.Results, bodyResult)
					services.SaveResultToDb(bodyResult)
				}
			}
		}
	}

	marshaledResult, err := json.Marshal(bodyResults)
	if err != nil {
		fmt.Println("Error on marshal sync result: ", err.Error())
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(marshaledResult), StatusCode: http.StatusOK}, nil
}

func getResponseCodeLabels() map[string]int {
	return map[string]int{
		"User created successfylly":       1,
		"User updated successfylly":       2,
		"Error creating or updating user": 3,
	}
}

func main() {
	err := services.DefineEnv(stage)
	if err == nil {
		lambda.Start(SyncUsers)
	} else {
		fmt.Println("Error stage (" + stage + ") not recognized: ")
	}
}
