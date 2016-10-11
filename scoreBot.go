package main

import (
	"fmt"
	"log"
	"flag"
	"io/ioutil"
	"strings"
	"net/http"
	
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	
	sheets "google.golang.org/api/sheets/v4"
	
	discordgo "github.com/bwmarrin/discordgo"
)

type Record struct{
	Index int
	Name string
	Holder string
	Time string
	IsTime bool
}

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
	// Ignore all messages created by the bot itself
	if m.Author.ID == discBotID {
		return
	}

	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}


func main(){
	flag.Parse()
	//initializeSheets()
	//updateInformation()
	
	initializeDiscord()
}

func updateInformation(){
	svc, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to create Sheets service: %v", err)
	}
	//SheetID: 1wECFB9MpiXswANJpyVSTdozncwwebO2NVP9o4_7FsKk
	
	//valueRequestStruct SpreadsheetsValuesGetCall
	
	// Call for the SMB IL Spreadsheet
	getCall := svc.Spreadsheets.Get("1KoneeqJzheHFYapQ_JfyxL9sI0X8_BE7ZEVMZt0t0bI")
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
			}
			/*for k, rowData := range data.RowData {
				fmt.Printf("ROW: %d", k)
				for _, value := range rowData.Values {
					fmt.Printf("Value: %s\n", value.FormattedValue)
				}
			}*/
		}
	}
}

func parseSMB1Time(data *sheets.GridData) {
	//Store the row data and the key for this category
	rowData := data.RowData
	
	// Work on Beginner Time Winners
	mapKey := "SMB1BeginnerTime"
	records[mapKey] = make([]Record, 0, 250)
	records[mapKey] = append(records[mapKey], Record{Index: 1, Name: "", Time: "", IsTime: false})
	currentIndex := 1
	
	// Retrieve Beginner Time Winners
	for i := 3; i < 13; i++ {
		name := rowData[i].Values[2].FormattedValue
		time := rowData[i].Values[3].FormattedValue
		isTime := true
		records[mapKey] = append(records[mapKey], Record{Index: currentIndex, Name: name, Time: time, IsTime: isTime})
		
		currentIndex++
	}
	records[mapKey][0].Index = currentIndex
	
	// Print information for manual checking
	for _, record := range records[mapKey] {
		fmt.Printf("%d: %s\n", record.Index, record.Name)
	}
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