make && 
serverless deploy function --function calc && 
serverless invoke -f calc
serverless logs -f calc -t