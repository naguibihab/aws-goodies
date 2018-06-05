package main

import (
  "log"
  "net/http"
  "encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
  "github.com/aws/aws-lambda-go/events"
  
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/dynamodb"
  "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Item struct {
  Name string `json:"name"`
  Quantity string `json:"quantity"`
}

type Promotion struct {
  UUID string `json:"uuid"`
}

type CartSession struct {
  Session string `json:"session"`
  Cart []Item `json:"cart"`
  Total float64 `json:"total"`
  PromosApplied []Promotion `json:"promos"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
  
  // ************
  // Preparation
  // ************
  log.Printf("Processing Lambda request %s\n", request.PathParameters)
  
  sess, err := session.NewSession(&aws.Config{
    Region: aws.String("us-west-2")},
  )
  
  // Create DynamoDB client
  svc := dynamodb.New(sess)
  
  // ************
  // Operation
  // ************
  result, err := svc.GetItem(&dynamodb.GetItemInput{
    TableName: aws.String("Cart"),
    Key: map[string]*dynamodb.AttributeValue{
      "session": {
          S: aws.String(request.PathParameters["session"]),
      },
    },
  })
  
  if err != nil {
    log.Println(err.Error())
    return serverError(err)
  }
  
  cartSession := CartSession{}
  
  err = dynamodbattribute.UnmarshalMap(result.Item, &cartSession)
  
  if err != nil {
    log.Printf("Failed to unmarshal Record")
    return serverError(err)
  }

  if cartSession.Session == "" {
      log.Println("Could not find cart session")
  }
  
  // ************
  // Return
  // ************
  js, err := json.Marshal(cartSession)
  if err != nil {
    return serverError(err)
  }
  
  return events.APIGatewayProxyResponse{
    Headers:    map[string]string{"content-type": "application/json"},
    Body:       string(js),
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
