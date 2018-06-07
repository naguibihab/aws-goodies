# ShoppingCart Example

This is an example using golang with Lambda and DynamoDB to create a simple shopping cart API

## Usage
To deploy a function run deploy.sh script located in the folder of that function

To create a new function:
- Create a `main.go` file, this file will have a single handler that will receive and handle the request 
- Create a `deploy.sh` file while noting that you'll need to change the lambda name
- Create a lambda function in AWS with the same name as the function in camel case
- Connect the lambda function to an API gateway endpoint while following RESTful design