rm myzip.zip
rm main
GOOS=linux go build -o main
zip myzip main
aws lambda update-function-code --function-name editItemInCart --zip-file fileb://myzip.zip
