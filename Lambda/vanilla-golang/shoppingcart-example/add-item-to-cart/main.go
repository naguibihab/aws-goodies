package main

import (
  "log"
  "net/http"
  "io/ioutil"
  "encoding/json"
//   "encoding/hex"
  "github.com/satori/go.uuid"
	"github.com/aws/aws-lambda-go/lambda"
  "github.com/aws/aws-lambda-go/events"
  
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/dynamodb"
  "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const baseUrl = "https://tpu6ofm6o6.execute-api.us-west-2.amazonaws.com/dev"

type ItemInventory struct {
  Name string `json:"name"`
  Stock int `json:"stock"`
  Price float64 `json:"price"`
}

type ItemCart struct {
  Name string `json:"name"`
  Quantity string `json:"quantity"`
}

type Promotion struct {
  UUID string `json:"uuid"`
}

type CartSession struct {
  Session string `json:"session"`
  Cart []ItemCart `json:"cart"`
  Total float64 `json:"total"`
  PromosApplied []Promotion `json:"promos"`
}

// type Request struct {
//   Session string `json:"session"`
//   Name string `json:"name"`
//   Quantity int `json:"quantity"`
// }

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
  
  // ************
  // Preparation
  // ************
  log.Printf("Processing Lambda request %s\n", request.PathParameters)
  
  sess, err := session.NewSession(&aws.Config{
    Region: aws.String("us-west-2")},
  )
  if err != nil {
    serverError(err)
  }
  
  // Create DynamoDB client
  svc := dynamodb.New(sess)
  
  cartSession := new(CartSession)
  
  // ************
  // Operation
  // ************
  
  // Step 1: Find existing session
  if request.PathParameters["session"] != "" {
    log.Println("Reach")
    cartString := getUrl("/cart/"+request.PathParameters["session"])
    err := json.Unmarshal(cartString, cartSession)
    if err != nil {
      serverError(err)
    }
    
//     emptyUUID, err := uuid.FromString("00000000-0000-0000-0000-000000000000")
    if cartSession.Session == "" {
      cartSession, err = addCart(svc)
    }
  } else {
    cartSession, err = addCart(svc)
  }
  
  // Step 2: 
  
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

func addCart(svc *dynamodb.DynamoDB) (*CartSession, error) {
  log.Println("Reach")
  // Create UUID for new session
  uid := uuid.Must(uuid.NewV4())
  
  cartSession := CartSession{
    Session: uid.String(),
  }
  
  // Add new cart session in database
  av, err := dynamodbattribute.MarshalMap(cartSession)
  if err != nil {
      log.Println("Got error marshalling map")
      serverError(err)
      return nil, err
  }
  
  input := &dynamodb.PutItemInput{
      Item: av,
      TableName: aws.String("Cart"),
  }
  
  _, err = svc.PutItem(input)
  
  if err != nil {
      log.Println("Got error calling PutItem")
      serverError(err)
      return nil, err
  }
  
  return &cartSession, nil
}

// Function used to call other lambda functions
func getUrl(url string) ([]byte) {
  // Make a get request
  rs, err := http.Get(baseUrl + url)
  // Process response
  if err != nil {
    log.Printf("error calling url")
    serverError(err)
  }
  defer rs.Body.Close()

  bodyBytes, err := ioutil.ReadAll(rs.Body)
  if err != nil {
    log.Printf("error reading body from url")
    serverError(err)
  }
  
  return bodyBytes
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
