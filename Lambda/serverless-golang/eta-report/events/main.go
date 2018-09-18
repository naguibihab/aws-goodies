package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"fmt"
	"encoding/csv"
	"net/http"
	"reflect"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const tmpDir = "/tmp/"
const csvFile = "events.csv"

type RequestStruct struct {
	StartAt int `json:"startAt"`
	EndAt 	int `json:"endAt"`
	Orders []string `json:"orders"`
}

// Flattenning out the structure & getting the fields we need only
type EventStruct struct {
	DataId 				string		`bson:"data._id"`
	EtaToStore			int 		`bson:"data.driver.etaToStore"`
	CreatedAt			int 		`bson:"data.createdAt"`
}

// Response is of type APIGatewayProxyResponse since we"re leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context,request RequestStruct) (Response, error) {
	var buf bytes.Buffer

	fmt.Printf("Running report on start: %v and end %v and with order_ids %v \n",request.StartAt, request.EndAt, request.Orders)

	// Get data
	queryResultJson, err := runQuery(request.StartAt,request.EndAt,request.Orders)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	fmt.Printf("queryResultJson %v\n",queryResultJson)

	queryResultString := convertJsonToString(queryResultJson)

	queryResultDoubleArr := breakEachRow(queryResultString)

	// Create CSV
	// This is the first row in the csv file
	// TODO: Get the headers from the struct
	headers := []string{"id",
		"eta_to_store",
		"created_at",
	}
	err = turnToCsv(csvFile, queryResultDoubleArr, headers)
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

	// Responses
	body, err := json.Marshal(map[string]interface{}{
		"message": "Exported to "+os.Getenv("s3_bucket")+os.Getenv("s3_dir")+"/"+csvFile,
		// "rows": len(queryResultDoubleArr),
		// "data": queryResultDoubleArr,
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
			"X-MyCompany-Func-Reply": "mongo-handler",
		},
	}

	return resp, nil
}


func runQuery(startTime int, endTime int, orders []string) ([]EventStruct, error) {
    session, err := mgo.Dial(os.Getenv("events_db_con"))
    if err != nil {
        return nil, err
    }
    defer session.Close()

    // // Optional. Switch the session to a monotonic behavior.
    // session.SetMode(mgo.Monotonic, true)

    c := session.DB(os.Getenv("events_db_name")).C(os.Getenv("events_events_collection"))

    var result []EventStruct
    types := []string{"order_assigned"}

	// define private models for the data format
	type driverInner struct {
		EtaToStore int `bson:"etaToStore"`
	}

	type dataInner struct {
		Id        string      `bson:"_id" json:"_id"`
		Driver    driverInner `bson:"driver"`
		CreatedAt int         `bson:"createdAt"`
	}

	type nestedEvent struct {
		Data dataInner `bson:"data"`
	}

	var nestedResult []nestedEvent

	// Running the query
    err = c.Find(bson.M{
    	"type": bson.M{"$in" : types},
    	"data._id": bson.M{"$in" : orders},
    	"createdAt": bson.M{"$gte": startTime, "$lte": endTime},
    }).All(&nestedResult)

    /** 
	* Flattening out the structure
	**/

	for _,ne := range nestedResult {
		// create the struct in desired format
		tmp := EventStruct{
			DataId:     ne.Data.Id,
			EtaToStore: ne.Data.Driver.EtaToStore,
			CreatedAt:  ne.Data.CreatedAt,
		}

		result = append(result,tmp)
	}

    if err != nil {
		return nil, err
    }

    return result, err
}

func convertJsonToString(data []EventStruct) []string {
	var result []string
	for _, jsonRow := range data {

		v := reflect.ValueOf(jsonRow)
		n := v.NumField()
		
		st := reflect.TypeOf(jsonRow)
		headers := make([]string, n)
		for i := 0; i < n; i++ {
			// TODO: Get column names from mongo here
			headers[i] = fmt.Sprintf(`"%s": %d`, st.Field(i).Name, i)
		}
		
		rowContents := make([]string, n)
		for i := 0; i < n; i++ {
			x := v.Field(i)
			s := fmt.Sprintf("%v", x.Interface())
			if x.Type().String() == "string" {
				s = `"` + s + `"`
			}
			rowContents[i] = s
		}
		
		result = append(result,strings.Join(rowContents, ", "))
	}
	return result
}

func breakEachRow(data []string) [][]string {
	result := make([][]string, len(data))
	for key, row := range data {
		result[key] = strings.Split(row,",")
	}
	return result
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

    // write headers in the first row
    err = writer.Write(headers)
    if err != nil {
    	fmt.Println("Cannot write to file 2", err)
    	return err
    }

    for _, value := range data {
        err = writer.Write(value)
        if err != nil {
        	fmt.Println("Cannot write to file 3", err)
        	return err
        }
    }

    return nil
}

func uploadToS3(csvName string, bucketRegion string, bucketName string, destinationPath string) error {
    // Create a single AWS session (we can re use this if we"re uploading many files)
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
    // of the file you"re uploading.
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