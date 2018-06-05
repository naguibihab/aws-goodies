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

type Affectee struct {
  Name string `json:"name"`
  Quantity int `json:"quantity"`
}

type Affected struct {
  Name  string `json:"name"`
  CostPtg float64 `json:"costPtg"`
  CostFixed float64 `json:"costFixed"`
}

type Promotion struct {
  UUID string `json:"uuid"`
  Affectee Affectee `json:"affectee"`
  Affected Affected `json:"affected"`
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
    TableName: aws.String("Promotion"),
    Key: map[string]*dynamodb.AttributeValue{
      "uuid": {
          S: aws.String(request.PathParameters["uuid"]),
      },
    },
  })
  if err != nil {
    log.Println(err.Error())
    return serverError(err)
  }
  
  promotion := Promotion{}
  
  err = dynamodbattribute.UnmarshalMap(result.Item, &promotion)
  if err != nil {
    log.Printf("Failed to unmarshal Record")
    return serverError(err)
  }

  if promotion.UUID == "" {
     log.Println("Could not find promotion, getting all promotions")
     
  }
  
  // ************
  // Return
  // ************
  js, err := json.Marshal(promotion)
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
