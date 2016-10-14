package main

import (
	"fmt"
	"log"
	"flag"
	"io/ioutil"
	"strings"
	"net/http"
	"strconv"
	"time"
	
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	
	sheets "google.golang.org/api/sheets/v4"
	
	discordgo "github.com/bwmarrin/discordgo"
)

type Record struct{
	Index int
	Game string
	Name string
	Holder string
	Time string
	IsTime bool
}

var retrievingData bool

var skip int

var records map[string][]Record
var conf *jwt.Config
var client *http.Client

var (
	email			= flag.String("email", "", "Developer Credential email")
	emailFile			= flag.String("email-file", "clientemail.dat", "Developer Credential email stored in a file")
	privateKeyFile		= flag.String("privatekey", "", "OAuth 2.0 private key")
	applicationName = "SMB_Score_Bot"
	discToken    = flag.String("disctoken", "dtoken.dat", "Discord token stored in a file")
	discBotID	string
)

func initializeSheets() {

	// Your credentials should be obtained from the Google
	// Developer Console (https://console.developers.google.com).
	conf = &jwt.Config{
		Email: valueOrFileContents(*email, *emailFile),
		// The contents of your RSA private key or your PEM file
		// that contains a private key.
		// If you have a p12 file instead, you
		// can use `openssl` to export the private key into a pem file.
		//
		//    $ openssl pkcs12 -in key.p12 -passin pass:notasecret -out key.pem -nodes
		//
		// The field only supports PEM containers with no passphrase.
		// The openssl command will convert p12 keys to passphrase-less PEM containers.
		PrivateKey: []byte(valueOrFileContents("", *privateKeyFile)),
		Scopes: []string{sheets.SpreadsheetsReadonlyScope},
		TokenURL: google.JWTTokenURL,
		// If you would like to impersonate a user, you can
		// create a transport with a subject. The following GET
		// request will be made on the behalf of user@example.com.
		// Optional.
		//Subject: "user@example.com",
	}
	// Initiate an http.Client, the following GET request will be
	// authorized and authenticated on the behalf of user@example.com.
	client = conf.Client(oauth2.NoContext)
}

