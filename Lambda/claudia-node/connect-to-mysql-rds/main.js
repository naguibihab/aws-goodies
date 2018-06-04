const ApiBuilder = require('claudia-api-builder'),
	api = new ApiBuilder();
const mysql = require('mysql'),
connection = mysql.createConnection({
    host : process.env.DB_HOST,
    user : process.env.DB_USERNAME,
    password : process.env.DB_PASSWORD,
    database: 'mock'
});

module.exports = api;

connection.connect();

api.get('/users', function() {
	return new Promise(function (resolve,reject){
		var query = "SELECT * FROM users";
		console.log("Running query",query);
		connection.query(query, function(err, rows, fields) {
			console.log('result',err,rows,fields);
			resolve(rows);
  	});
	});
});

api.post('/user', function(request) {
	return new Promise(function (resolve,reject){
		var query = "INSERT INTO users VALUES(DEFAULT,'"+request.body.user+"')";
		console.log("Running query",query);
		connection.query(query, function(err, rows, fields) {
			console.log('result',err,rows,fields);
			resolve(rows);
  	});
	});
});
