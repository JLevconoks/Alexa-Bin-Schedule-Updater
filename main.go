package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf16"
)

type BinSchedule struct {
	DocumentId  string   `json:"documentId"`
	PremisesId  string   `json:"premisesId"`
	Black       []string `json:"black"`
	Green       []string `json:"green"`
	Brown       []string `json:"brown"`
	LastUpdated string   `json:"lastUpdated"`
}

func main() {
	lambda.Start(run)
}

func run() {
	premisesId, exist := os.LookupEnv("premisesid")
	if !exist {
		log.Fatal("Missing 'premisesid' parameter")
	}

	resp, err := http.Get("http://opendata.leeds.gov.uk/downloads/bins/dm_jobs.csv")

	if err != nil {
		log.Fatal("Error opening schedule file", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)

	schedule := BinSchedule{DocumentId: fmt.Sprintf("LEEDS_%v", premisesId), PremisesId: premisesId, LastUpdated: time.Now().Format(time.RFC1123Z)}

	log.Println("Processing file")
	for scanner.Scan() {
		b := scanner.Bytes()
		s, _ := decodeUtf16([]byte(b), binary.BigEndian)
		split := strings.Split(s, ",")

		if split[0] == schedule.PremisesId {
			dateString := split[2][:strings.LastIndex(split[2], "\r")]

			switch split[1] {
			case "BLACK":
				schedule.Black = append(schedule.Black, dateString)
			case "GREEN":
				schedule.Green = append(schedule.Green, dateString)
			case "BROWN":
				schedule.Brown = append(schedule.Brown, dateString)
			}
		}
	}
	log.Println("Done processing")
	store(schedule)
	log.Println("All done.")
}

func store(binSchedule BinSchedule) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	})

	if err != nil {
		log.Fatal(err)
	}
	db := dynamodb.New(sess)

	di, err := dynamodbattribute.MarshalMap(binSchedule)
	if err != nil {
		log.Fatal(err)
	}

	input := &dynamodb.PutItemInput{
		Item:      di,
		TableName: aws.String("BinSchedule"),
	}

	_, err = db.PutItem(input)

	if err != nil {
		log.Fatal(err)
	}
}

func decodeUtf16(b []byte, order binary.ByteOrder) (string, error) {
	ints := make([]uint16, len(b)/2)
	if err := binary.Read(bytes.NewReader(b), order, &ints); err != nil {
		return "", err
	}
	return string(utf16.Decode(ints)), nil
}