func initializeDiscord(){
	// Create a new Discord session using the provided login information.
	dg, err := discordgo.New("", "", valueOrFileContents("", *discToken))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	
	// Get the account information.
	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	// Store the account ID for later use.
	discBotID = u.ID
	
		// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {	
	whiteListChannel1 := "235536714607755264"
	whiteListChannel2 := "216643973932908544"

	// Ignore all messages created by the bot itself or if we are updating
	if retrievingData || m.Author.ID == discBotID {
		return
	}
	
	// Check for channel whiteListChannel1
	if m.ChannelID != whiteListChannel1 && m.ChannelID != whiteListChannel2 {
		return
	}

	message := m.Content
	
	if message == "!source" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Source: https://github.com/bobjrsenior/SMB_Score_Bot")
		return
	}else if message == "!author" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Created by Bobjrsenior using the SMB IL !data")
		return
	}else if message == "!data" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "IL Data: https://docs.google.com/spreadsheets/d/1KoneeqJzheHFYapQ_JfyxL9sI0X8_BE7ZEVMZt0t0bI/")
		return
	}else if message == "!update" && ((m.Author.Username == "Alex" && m.Author.Discriminator == "1806") || (m.Author.Username == "CyclopsDragon" && m.Author.Discriminator == "8762") || (m.Author.Username == "bobjrsenior" && m.Author.Discriminator == "8628")) {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Updating")
		skip++
		initializeSheets()
		updateInformation()
		return
	}
	
	// Where we are in the message
	index := 0
	
	// Make sure the message is long enough
	if len(message) >= 3 && message[index] == '!' {
		index++
		// TODO after the bot actually works
		/*
		game := ""
		if message[index] == '1' {
			game = "SMB1"
			index++
		}
		else if message[index] == '2' {
			game = "SMB2"
			index++
		}
		else if message[index] == 'D' {
			game = "SMBD"
			index++
		}
		*/
		
		// Retrieve the difficulty
		difficulty := ""
		if message[index] == 'b' {
			difficulty = "Beginner"
		}else if message[index] == 'a' {
			difficulty = "Advanced"
		}else if message[index] == 'e' {
			difficulty = "Expert"
		}else if message[index] == 'm' {
			difficulty = "Master"
		}else{
			return
		}
		index++
		
		// Check if it is an Extra stage
		if message[index] == 'x' {
			index++
			if len(message) == 3 {
				return
			}
			difficulty += "Extra"
		}
		
		// Get the level requested and convert it into an int
		levelString := message[index:len(message)]
		level, err := strconv.Atoi(levelString)
		if err != nil {
			return
		}
		// Start building the return message
		returnMessage := ""
		
		// Get the formatted record holders for smb1
		smb1Time := retrieveRecordString("SMB1", difficulty, "Time", level)
		smb1Score := retrieveRecordString("SMB1", difficulty, "Score", level)
		// If a record exists
		if smb1Time != "" {
			// Add smb1 information to the return message
			smb1Name := "SMB1 " + getLevelName("SMB1", difficulty, "Time", level)
			returnMessage += smb1Name + ": " + smb1Time + ", " + smb1Score + "\n"
		}
		
		// Get the formatted record holders for smb2
		smb2Time := retrieveRecordString("SMB2", difficulty, "Time", level)
		smb2Score := retrieveRecordString("SMB2", difficulty, "Score", level)
		// If a record exists
		if smb2Time != "" {
			// Add smb2 information to the return message
			smb2Name := "SMB2 " + getLevelName("SMB2", difficulty, "Time", level)
			returnMessage += smb2Name + ": " + smb2Time + ", " + smb2Score + "\n"
		}
		
		// Get the formatted record holders for smbd
		smbDTime := retrieveRecordString("SMBD", difficulty, "Time", level)
		smbDScore := retrieveRecordString("SMBD", difficulty, "Score", level)
		// If a record exists
		if smbDTime != "" {
			// Add smbd information to the return message
			smbDName := "SMBDX " + getLevelName("SMBD", difficulty, "Time", level)
			returnMessage += smbDName + ": " + smbDTime + ", " + smbDScore + "\n"
		}
		
		_, _ = s.ChannelMessageSend(m.ChannelID, returnMessage)
		
	}
}


func main(){
	flag.Parse()
	
	retrievingData = false
	skip = 0
	
	// Initialize google sheets connected
	initializeSheets()
	// Retrieve Information form the sheet
	updateInformation()
	
	// Connect to discord
	initializeDiscord()
}

func updateInformation(){
	retrievingData = true;
	svc, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to create Sheets service: %v", err)
	}

	// Call for the SMB IL Spreadsheet
	getCall := svc.Spreadsheets.Get("1H2LdtEo8S4tQ0MIw05UWLDY7-ARv3-W-Wu58mD7p0yY")
	if getCall == nil{
		fmt.Print("Error")
		return
	}
	//Get the data from it
	getCall = getCall.IncludeGridData(true)
	if getCall == nil{
		fmt.Print("Error")
		return
	}
	// Execute request
	spreadsheet, err := getCall.Do()
	if err != nil{
		fmt.Print("Error Executing Query")
		log.Fatal(err)
		return
	}
	
	// Initialize map for records
	records = make(map[string][]Record)
	
	// Go through every sheet and the data in each sheet
	for _, sheet := range spreadsheet.Sheets {
		for _, data := range sheet.Data {
			// Parse each sheet differently (data is in different places/a different category
			switch(sheet.Properties.Title){
				case "SMB1 Time":
					parseSMB1Time(data)
				break
				case "SMB1 Score":
					parseSMB1Score(data)
				break
				case "SMB2 Challenge Time":
					parseSMB2Time(data)
				break
				case "SMB2 Challenge Score":
					parseSMB2Score(data)
				break
				case "SMBDX Challenge Time":
					parseSMBDTime(data)
				break
				case "SMBDX Challenge Score":
					parseSMBDScore(data)
				break
			}
		}
	}
	
	retrievingData = false
	
	// Wait for a bit, then update the information
	go timeOut(7200)
	
}

