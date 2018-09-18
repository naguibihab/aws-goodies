package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"fmt"
	"encoding/csv"
    "net/http"
    "strings"

    "database/sql"
    _ "github.com/go-sql-driver/mysql"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

const tmpDir = "/tmp/"
const csvFile = "logs.csv"

type RequestStruct struct {
    RoamerRequestIds []string `json:"roamerRequestIds"`
}

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context,request RequestStruct) (Response, error) {
	var buf bytes.Buffer

    fmt.Printf("Running query on logs with ids: %v",request.RoamerRequestIds)

	// Get data
	queryResult, err := runCalculationsQuery(request.RoamerRequestIds)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	// Create CSV
    // This is the first row in the csv file
    // TODO: Get the headers from the struct
    headers := []string{"id",
        "created_at",
        "sub_status",
    }
	err = turnToCsv(csvFile, queryResult, headers)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

    // Upload to S3
    err = uploadToS3(csvFile,  
        os.Getenv("s3_environment"), 
        os.Getenv("s3_bucket"), 
        os.Getenv("s3_dir"))
    if err != nil {
        return Response{StatusCode: 500}, err
    }

	// Response
	body, err := json.Marshal(map[string]interface{}{
		"message": "Exported to "+os.Getenv("s3_bucket")+os.Getenv("s3_dir")+"/"+csvFile,
        "rows": len(queryResult),
	})
	if err != nil {
		return Response{StatusCode: 404}, err
	}
	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "mysql-handler",
		},
	}

	return resp, nil
}

func runCalculationsQuery(roamerRequestIds []string)([][]string, error) {

	db, err := sql.Open("mysql", os.Getenv("mysql_db_con"))
	if err != nil {
		return nil, err
	}
	defer db.Close()

    fmt.Printf("roamerRequestIds %s",strings.Join(roamerRequestIds, ","))

	// Execute the query
	rows, err := db.Query(`SELECT roamer_request.id, 
    from_unixtime(roaming_log.createdAt), 
    roaming_log.sub_status 
    FROM roaming_log 
    JOIN roamer_request ON roamer_request.id = request_id
    WHERE request_status = 2 
    AND roaming_log.sub_status = "Ready to pickup" 
    AND request_id IN ( `+strings.Join(roamerRequestIds, ",")+` ) 
    GROUP BY request_id
    ORDER BY uuid asc`)
    	if err != nil {
    		return nil, err
    	}

    	// Get column names
    	cols, err := rows.Columns()
    	if err != nil {
    		return nil, err
    	}


        // Result is your slice string.
        rawResult := make([][]byte, len(cols))
        rowResult := make([]string, len(cols))
        var allResult [][]string

        dest := make([]interface{}, len(cols)) // A temporary interface{} slice
        for i, _ := range rawResult {
            dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
        }

        for rows.Next() {
            err = rows.Scan(dest...)
            if err != nil {
                fmt.Println("Failed to scan row", err)
    			return nil, err
            }

            for i, raw := range rawResult {
                if raw == nil {
                    rowResult[i] = "\\N"
                } else {
                    rowResult[i] = string(raw)
                }

                allResult = append(allResult, rowResult)
            }
        }

        return allResult, nil
}

func turnToCsv(csvName string, data [][]string, headers []string) error {
	file, err := os.Create(tmpDir+csvName)
	if err != nil {
		fmt.Println("Cannot write to file 1", err)
		return err
	}
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Add in headers
    err = writer.Write(headers)
    if err != nil {
        fmt.Println("Cannot write to file 2", err)
        return err
    }

    for _, value := range data {
        err := writer.Write(value)
        if err != nil {
        	fmt.Println("Cannot write to file 3", err)
        	return err
        }
    }

    return nil
}

func uploadToS3(csvName string, bucketRegion string, bucketName string, destinationPath string) error {
    // Create a single AWS session (we can re use this if we're uploading many files)
    s, err := session.NewSession(&aws.Config{Region: aws.String(bucketRegion)})
    if err != nil {
        return err
    }

    // Open the file for use
    var fileDir = tmpDir+csvName
    var fileDst = destinationPath+"/"+csvName
    file, err := os.Open(fileDir)
    if err != nil {
        return err
    }
    defer file.Close()

    // Get file size and read the file content into a buffer
    fileInfo, _ := file.Stat()
    var size int64 = fileInfo.Size()
    buffer := make([]byte, size)
    file.Read(buffer)

    // Config settings: this is where you choose the bucket, filename, content-type etc.
    // of the file you're uploading.
    _, err = s3.New(s).PutObject(&s3.PutObjectInput{
        Bucket:               aws.String(bucketName),
        Key:                  aws.String(fileDst),
        ACL:                  aws.String("private"),
        Body:                 bytes.NewReader(buffer),
        ContentLength:        aws.Int64(size),
        ContentType:          aws.String(http.DetectContentType(buffer)),
        ContentDisposition:   aws.String("attachment"),
        ServerSideEncryption: aws.String("AES256"),
    })
    return err
}

func main() {
	lambda.Start(Handler)
}