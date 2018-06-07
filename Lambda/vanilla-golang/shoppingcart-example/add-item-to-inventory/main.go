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

  // Getting the body of the request
  item := new(Item)
  err := json.Unmarshal([]byte(request.Body), item)
  if err != nil {
    return serverError(err)
  }
  
  // Getting the name from the url parameters
  item.Name = request.PathParameters["name"]

  // Preparing Dynamo
  sess, err := session.NewSession(&aws.Config{
    Region: aws.String("us-west-2")},
  )
  if err != nil {
    return serverError(err)
  }

  // Create DynamoDB client
  svc := dynamodb.New(sess)

  // Validating request
  if item.Name == "" || item.Stock <= 0 || item.Cost <= 0 {
    return parametersError()
  }

  // ************
  // Operation
  // ************
  item, err = addItem(svc, item)
  if err != nil {
    serverError(err)
  }

  // ************
  // Return
  // ************
  js, err := json.Marshal(item)
  if err != nil {
    return serverError(err)
  }

  return events.APIGatewayProxyResponse{
    Headers:    map[string]string{"content-type": "application/json"},
    Body:       string(js),
    StatusCode: 200,
  }, nil
}

func addItem(svc *dynamodb.DynamoDB, item *Item) (*Item, error) {
  // Add new item in database
  av, err := dynamodbattribute.MarshalMap(item)
  if err != nil {
    log.Println("Got error marshalling map")
    serverError(err)
    return nil, err
  }

  input := &dynamodb.PutItemInput{
    Item:      av,
    TableName: aws.String("Inventory"),
  }

  _, err = svc.PutItem(input)

  if err != nil {
    log.Println("Got error calling PutItem")
    serverError(err)
    return nil, err
  }

  return item, nil
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
  log.Println(err.Error())
  return events.APIGatewayProxyResponse{
    StatusCode: http.StatusInternalServerError,
    Body:       http.StatusText(http.StatusInternalServerError),
  }, nil
}

func parametersError() (events.APIGatewayProxyResponse, error) {
  log.Println("Parameters requirment not met")
  return events.APIGatewayProxyResponse{
    StatusCode: http.StatusPreconditionFailed,
    Body:       http.StatusText(http.StatusPreconditionFailed),
  }, nil
}

func main() {
  lambda.Start(Handler)
}
