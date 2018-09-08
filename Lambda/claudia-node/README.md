# Lambda functions

## Usage
- Install AWS CLI and login
- Install claudia.js `npm install claudia -g`
- [https://claudiajs.com/documentation.html](Claudia documentation)

### To create a lambda function
1: Create a folder with the name of that func

2: Run npm init to initate the function

3: Create a `main.js` file with a simple function

If you're using lambda without API Gateway
```
exports.handler  = function(event, context) {
    console.log("I work");
}
```

OR

If you're going to connect an API Gateway
```
const ApiBuilder = require('claudia-api-builder'),
    api = new ApiBuilder();

module.exports = api;

api.get('/hello', function() {
    return 'hello world';
});
```

4: Add a `.gitingore` file and write `node_modules` so that any modules wouldn't be committed

5: Copy and paste these scripts in package.json & adjust where necessary IF the function you're creating isn't triggered by an API Gateway
```
  "scripts": {
    "create": "claudia create --region us-west-2 --handler main.handler --version dev --security-group-ids sg-id --subnet-ids subnet-1,subnet-2,subnet-3 --policies policies/*.json --set-env-from-json env.json --profile my-aws-profile",
    "release": "claudia set-version --version production --profile my-aws-profile",
    "deploy": "claudia update --version dev --set-env-from-json env.json --profile my-aws-profile",
    "test": "claudia test-lambda --event event.json --profile my-aws-profile",
    "logs": "aws logs filter-log-events --log-group-name /aws/lambda/my-lambda --profile my-aws-profile",
    "smart-deploy": "npm run deploy && npm run test && npm run logs",
    "recreate": "claudia destroy --profile my-aws-profile && npm run create",
    "smart-recreate": "npm run recreate && npm run test && npm run logs"
  }
```

Make sure to replace: `sg-id`, `subnet-*`, `my-aws-profile`, `my-lambda`

6: If you're however creating a lambda function to be called by an API Gateway then you'd need to do the following two steps:

6.a: On the folder you've created run `npm install claudia-api-builder`

6.b: Now use the scripts mentioned in step 5 but replace the create script with this one: `claudia create --region us-west-2 --api-module main --version dev --security-group-ids sg-id --subnet-ids subnet-1,subnet-2,subnet-3 --policies policies/*.json --set-env-from-json env.json --profile my-aws-profile`

7: Add a README.md file using this template
```
# My Lambda Name

The purpose of this function is...

The function is triggered by...


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
```
