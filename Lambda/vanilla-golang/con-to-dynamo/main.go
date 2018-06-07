package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
  
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/dynamodb"
  "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Request struct {
	Name float64 `json:"id"`
}

type Item struct {
  Name string `json:"name"`
  Stock int `json:"stock"`
  Price float64 `json:"price"`
}

func Handler(request Request) (Item, error) {
  
  sess, err := session.NewSession(&aws.Config{
    Region: aws.String("us-west-2")},
  )

  // Create DynamoDB client
  svc := dynamodb.New(sess)
  
  result, err := svc.GetItem(&dynamodb.GetItemInput{
    TableName: aws.String("items"),
    Key: map[string]*dynamodb.AttributeValue{
      "name": {
          S: aws.String("test"),
      },
    },
  })
  
  if err != nil {
    fmt.Println(err.Error())
  }
  
  item := Item{}
  
  err = dynamodbattribute.UnmarshalMap(result.Item, &item)
  
  if err != nil {
    panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
  }

  if item.Name == "" {
      fmt.Println("Could not find item")
  }
  
  fmt.Println("Found item:")
  fmt.Println("Name:  ", item.Name)
  fmt.Println("Stock: ", item.Stock)
  fmt.Println("Price:  ", item.Price)
  
	return item, nil
}

func main() {
	lambda.Start(Handler)
}
