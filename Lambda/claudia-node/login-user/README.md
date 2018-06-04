# Login user
This function checks if the username and password exist in the database or not.

This function is meant to be used with an API gateway as a custom authorizer

To use this function with an API gateway custom authroizer:
- In the API gateway that you're working on go to "Authorizers"
- Create a new authroizer
- Point to this lambda function
- Add in the token source 'authorizationToken'
- Enable authroization caching (it must be enabled for the token source to pass)

## Usage:
### NPM Commands
- `npm run create` creates the function for the first time on AWS Lambda with the version set to `dev`
- `npm run release` changes the version to `production`
- `npm run deploy` updates any changes to the existing function
- `npm run test` runs a test on the function with the data in `event.json` as the input
- `npm run logs` fetches the logs from AWS CloudWatch. **Note** that logs might not get generated directly after running the lambda function, in that case it's easier to check the logs on the AWS Console
- `npm run smart-deploy` deploys, tests then fetches the log
- `npm run recreate` destroys and recreates the function
- `npm run smart-recreate` same as `recreate` but it also tests the function and then fetches the log