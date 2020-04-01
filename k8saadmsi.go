package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
	mssql "github.com/denisenkom/go-mssqldb"
)

func main() {
	fmt.Println("Starting token tester...")

	tokenProvider, err := getTokenProvider()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	server := os.Getenv("SERVER_NAME")
	database := os.Getenv("DATABASE_NAME")
	connectionString := fmt.Sprintf("Server=%s; Database=%s;", server, database)

	connector, err := mssql.NewAccessTokenConnector(
		connectionString,
		tokenProvider,
	)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	db := sql.OpenDB(connector)
	defer db.Close()

	var message string
	for {
		row := db.QueryRow(`
			select top 1 message_text
			from messagelist
			order by id desc;
		`)
		err = row.Scan(&message)
		if err != nil {
			fmt.Printf("Error %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Message is: %s\n", message)
		time.Sleep(10 * time.Second)
	}
}

func getTokenProvider() (func() (string, error), error) {
	endpoint, err := adal.GetMSIEndpoint()
	if err != nil {
		return nil, err
	}

	servicePrincipalToken, err := adal.NewServicePrincipalTokenFromMSI(endpoint, "https://database.windows.net/")
	if err != nil {
		return nil, err
	}

	return func() (string, error) {
		servicePrincipalToken.EnsureFresh()
		token := servicePrincipalToken.OAuthToken()
		return token, nil
	}, nil
}
