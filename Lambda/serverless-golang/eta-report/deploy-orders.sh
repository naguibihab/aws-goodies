make && 
serverless deploy function --function orders && 
serverless invoke -f orders --path test-events/startAndEnd.json
serverless logs -f orders -t