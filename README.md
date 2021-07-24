
# noclutter

Small Go Tool to delete emails via command line on a remote IMAP server when the email client becomes unresponsive because of too many emails.

**Requirements** - Go must be installed.

**To install**, simply use  `go get github.com/akshaykhairmode/noclutter`

This will install go binary in your $GOBIN (If its set) or at ~/go/bin/noclutter

Then you can run the below command to execute

Example :  `$GOBIN/noclutter -s=imap.gmail.com -p=993 -u=username@gmail.com`

**Currently email filtering works only on subject.**