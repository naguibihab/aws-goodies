make && 
serverless deploy function --function logs && 
serverless invoke -f logs --path test-events/roamerRequestIds.json
serverless logs -f logs -t