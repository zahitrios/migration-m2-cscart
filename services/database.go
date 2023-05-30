package services

import (
	"fmt"
	"os"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type BodyResult struct {
	Email        string `json:"email"`
	ResponseCode int    `json:"response_code"`
	Reason       string `json:"reason"`
}

type AddressProfile struct {
	MagentoId int    `json:"magento_id"`
	GamaId    int    `json:"gama_id,omitempty"`
	Email     string `json:"email"`
	Result    bool   `json:"response_code"`
}

type UserHash struct {
	Email string `json:"email"`
	Hash  string `json:"hash"`
}

func SaveResultToDb(bodyResult BodyResult) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})) // Creating session for client
	svc := dynamodb.New(sess) // Create DynamoDB client

	av, err := dynamodbattribute.MarshalMap(bodyResult)
	if err != nil {
		fmt.Println("Error marshalling item: ", err.Error())
	}

	tableName := os.Getenv("MIGRATED_USERS_TABLE")
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem: ", err.Error())
	}

}

func GetMigratedUser(email string) (BodyResult, error) {
	item := BodyResult{}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})) // Creating session for client
	svc := dynamodb.New(sess) // Create DynamoDB client

	// GetItem request
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("MIGRATED_USERS_TABLE")),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
	})

	// Checking for errors, return error
	if err != nil {
		fmt.Println(err.Error())
		return item, err
	}

	// Checking type
	if len(result.Item) == 0 {
        fmt.Println("element not found")
    }

	// result is of type *dynamodb.GetItemOutput
	// result.Item is of type map[string]*dynamodb.AttributeValue
	// UnmarshallMap result.item into item
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	if err != nil {
		return item, err
	}

	return item, nil
}

func SaveAddressToDb(addressProfile AddressProfile) error{
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(addressProfile)
	if err != nil {
		fmt.Println("Error marshalling item: ", err.Error())
		return err
	}

	tableName := os.Getenv("MIGRATED_ADDRESSES_TABLE")
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem: ", err.Error())
		return err
	}

	return nil
}

func GetAddressFromDb(magentoId string) (AddressProfile, error) {
	item := AddressProfile{}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("MIGRATED_ADDRESSES_TABLE")),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(magentoId),
			},
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return item, err
	}

	if len(result.Item) == 0 {
        fmt.Println("Element not found")
		return item, errors.New("Element not found")
    }

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	if err != nil {
		return item, err
	}

	return item, nil
}

func SaveHashToDb(userHash UserHash) error{
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(userHash)
	if err != nil {
		fmt.Println("Error marshalling item: ", err.Error())
		return err
	}

	tableName := os.Getenv("MIGRATED_HASH_TABLE")
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem: ", err.Error())
		return err
	}

	return nil
}

func GetHashFromDb(email string) (UserHash, error) {
	item := UserHash{}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("MIGRATED_HASH_TABLE")),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return item, err
	}

	if len(result.Item) == 0 {
        fmt.Println("Element not found")
		return item, errors.New("Element not found")
    }

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	if err != nil {
		return item, err
	}

	return item, nil
}