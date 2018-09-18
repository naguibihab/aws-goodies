make && 
serverless deploy function --function events && 
serverless invoke -f events --path test-events/startAndEndAndOrders.json
serverless logs -f events -t