
# noclutter

Small Go Tool to delete emails via command line on a remote IMAP server when the email client becomes unresponsive because of too many emails.

**Requirements** - Go must be installed. Download From https://golang.org/doc/install

**To install**, simply use  `go get github.com/akshaykhairmode/noclutter`

This will install go binary in your $GOBIN (If its set) or at ~/go/bin/noclutter

Then you can run the below command to execute

Example :  `$GOBIN/noclutter -s=imap.gmail.com -p=993 -u=username@gmail.com`

This will prompt for a password, and display all the available mailboxes, we can then select a mailbox and provide subject search string which will provide number of emails matched, after user confirmation it will proceed for deletion.

**Options** 

 - -f force, allows insecure tls certificate
 - -s imap server host 
 - -p imap server port
 - -u username / email id
 - -h prints the available options
 - -e env mode, picks password from NOCLUTTER_PASS env variable

**Currently email filtering works only on subject.**