package arweave

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/event"
	"github.com/gin-gonic/gin"
)

type SharedArweaveDBBackupData struct {
	LastBackupTime  time.Time
	LastTransaction entity.ArweaveTransaction
	AllTransactions []entity.ArweaveTransaction
}

var (
	arweaveInstance *SharedArweaveDBBackupData
	arweaveOnce     sync.Once
)

func GetArweaveDBBackupInstance() *SharedArweaveDBBackupData {
	arweaveOnce.Do(func() {
		arweaveInstance = &SharedArweaveDBBackupData{}
		go arweaveInstance.startArweaveDBBackupBackgroundJob()
	})
	return arweaveInstance
}

var log = event.Log

// do an s.backupDBToArweave each 24 hours
func (s *SharedArweaveDBBackupData) startArweaveDBBackupBackgroundJob() {
	// get the last transaction time

	dbLatestTransaction, err := GetLatestArweaveTransactionFromDB()
	if err != nil {
		log.Error("Error getting latest transaction from DB: ", err)
	}

	s.LastTransaction = dbLatestTransaction

	//get all transactions from the database
	var allTransactions []entity.ArweaveTransaction

	if err := db.Db().Find(&allTransactions).Error; err != nil {
		log.Error("Error getting all transactions from DB: ", err)
	}

	s.AllTransactions = allTransactions

	for {
		now := time.Now()
		// Get the start of today (midnight)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		// Get  the start of the day of the last transaction
		lastTransactionDay := time.Date(s.LastTransaction.Date.Year(), s.LastTransaction.Date.Month(), s.LastTransaction.Date.Day(), 0, 0, 0, 0, s.LastTransaction.Date.Location())

		if today.After(lastTransactionDay) {
			// Perform the backup
			if initialArweaveDbData, err := s.BackupDBToArweave(); err != nil {
				log.Error("Error backing up to Arweave!: ", err)
			} else {
				s = initialArweaveDbData
				s.LastBackupTime = time.Now()

			}
		}

		// Calculate the start of the next day to determine sleep duration
		nextDay := today.Add(1 * time.Hour * 24)
		timeUntilNextBackup := nextDay.Sub(now)
		time.Sleep(timeUntilNextBackup)
	}
}

func GetLatestArweaveTransactionFromDB() (entity.ArweaveTransaction, error) {
	// Get the latest transaction from the database
	latestTransaction := &entity.ArweaveTransaction{}

	if err := db.Db().Order("date desc").First(latestTransaction).Error; err != nil {
		return *latestTransaction, err
	}

	return *latestTransaction, nil
}

func (s *SharedArweaveDBBackupData) BackupDBToArweave() (*SharedArweaveDBBackupData, error) {
	log.Println("Backing up database to Arweave.")

	// Make a POST request to the Arweave API to backup the database

	// TODO: Get the database
	// db := db.Db()

	// Prepare the request body
	requestBody, err := json.Marshal(map[string]string{})
	if err != nil {
		log.Errorf("cannot marshal request body: %s", err)
		return s, err
	}

	// Make the POST request to nodejs container
	resp, err := http.Post(os.Getenv("NODEJS_SERVER_ENDPOINT")+"/arweave/upload/string", "application/json", bytes.NewReader(requestBody))
	if err != nil {
		log.Errorf("cannot make POST request: %s", err)
		return s, err
	}
	defer resp.Body.Close()

	// Handle the response
	if resp.StatusCode != http.StatusOK {
		log.Errorf("unexpected status code: %d", resp.StatusCode)
		return s, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("cannot read response body: %s", err)
		return s, err
	}
	log.Printf("response body: %s", body)

	//get id, message and owner from body
	var bodyJson struct {
		TransactionId string `json:"transactionId"`
		Message       string `json:"message"`
		Owner         string `json:"owner"`
	}

	if err := json.Unmarshal(body, &bodyJson); err != nil {
		log.Errorf("cannot unmarshal response body: %s", err)
		return s, err
	}

	// Do the backup here
	s.LastTransaction = entity.ArweaveTransaction{
		Id:      bodyJson.TransactionId,
		Owner:   bodyJson.Owner,
		Message: bodyJson.Message,
		Date:    time.Now(),
	}

	// Save the transaction to the database
	if err := s.LastTransaction.Create(); err != nil {
		return s, err
	}

	// Append the transaction to the list of all transactions
	s.AllTransactions = append(s.AllTransactions, s.LastTransaction)

	return s, nil
}

func GetArweaveTransactions(router *gin.RouterGroup) {
	router.GET("/snapshots", func(c *gin.Context) {
		s := GetArweaveDBBackupInstance()
		c.JSON(http.StatusOK, s)
	})
}
