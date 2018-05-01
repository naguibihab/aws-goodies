<?php

echo '<pre>';

require 'aws-autoloader.php';

use Aws\DynamoDb\DynamoDbClient;

$client = DynamoDbClient::factory(array(
    'region'  => 'ap-southeast-2',
    'version' => 'latest'
));

// https://docs.aws.amazon.com/aws-sdk-php/v3/api/api-dynamodb-2012-08-10.html#putitem
$resultPut = $client->putItem([
	'Item' => [ // REQUIRED
        'social_id' => [
            // 'B' => <string || resource || Psr\Http\Message\StreamInterface>,
            // 'BOOL' => true || false,
            // 'BS' => [<string || resource || Psr\Http\Message\StreamInterface>, ...],
            // 'L' => [
            //     [...], // RECURSIVE
            //     // ...
            // ],
            // 'M' => [
            //     '<AttributeName>' => [...], // RECURSIVE
            //     // ...
            // ],
            // 'N' => '<string>',
            // 'NS' => ['<string>', ...],
            // 'NULL' => true || false,
            'S' => 'test123'
            // 'SS' => ['<string>', ...],
        ],
        'social_type' => [
        	'S' => 'local'
        ]
    ],
    'TableName' => 'voia-dev-users'
]);

// https://docs.aws.amazon.com/aws-sdk-php/v3/api/api-dynamodb-2012-08-10.html#getitem
$resultGet = $client->getItem([
    // 'AttributesToGet' => ['<string>', ...],
    // 'ConsistentRead' => true || false,
    // 'ExpressionAttributeNames' => ['<string>', ...],
    'Key' => [ // REQUIRED
        'social_id' => [
            // 'B' => <string || resource || Psr\Http\Message\StreamInterface>,
            // 'BOOL' => true || false,
            // 'BS' => [<string || resource || Psr\Http\Message\StreamInterface>, ...],
            // 'L' => [
            //     [...], // RECURSIVE
            //     // ...
            // ],
            // 'M' => [
            //     '<AttributeName>' => [...], // RECURSIVE
            //     // ...
            // ],
            // 'N' => '<string>',
            // 'NS' => ['<string>', ...],
            // 'NULL' => true || false,
            'S' => 'test123',
            // 'SS' => ['<string>', ...],
        ],
        // ...
    ],
    'TableName' => 'voia-dev-users', // REQUIRED
]);

echo $resultGet;

?>