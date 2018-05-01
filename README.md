# aws-goodies
Some useful reusable templates for different services in AWS

**Note** Make sure to put your AWS credentials in an `env.json` file in the root folder in the following format:
`environmentVariables = '[{"accessKey": "XXXX","secretKey": "XXXXX","region": "eu-west-2"}]'`

If you're using the PHP-SDK scripts then add these environment variables in your `httpd.conf` file

### Some AWS cli tips
If you'll use a profile for a while, you may want to create a session by running this command on windows: `set AWS_PROFILE=my_profile` or on Linux & macOs `export AWS_PROFILE=my_profile`