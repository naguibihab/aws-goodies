package main

import (
  "encoding/json"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/satori/go.uuid"
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

type RequestBody struct {
  AffecteeName      string  `json:"affecteeName"`
  AffecteeQuantity  int     `json:"affecteeQuantity"`
  AffectedName      string  `json:"affectedName"`
  AffectedCostPtg   float64 `json:"affectedCostPtg"`
  AffectedCostFixed float64 `json:"affectedCostFixed"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

  // ************
  // Preparation
  // ************

  // Getting the body of the request
  requestBody := new(RequestBody)
  err := json.Unmarshal([]byte(request.Body), requestBody)
  if err != nil {
    return serverError(err)
  }

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
  if requestBody.AffecteeName == "" || requestBody.AffecteeQuantity <= 0 || requestBody.AffectedName == "" {
    return parametersError()
  }

  if requestBody.AffectedCostPtg <= 0 && requestBody.AffectedCostFixed <= 0 {
    return parametersError()
  }

  // ************
  // Operation
  // ************
  affectee := Affectee{
    Name:     requestBody.AffecteeName,
    Quantity: requestBody.AffecteeQuantity,
  }
  affected := Affected{
    Name:      requestBody.AffectedName,
    CostPtg:   requestBody.AffectedCostPtg,
    CostFixed: requestBody.AffectedCostFixed,
  }
  promotion := new(Promotion)
  promotion.Affectee = affectee
  promotion.Affected = affected
  promotion, err = addPromo(svc, promotion)
  if err != nil {
    serverError(err)
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

func addPromo(svc *dynamodb.DynamoDB, promo *Promotion) (*Promotion, error) {
  // Create UUID for new session
  uid := uuid.Must(uuid.NewV4())

  promo.UUID = uid.String()

  // Add new promotion in database
  av, err := dynamodbattribute.MarshalMap(promo)
  if err != nil {
    log.Println("Got error marshalling map")
    serverError(err)
    return nil, err
  }

  input := &dynamodb.PutItemInput{
    Item:      av,
    TableName: aws.String("Promotion"),
  }

  _, err = svc.PutItem(input)

  if err != nil {
    log.Println("Got error calling PutItem")
    serverError(err)
    return nil, err
  }

  return promo, nil
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
