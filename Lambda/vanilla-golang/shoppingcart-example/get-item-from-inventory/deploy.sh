GOOS=linux go build -o main
zip myzip main
aws lambda update-function-code --function-name getItemFromInventory --zip-file fileb://myzip.zip