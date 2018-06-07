To deploy to AWS lambda:
- Build binary file: `GOOS=linux go build -o main`
- Zip file: `zip myzip main`
- Deploy to lambda function: `aws lambda update-function-code --function-name golang-sample --zip-file fileb://myzip.zip`