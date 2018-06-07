rm myzip.zip
rm main
GOOS=linux go build -o main
zip myzip main
aws lambda update-function-code --function-name addItemToCart --zip-file fileb://myzip.zip