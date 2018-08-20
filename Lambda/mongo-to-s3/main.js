var mongoBackup = require('mongo-backup-to-s3');

var config = {
    mongodb:{
        url: process.env.MONGO_URL
    },
    s3:{
        bucket:'naguib-testing-glue',
        folder:'crawlThisLambda',
        key: process.env.AWS_ACCESS_KEY,
        secret: process.env.AWS_SECRET_KEY
    }
};

exports.handler =  function(event, context, callback) {
	mongoBackup.dumpToS3(config);
};