func timeOut(waitInSeconds time.Duration) {
	time.Sleep(waitInSeconds * time.Second)
	if !retrievingData {
		initializeSheets()
		updateInformation()
	}else if skip != 0{
		go timeOut(7200)
	}else{
		skip--
	}
}

func parseSMB1Time(data *sheets.GridData) {
	//Store the row data and the key for this category
	rowData := data.RowData
	
	parseSection(rowData,"SMB1BeginnerTime", "SMB1", 3, 2, 10, true)
	parseSection(rowData,"SMB1BeginnerExtraTime", "SMB1", 18, 2, 3, true)
	parseSection(rowData,"SMB1BeginnerTimeAlt", "SMB1", 7, 22, 1, true)
	
	parseSection(rowData,"SMB1AdvancedTime", "SMB1", 3, 7, 30, true)
	parseSection(rowData,"SMB1AdvancedExtraTime", "SMB1", 38, 7, 5, true)
	parseSection(rowData,"SMB1AdvancedTimeAlt", "SMB1", 13, 22, 5, true)
	
	parseSection(rowData,"SMB1ExpertTime", "SMB1", 3, 12, 50, true)
	parseSection(rowData,"SMB1ExpertExtraTime", "SMB1", 58, 12, 10, true)
	parseSection(rowData,"SMB1ExpertTimeAlt", "SMB1", 24, 22, 5, true)
	
	parseSection(rowData,"SMB1MasterTime", "SMB1", 3, 17, 10, true)
}

func parseSMB1Score(data *sheets.GridData) {
	//Store the row data and the key for this category
	rowData := data.RowData
	
	parseSection(rowData,"SMB1BeginnerScore", "SMB1", 3, 2, 10, false)
	parseSection(rowData,"SMB1BeginnerExtraScore", "SMB1", 18, 2, 3, false)
	
	parseSection(rowData,"SMB1AdvancedScore", "SMB1", 3, 7, 30, false)
	parseSection(rowData,"SMB1AdvancedExtraScore", "SMB1", 38, 7, 5, false)
	
	parseSection(rowData,"SMB1ExpertScore", "SMB1", 3, 12, 50, false)
	parseSection(rowData,"SMB1ExpertExtraScore", "SMB1", 58, 12, 10, false)
	
	parseSection(rowData,"SMB1MasterScore", "SMB1", 3, 17, 10, false)
}

// Finish Alts
func parseSMB2Time(data *sheets.GridData) {
	//Store the row data and the key for this category
	rowData := data.RowData
	
	parseSection(rowData,"SMB2BeginnerTime", "SMB2", 3, 2, 10, true)
	parseSection(rowData,"SMB2BeginnerExtraTime", "SMB2", 18, 2, 10, true)
	
	parseSection(rowData,"SMB2AdvancedTime", "SMB2", 3, 7, 30, true)
	parseSection(rowData,"SMB2AdvancedExtraTime", "SMB2", 38, 7, 10, true)
	
	parseSection(rowData,"SMB2ExpertTime", "SMB2", 3, 12, 50, true)
	parseSection(rowData,"SMB2ExpertExtraTime", "SMB2", 58, 12, 10, true)
	
	parseSection(rowData,"SMB2MasterTime", "SMB2", 3, 17, 10, true)
	parseSection(rowData,"SMB2MasterExtraTime", "SMB2", 19, 17, 10, true)
}

func parseSMB2Score(data *sheets.GridData) {
	//Store the row data and the key for this category
	rowData := data.RowData
	
	parseSection(rowData,"SMB2BeginnerScore", "SMB2", 3, 2, 10, false)
	parseSection(rowData,"SMB2BeginnerExtraScore", "SMB2", 3, 2, 10, false)
	
	parseSection(rowData,"SMB2AdvancedScore", "SMB2", 3, 7, 30, false)
	parseSection(rowData,"SMB2AdvancedExtraScore", "SMB2", 38, 7, 10, false)
	
	parseSection(rowData,"SMB2ExpertScore", "SMB2", 3, 12, 50, false)
	parseSection(rowData,"SMB2ExpertExtraScore", "SMB2", 58, 12, 10, false)
	
	parseSection(rowData,"SMB2MasterScore", "SMB2", 3, 17, 10, false)
	parseSection(rowData,"SMB2MasterExtraScore", "SMB2", 19, 17, 10, false)
	
}

