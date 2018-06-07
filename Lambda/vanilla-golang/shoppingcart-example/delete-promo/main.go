package main

import (
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
  "log"
  "net/http"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/dynamodb"
)

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

  // Create DynamoDB client
  svc := dynamodb.New(sess)

  // ************
  // Operation
  // ************
  input := &dynamodb.DeleteItemInput{
    Key: map[string]*dynamodb.AttributeValue{
      "uuid": {
        S: aws.String(request.PathParameters["uuid"]),
      },
    },
    TableName: aws.String("Promotion"),
  }

  _, err = svc.DeleteItem(input)

  if err != nil {
    log.Println("Got error calling DeleteItem")
    return serverError(err)
  }

  // ************
  // Return
  // ************
  return events.APIGatewayProxyResponse{
    Headers:    map[string]string{"content-type": "application/json"},
    Body:       http.StatusText(http.StatusOK),
    StatusCode: http.StatusOK,
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
