package main

import (
  "encoding/json"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
  "log"
  "net/http"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/dynamodb"
  "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Affectee struct {
  Name     string `json:"name"`
  Quantity int    `json:"quantity"`
}

type Affected struct {
  Name      string  `json:"name"`
  CostPtg   float64 `json:"costPtg"`
  CostFixed float64 `json:"costFixed"`
}

type Promotion struct {
  UUID     string   `json:"uuid"`
  Affectee Affectee `json:"affectee"`
  Affected Affected `json:"affected"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

  // ************
  // Preparation
  // ************
  sess, err := session.NewSession(&aws.Config{
    Region: aws.String("us-west-2")},
  )
  if err != nil {
    return serverError(err)
  }

  returnBody := ""

  // Create DynamoDB client
  svc := dynamodb.New(sess)

  // ************
  // Operation
  // ************
  if request.PathParameters["uuid"] != "" {
    result, err := svc.GetItem(&dynamodb.GetItemInput{
      TableName: aws.String("Promotion"),
      Key: map[string]*dynamodb.AttributeValue{
        "uuid": {
          S: aws.String(request.PathParameters["uuid"]),
        },
      },
    })
    if err != nil {
      return serverError(err)
    }

    promotion := Promotion{}

    err = dynamodbattribute.UnmarshalMap(result.Item, &promotion)
    if err != nil {
      log.Printf("Failed to unmarshal Record")
      return serverError(err)
    }

    if promotion.UUID == "" {
      log.Println("Could not find promotion")
    }

    // Preparing returned data
    js, err := json.Marshal(promotion)
    if err != nil {
      return serverError(err)
    }
    returnBody = string(js)
  } else {
    // Get all promotions
    params := &dynamodb.ScanInput{
      TableName: aws.String("Promotion"),
    }
    result, err := svc.Scan(params)
    if err != nil {
      return serverError(err)
    }

    var promotions []Promotion

    for _, i := range result.Items {
      promotion := Promotion{}

      err = dynamodbattribute.UnmarshalMap(i, &promotion)
      if err != nil {
        return serverError(err)
      }

      promotions = append(promotions, promotion)
    }

    // Preparting returned data
    js, err := json.Marshal(promotions)
    if err != nil {
      return serverError(err)
    }
    returnBody = string(js)
  }

  // ************
  // Return
  // ************

  return events.APIGatewayProxyResponse{
    Headers:    map[string]string{"content-type": "application/json"},
    Body:       returnBody,
    StatusCode: 200,
  }, nil
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
  log.Println(err.Error())
  return events.APIGatewayProxyResponse{
    StatusCode: http.StatusInternalServerError,
    Body:       http.StatusText(http.StatusInternalServerError),
  }, nil
}

func main() {
  lambda.Start(Handler)
}
