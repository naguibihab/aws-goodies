<?php

echo '<pre>';

require 'aws-autoloader.php';

// https://docs.aws.amazon.com/aws-sdk-php/v2/guide/service-sns.html
use Aws\Sns\SnsClient;

$client = SnsClient::factory(array(
    // 'profile' => '<profile in your aws credentials file>',
    'region'  => 'ap-southeast-2',
    'version' => 'latest'
));

// https://docs.aws.amazon.com/aws-sdk-php/v2/api/class-Aws.Sns.SnsClient.html#_publish
$result = $client->publish(array(
    'TopicArn' => 'arn:aws:sns:ap-southeast-2:350706598229:dev-pushnotifications',
    // Message is required
    'Message' => '{"default":"This is me"}',
    // Constraints: Subjects must be ASCII text that begins with a letter, number, or punctuation mark; must not include line breaks or control characters; and must be less than 100 characters long.
    'Subject' => 'This is my subject',
    'MessageStructure' => 'json'
    // 'MessageAttributes' => array(
    //     // Associative array of custom 'String' key names
    //     'String' => array(
    //         // DataType is required
    //         'DataType' => 'string',
    //         'StringValue' => 'string',
    //         'BinaryValue' => 'string',
    //     ),
    //     // ... repeated
    // ),
));

?>