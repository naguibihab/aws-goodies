package main

import (
  "log"
  "net/http"
  "io/ioutil"
  "encoding/json"
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
  Cost float64 `json:"cost"`
}

type ItemCart struct {
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
  Cart []ItemCart `json:"cart"`
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
  itemInventory := new(ItemInventory)
  
  // ************
  // Operation
  // ************
  // Step 1: Find existing session or create one
  if request.PathParameters["session"] != "" {
    cartString := getUrl("/cart/"+request.PathParameters["session"])
    err := json.Unmarshal(cartString, cartSession)
    if err != nil {
      return serverError(err)
    }
    
    if cartSession.Session == "" {
      cartSession, err = addCart(svc)
    }
  } else {
    cartSession, err = addCart(svc)
  }
  
  // Step 2: Modify cart array
  
  // Step 2.1: Check if inventory has enough stock
  itemCart := new(ItemCart)
  
  // Get the item from inventory
  inventoryString := getUrl("/item/"+requestBody.Name)
  err = json.Unmarshal(inventoryString, itemInventory)
  if err != nil {
    return serverError(err)
  }
  
  // Check if quantity exceeds stock
  if itemInventory.Stock < (requestBody.Quantity + itemCart.Quantity) {
    log.Println("Error: Not enough stock",itemInventory.Stock,requestBody.Quantity,itemCart.Quantity)
    return notEnoughStockError()
  } else {
    itemCart.Cost = itemInventory.Cost
  }
  
  // Get the item from cart array
  itemIndexInCart := -1
  for i, item := range cartSession.Cart {
    if item.Name == requestBody.Name {
        itemCart = &item
        itemIndexInCart = i
    }
  }
  // if item isn't found in cart then create it
  if itemIndexInCart == -1 {
    itemCart.Name = requestBody.Name
    itemCart.Quantity = requestBody.Quantity
  }
  
  // Step 2.2: Update cart
  if itemIndexInCart > -1 {
    cartSession.Cart[itemIndexInCart].Quantity += itemCart.Quantity
  } else {
    cartSession.Cart = append(cartSession.Cart, *itemCart)
  }
  
  // Step 3: Apply promotions
  
  // Get all promotions
  var promotions []Promotion
  promoString := getUrl("/promo/")
  err = json.Unmarshal(promoString, &promotions)
  if err != nil {
    return serverError(err)
  }
  
  OUTER:
  for i, item := range cartSession.Cart {
    for _, promo := range promotions {
      // Skip applied promos
      alreadyApplied := false
      for _, appliedPromo := range cartSession.Promos {
        if promo.UUID == appliedPromo.UUID {
          alreadyApplied = true
          break
        } 
      }
      log.Println("Iterating over promo ",promo,alreadyApplied)
      if alreadyApplied {
        alreadyApplied = false
        continue
      }
      log.Println("Reach")
      
      if item.Name == promo.Affected.Name {
        // If an item in the cart can be affected by the promo
        // then start investigating if we have the affectee
        if item.Name == promo.Affectee.Name {
          // If the item is the affected and affectee
          if promo.Affected.CostPtg != 0 {
            cartSession.Cart[i].Cost *= promo.Affected.CostPtg
          } else {
            cartSession.Cart[i].Cost = promo.Affected.CostFixed
          }
          // Add promo to cart
          log.Println("Applying promo ",promo)
          cartSession.Promos = append(cartSession.Promos, promo)
          continue OUTER
        } else {
          for _, subItem := range cartSession.Cart {
            if subItem.Name == promo.Affectee.Name {
              // We have both the affected & affectee
              // time to apply the promo affect
              if promo.Affected.CostPtg != 0 {
                cartSession.Cart[i].Cost *= promo.Affected.CostPtg
              } else {
                cartSession.Cart[i].Cost = promo.Affected.CostFixed
              }
              // Add promo to cart
              log.Println("Applying promo ",promo,cartSession.Promos)
              cartSession.Promos = append(cartSession.Promos, promo)
              log.Println("Applied promo ",promo,cartSession.Promos)
              continue OUTER
            }
          }
        }
      }
    }
  }
  
  // Step 4: Calculate total cost
  // Even though this iteration already happened in another area
  // it is safer to keep different functionalities separate
  // and avoid spaghetti code as long as it does not have a big
  // impact on the performance
  cartSession.Total = 0
  for _, item := range cartSession.Cart {
    cartSession.Total += (item.Cost * float64(item.Quantity))
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

func addCart(svc *dynamodb.DynamoDB) (*CartSession, error) {
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
  log.Println("Error: "+err.Error())
  return events.APIGatewayProxyResponse{
      StatusCode: http.StatusInternalServerError,
      Body:       http.StatusText(http.StatusInternalServerError),
  }, nil
}

func notEnoughStockError() (events.APIGatewayProxyResponse, error) {
  log.Println("Not enough stock")
  return events.APIGatewayProxyResponse{
      StatusCode: http.StatusForbidden,
      Body:       http.StatusText(http.StatusForbidden),
  }, nil
}

func main() {
	lambda.Start(Handler)
}