// Finish Alts
func parseSMBDTime(data *sheets.GridData) {
	//Store the row data and the key for this category
	rowData := data.RowData
	
	parseSection(rowData,"SMBDBeginnerTime", "SMBD", 3, 2, 40, true)
	parseSection(rowData,"SMBDBeginnerExtraTime", "SMBD", 48, 2, 20, true)
	
	parseSection(rowData,"SMBDAdvancedTime", "SMBD", 3, 7, 70, true)
	parseSection(rowData,"SMBDAdvancedExtraTime", "SMBD", 78, 7, 20, true)
	
	parseSection(rowData,"SMBDExpertTime", "SMBD", 3, 12, 100, true)
	parseSection(rowData,"SMBDExpertExtraTime", "SMBD", 108, 12, 20, true)
	
	parseSection(rowData,"SMBDMasterTime", "SMBD", 3, 17, 12, true)
	parseSection(rowData,"SMBDMasterExtraTime", "SMBD", 28, 17, 10, true)
}

func parseSMBDScore(data *sheets.GridData) {
	//Store the row data and the key for this category
	rowData := data.RowData
	
	parseSection(rowData,"SMBDBeginnerScore", "SMBD", 3, 2, 40, false)
	parseSection(rowData,"SMBDBeginnerExtraScore", "SMBD", 48, 2, 20, false)
	
	parseSection(rowData,"SMBDAdvancedScore", "SMBD", 3, 7, 70, false)
	parseSection(rowData,"SMBDAdvancedExtraScore", "SMBD", 78, 7, 20, false)
	
	parseSection(rowData,"SMBDExpertScore", "SMBD", 3, 12, 100, false)
	parseSection(rowData,"SMBDExpertExtraScore", "SMBD", 108, 12, 20, false)
	
	parseSection(rowData,"SMBDMasterScore", "SMBD", 3, 17, 12, false)
	parseSection(rowData,"SMBDMasterExtraScore", "SMBD", 28, 17, 10, false)
}

func parseSection(rowData []*sheets.RowData, mapKey string, game string, startRow int, startCol int, amount int, isTime bool){
	// Find the last row
	endRow := startRow + amount
	
	// Initialize the value in the map
	records[mapKey] = make([]Record, 0, amount)
	records[mapKey] = append(records[mapKey], Record{Index: 1, Game: game, Name: "", Holder: "", Time: "", IsTime: isTime})
	currentIndex := 1
	
	// Copy all the data into it
	for i := startRow; i < endRow; i++ {
		name := rowData[i].Values[startCol].FormattedValue
		time := rowData[i].Values[startCol + 1].FormattedValue
		holder := rowData[i].Values[startCol + 2].FormattedValue
		records[mapKey] = append(records[mapKey], Record{Index: currentIndex, Game: game, Name: name, Holder: holder, Time: time, IsTime: isTime})
		
		currentIndex++
	}
	// Update the level count
	records[mapKey][0].Index = currentIndex
}

func retrieveRecordString(game string, difficulty string, scoreType string, level int) (string){
	// Construct the mapKey
	mapKey := game + difficulty + scoreType
	
	// Does a value exist?
	if _, ok := records[mapKey]; ok {
		// If the level is not out of bounds
		if level < records[mapKey][0].Index {
			// Construct a string of the level record
			record := records[mapKey][level]
			return scoreType + ": " + record.Time + " (" + record.Holder + ")"
		}
		return ""
	}
	return ""
}

func getLevelName(game string, difficulty string, scoreType string, level int) (string){
	// Construct the mapKey
	mapKey := game + difficulty + scoreType
	
	// Does a value exist?
	if _, ok := records[mapKey]; ok {
		// If the level is not out of bounds
		if level < records[mapKey][0].Index {
			// Return the level name inside of parenthesis
			record := records[mapKey][level]
			return "(" + record.Name + ")"
		}
		return ""
	}
	return ""
}

func valueOrFileContents(value string, filename string) string {
	if value != "" {
		return value
	}
	slurp, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading %q: %v", filename, err)
	}
	return strings.TrimSpace(string(slurp))
}