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

type Item struct {
  Name  string  `json:"name"`
  Stock int     `json:"stock"`
  Cost  float64 `json:"cost"`
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
  if request.PathParameters["name"] != "" {
    result, err := svc.GetItem(&dynamodb.GetItemInput{
      TableName: aws.String("Inventory"),
      Key: map[string]*dynamodb.AttributeValue{
        "name": {
          S: aws.String(request.PathParameters["name"]),
        },
      },
    })
    if err != nil {
      return serverError(err)
    }

    item := Item{}

    err = dynamodbattribute.UnmarshalMap(result.Item, &item)
    if err != nil {
      log.Printf("Failed to unmarshal Record")
      return serverError(err)
    }

    if item.Name == "" {
      log.Println("Could not find item")
    }

    // Preparing returned data
    js, err := json.Marshal(item)
    if err != nil {
      return serverError(err)
    }
    returnBody = string(js)
  } else {
    // Get all items
    params := &dynamodb.ScanInput{
      TableName: aws.String("Inventory"),
    }
    result, err := svc.Scan(params)
    if err != nil {
      return serverError(err)
    }

    var items []Item

    for _, i := range result.Items {
      item := Item{}

      err = dynamodbattribute.UnmarshalMap(i, &item)
      if err != nil {
        return serverError(err)
      }

      items = append(items, item)
    }

    // Preparting returned data
    js, err := json.Marshal(items)
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
