const ApiBuilder = require('claudia-api-builder'),
  api = new ApiBuilder();
const AWS = require('aws-sdk'),
	documentClient = new AWS.DynamoDB.DocumentClient(); 
const uuid = require('uuid');

module.exports = api;

api.post('/insert-log', function(request) {
	console.log("Received request",request.body);
	return new Promise(function (resolve,reject){
		var params = {
			Item: {
				"log_id": uuid.v1(),
				"message": request.body.message
			},
			TableName: process.env.TABLE_NAME
		};

		documentClient.put(params, function(err, data){
			if(err){
				reject(err);
			} else {
				resolve(data);
			}
		});
	});
});