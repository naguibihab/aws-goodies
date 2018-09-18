# Eta Report

## Prerequisites
### Env files:
You need to have these env files: 
- `common.env.yml`
- `mysql/mysql.env.yml`
- `orders/orders.mongo.env.yml`
- TODO

A common env file and an env file for each database we're accessing. Ask someone for these files as they are not included in the repo intentionally.
Each env file contains at least a database connection: `db_con: <connection string goes here>`
other env files that may contain more have an example in their place

### AWS access:
This codebase uses the default credentials on your machine, to configure them use `aws configure` if you want to use profiles then please update the appropriate files (I'm not sure which files you need to update, google about serverless)


## Things you should know about
Rather than using `serverless deploy` each time and deploying the whole CloudFormation stack, use the `deploy-*-.sh` file to deploy each function separately and thus making the deployment quicker, especially since there is no way to test these functions locally.

The `deploy` scripts run `make` on all the go files, so if there's an error in function A and you can't deploy function B.

These scripts would deploy, test and watch the logs, to exit after running them just hit ctrl+c  


## Using serverless
You can read about serverless framework with golang by googling but the thing you need to know about here is that if you want to change something outside the scope of a function's code (i.e. security groups or iam role) then you need to run `serverless deploy`