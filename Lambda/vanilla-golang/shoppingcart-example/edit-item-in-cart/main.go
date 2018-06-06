package main

import (
  "log"
  "net/http"
  "io/ioutil"
  "encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
  "github.com/aws/aws-lambda-go/events"
  
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/dynamodb"
  "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const baseUrl = "https://tpu6ofm6o6.execute-api.us-west-2.amazonaws.com/dev"

type Item struct {
  Name string `json:"name"`
  Quantity int `json:"quantity"`
  Cost float64 `json:"cost"`
}

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

type CartSession struct {
  Session string `json:"session"`
  Cart []Item `json:"cart"`
  Total float64 `json:"total"`
  Promos []Promotion `json:"promos"`
}

type RequestBody struct {
  Name string `json:"name"`
  Quantity int `json:"quantity"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
  
  // ************
  // Preparation
  // ************
  log.Printf("Processing Lambda request %s\n", request.PathParameters)
  
  sess, err := session.NewSession(&aws.Config{
    Region: aws.String("us-west-2")},
  )
  if err != nil {
    return serverError(err)
  }
  
  // Get body of request
  requestBody := new(RequestBody)
  err = json.Unmarshal([]byte(request.Body), requestBody)
  if err != nil {
    return serverError(err)
  }
  
  // Create DynamoDB client
  svc := dynamodb.New(sess)
  
  cartSession := new(CartSession)
  itemCart := new(Item)
  itemCart.Name = requestBody.Name
  itemCart.Quantity = requestBody.Quantity
  
  // ************
  // Operation
  // ************
  // Step 1: Find existing session or throw error
  cartString := getUrl("/cart/"+request.PathParameters["session"])
  err = json.Unmarshal(cartString, cartSession)
  if err != nil {
    return serverError(err)
  }

  if cartSession.Session == "" {
     log.Println("Sesion not found")
    return parametersError()
  }
  
  // Step 2: Modify cart array
  
  // Get item from cart
  isItemInCart := false 
  for i, item := range cartSession.Cart {
    if item.Name == itemCart.Name {
      itemCart.Quantity = itemCart.Quantity
      // If quantity = 0 then delete item
      if itemCart.Quantity == 0 {
        // If we only have one element in the array
        if len(cartSession.Cart) == 1 {
          cartSession.Cart = nil
        } else {
          // Replace with the last one & chop off the last one
          cartSession.Cart[i] = cartSession.Cart[len(cartSession.Cart)-1]
          cartSession.Cart = cartSession.Cart[:len(cartSession.Cart)-1]
        }
      } else {
        cartSession.Cart[i].Quantity = itemCart.Quantity
      }
      isItemInCart = true
      break
    }
  }
  
  // if item isn't found in cart then return error
  if !isItemInCart {
     log.Println("Item not in cart")
    return parametersError()
  }
  
  // Step 3: Check if current promotions still apply
  
  // Check if the removed item is an affectee
  
  // Because it would be dangerous to update an array we're iterating on
  // we need to use a temporary array
  tempPromos := cartSession.Promos
  
  for i, promo := range cartSession.Promos {
    if itemCart.Name == promo.Affectee.Name {
      // Check if the quantity change should affect the promo
      if itemCart.Quantity < promo.Affectee.Quantity {
        // Step 3.1 Remove Promo
        
        // If we only have one element in the array
        if len(cartSession.Promos) == 1 {
          tempPromos = nil
        } else {
          // Replace with the last one & chop off the last one
          cartSession.Promos[i] = cartSession.Promos[len(cartSession.Promos)-1]
          cartSession.Promos = cartSession.Promos[:len(cartSession.Promos)-1]
        }
        
        // Step 3.2 Update cost for Affected item from inventory
        itemInventory := new(Item)
        inventoryString := getUrl("/item/"+promo.Affected.Name)
        err = json.Unmarshal(inventoryString, itemInventory)
        if err != nil {
          return serverError(err)
        }
        
        // Find item in cart and update cost
        for j, item := range cartSession.Cart {
          if itemInventory.Name == item.Name {
            cartSession.Cart[j].Cost = itemInventory.Cost
          }
        }
      }
    }
  }
  
  // Replace our cartSession promos with the temp one
  cartSession.Promos = tempPromos
  
  // Step 4: Calculate total cost
  // Even though this iteration already happened in another area
  // it is safer to keep different functionalities separate
  // and avoid spaghetti code as long as it does not have a big
  // impact on the performance
  cartSession.Total = 0
  for _, item := range cartSession.Cart {
    cartSession.Total += (item.Cost * float64(item.Quantity))
  }
  
  // Update Cart Session
  err = updateCart(svc, cartSession)
  if err != nil {
    return serverError(err)
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

func updateCart(svc *dynamodb.DynamoDB, cartSession *CartSession) (error) {
  
 // Add new cart session in database
  av, err := dynamodbattribute.MarshalMap(cartSession)
  if err != nil {
      log.Println("Got error marshalling map")
      return err
  }
  
  input := &dynamodb.PutItemInput{
      Item: av,
      TableName: aws.String("Cart"),
  }
  
  _, err = svc.PutItem(input)
  
  if err != nil {
      log.Println("Got error calling PutItem")
      return err
  }
  
  return nil
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
  log.Println("Error: "+err.Error())
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
