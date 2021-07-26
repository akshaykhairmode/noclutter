package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"
)

type logger struct{}

type Noclutter struct {
	uname, server, port, mailbox string
	help, force                  bool
	red, green                   func(a ...interface{}) string
}

const (
	MailBoxLimit = 50
	ctrlc        = "CTRL+C to cancel"
)

var NC Noclutter

func (l logger) Write(p []byte) (n int, err error) {
	return fmt.Print(">" + string(p))
}

func main() {

	log.SetFlags(0)
	log.SetOutput(new(logger))

	initialize()

	if err := run(); err != nil {
		log.Printf("Error : %s\n", NC.red(err))
	}

}

func run() error {

	var pass, proceed string
	var err error
	var seq []uint32
	var allMailBoxes []string

	log.Printf("Connecting to %s", NC.green(NC.server))

	tlsConfig := &tls.Config{}

	if NC.force {
		tlsConfig.InsecureSkipVerify = true
	}

	// Connect to server
	c, err := client.DialTLS(NC.getHost(), tlsConfig)
	if err != nil {
		return err
	}
	log.Println("Connected")
	defer c.Logout()

	//Get password
	if pass, err = getPasswordFromUser(); err != nil {
		return err
	}

	//Login to IMAP Server
	if err = c.Login(NC.uname, pass); err != nil {
		return err
	}

	//Get All Mailboxes from the server
	if allMailBoxes, err = getAllMailboxes(c); err != nil {
		return err
	}

	//Select mailbox for modification
	if err = selectMailbox(c, allMailBoxes); err != nil {
		return err
	}

	//search the emails based on the pattern passed for the subject
	if seq, err = searchEmails(c); err != nil {
		return err
	}

	//Get confirmation for deletion of the number of emails found
	if err := getUserInput(&proceed, "Do you want to proceed with deletion ? [%s/%s] : ", NC.green("Y"), NC.red("n")); err != nil {
		return err
	}

	if proceed != "Y" {
		fmt.Printf("Exiting\n")
		return nil
	}

	//Mark mails as deleted and expunge them
	if err := deleteEmails(c, seq); err != nil {
		return err
	}

	return nil

}

func initialize() {

	NC.red = color.New(color.FgRed).SprintFunc()
	NC.green = color.New(color.FgGreen).SprintFunc()

	flag.StringVar(&NC.mailbox, "m", "", "Mailbox which needs to be cleared")
	flag.StringVar(&NC.server, "s", "", "Email Server host / ip (Required)")
	flag.StringVar(&NC.port, "p", "", "Port on which to connect (Required)")
	flag.StringVar(&NC.uname, "u", "", "Username for the email account (Required)")
	flag.BoolVar(&NC.help, "h", false, "Help flag prints available options available")
	flag.BoolVar(&NC.force, "f", false, "force, allows insecure check when dialing")
	flag.Parse()

	if NC.help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if NC.uname == "" || NC.server == "" || NC.port == "" {
		log.Println("Please Pass the required flags")
		flag.PrintDefaults()
		os.Exit(1)
	}

}

func (nc Noclutter) getHost() string {
	return nc.server + ":" + nc.port
}

func searchEmails(c *client.Client) ([]uint32, error) {

	var search string
	if err := getUserInput(&search, "Please specify the pattern for SUBJECT for searching mails before deleting [* for all][%s]", NC.red(ctrlc)); err != nil {
		return []uint32{}, err
	}

	criteria := imap.NewSearchCriteria()
	if search != "*" {
		//in gmail partial search wont work, need to give full subject
		criteria.Header.Add("Subject", search)
	}

	seq, err := c.Search(criteria)
	if err != nil {
		return []uint32{}, err
	}

	if len(seq) <= 0 {
		return []uint32{}, fmt.Errorf(NC.red("No Mails matching this search criteria"))
	}

	log.Printf("Total Emails Found for this search are : %s\n", NC.green(len(seq)))

	return seq, nil
}

func selectMailbox(c *client.Client, allMailBoxes []string) error {

	//Print available mailboxes
	for index, mbox := range allMailBoxes {
		fmt.Printf("%v - %s\n", NC.green(index+1), mbox)
	}

	var selectedMailbox int
	if err := getUserInput(&selectedMailbox, "Please select a mailbox from which to delete mails, [%s][%s]: ", NC.green("Enter the number and press enter"), NC.red(ctrlc)); err != nil {
		return err
	}

	//Select the mailbox
	status, err := c.Select(allMailBoxes[selectedMailbox-1], false)
	if err != nil {
		return err
	}

	log.Printf("Selected : %s\n", NC.green(status.Name))

	return nil
}

func getUserInput(v interface{}, msgToPrint string, vals ...interface{}) error {

	if v == nil {
		return fmt.Errorf("Got nil input")
	}

	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("Input Must be a pointer")
	}

	log.Printf(msgToPrint+"\n", vals...)

	_, err := fmt.Scanln(v)

	return err

}

func getPasswordFromUser() (string, error) {

	log.Println("Please enter password")

	state, err := terminal.MakeRaw(0)
	if err != nil {
		return "", err
	}
	defer terminal.Restore(0, state)
	term := terminal.NewTerminal(os.Stdout, "")
	pass, err := term.ReadPassword("")
	if err != nil {
		return "", err
	}
	return pass, nil

}

func deleteEmails(c *client.Client, seq []uint32) error {

	log.Println("Deletion Started")

	seqset := new(imap.SeqSet)
	seqset.AddNum(seq...)

	if err := c.Store(seqset, imap.AddFlags, "\\Deleted", nil); err != nil {
		return err
	}

	log.Println("Mark as deleted done")

	seqChan := make(chan uint32, len(seq))
	if err := c.Expunge(seqChan); err != nil {
		return err
	}

	log.Print("Expunge Completed for seq : ")
	for s := range seqChan {
		fmt.Printf("%d ", s)
	}

	log.Println()

	return nil
}

func getAllMailboxes(c *client.Client) ([]string, error) {

	allMailbox := []string{}

	mailboxes := make(chan *imap.MailboxInfo, MailBoxLimit)

	if err := c.List("", "*", mailboxes); err != nil {
		return allMailbox, err
	}

	log.Printf("Mailboxes:")
	for m := range mailboxes {
		allMailbox = append(allMailbox, m.Name)
	}

	return allMailbox, nil
}
