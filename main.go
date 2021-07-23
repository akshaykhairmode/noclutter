package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"
)

var uname, server, port, mailbox string
var help bool

const (
	MailBoxLimit = 50
)

//TODO :: Need to refactor
func main() {

	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	flag.StringVar(&mailbox, "m", "", "Mailbox which needs to be cleared")
	flag.StringVar(&server, "s", "", "Email Server host / ip (Required)")
	flag.StringVar(&port, "p", "", "Port on which to connect (Required)")
	flag.StringVar(&uname, "u", "", "Username for the email account (Required)")
	flag.BoolVar(&help, "h", false, "Help flag prints available options available")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if uname == "" || server == "" || port == "" {
		fmt.Printf("Please Pass the required flags\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Print("Please enter password : ")

	//TODO :: Fix terminal breaking if done ctrl + c when waiting for password
	pass, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("Error : %s\n", red(err))
		return
	}

	fmt.Println("\nConnecting to server...")

	// Connect to server
	c, err := client.DialTLS(server+":"+port, nil)
	if err != nil {
		fmt.Printf("Error : %s,\n", red(err))
		return
	}
	fmt.Printf("Connected to %s\n", green(server))

	defer c.Logout()

	if err := c.Login(uname, string(pass)); err != nil {
		fmt.Printf("Error : %s\n", red(err))
		return
	}

	allMailBoxes, err := getAllMailboxes(c)
	if err != nil {
		fmt.Printf("Error : %s\n", red(err))
		return
	}

	for index, mbox := range allMailBoxes {
		fmt.Printf("%v - %s\n", green(index+1), mbox)
	}

	var selectedMailbox int

	fmt.Printf("\nPlease select a mailbox from which to delete mails, %s : ", green("Enter the number and press enter"))
	if _, err := fmt.Scanln(&selectedMailbox); err != nil {
		fmt.Printf("Error : %s\n", red(err))
		return
	}

	status, err := c.Select(allMailBoxes[selectedMailbox-1], false)
	if err != nil {
		fmt.Printf("Error : %s\n", err)
		return
	}

	fmt.Printf("Selected : %s\n", green(status.Name))

	var search string
	fmt.Print("Please specify the pattern for SUBJECT for searching mails before deleting [* for all]\n")
	if _, err := fmt.Scan(&search); err != nil {
		fmt.Printf("Error : %s\n", err)
		return
	}

	criteria := imap.NewSearchCriteria()
	if search != "*" {
		//in gmail partial search wont work, need to give full subject
		criteria.Header.Add("Subject", search)
	}

	seq, err := c.Search(criteria)
	if err != nil {
		fmt.Printf("Error : %s\n", err)
		return
	}

	fmt.Printf("Total Emails Found for this search are : %s\n", green(len(seq)))
	fmt.Printf("Do you want to proceed with deletion ? [%s/%s] : ", green("Y"), red("n"))

	var proceed string
	if _, err := fmt.Scanln(&proceed); err != nil {
		fmt.Printf("Error : %s\n", err)
		return
	}

	if proceed != "Y" {
		fmt.Printf("Exiting\n")
		return
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(seq...)

	if err := c.Store(seqset, imap.AddFlags, "\\Deleted", nil); err != nil {
		fmt.Printf("Error : %s", err)
		return
	}

	fmt.Println("Mark as deleted done")

	seqChan := make(chan uint32, len(seq))
	if err := c.Expunge(seqChan); err != nil {
		fmt.Printf("Error : %s", err)
		return
	}

	fmt.Print("Expunge Completed for seq : ")
	for s := range seqChan {
		fmt.Printf("%d ", s)
	}

	fmt.Println()
}

func getAllMailboxes(c *client.Client) ([]string, error) {

	allMailbox := []string{}

	mailboxes := make(chan *imap.MailboxInfo, MailBoxLimit)

	if err := c.List("", "*", mailboxes); err != nil {
		return allMailbox, err
	}

	fmt.Println("Mailboxes:")
	for m := range mailboxes {
		allMailbox = append(allMailbox, m.Name)
	}

	return allMailbox, nil
}
