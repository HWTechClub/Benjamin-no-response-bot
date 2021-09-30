 
package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
)

//--------------------------------------------------------------------- Struct for broadcast.json -------------------------------------------------------------------------------------
type Broadcast struct {
	Broadcast []User_info `json:"broadcast_field"`
}

type User_info struct {
	Name string `json: "name"`
	Mobile_no   int    `json: "mobile_no"`
}

//---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

type waHandler struct {
	c *whatsapp.Conn
}

//HandleError needs to be implemented to be a valid WhatsApp handler
func (h *waHandler) HandleError(err error) {

	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("Reconnecting...")
		err := h.c.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		log.Printf("error occoured: %v\n", err)
	}
}

//Optional to be implemented. Implement HandleXXXMessage for the types you need.
func (waHandler *waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	fmt.Printf("%v %v %v %v\n\t%v\n", message.Info.Timestamp, message.Info.Id, message.Info.RemoteJid, message.ContextInfo.QuotedMessageID, message.Text)

		fmt.Println(message.Text)

	   res := strings.Contains(message.Text, "Benjamin")
	   res1 := strings.Contains(message.Text, "benjamin")

		if res || res1 {
			       fmt.Println("here")

					msg := whatsapp.TextMessage{
						Info: whatsapp.MessageInfo{
							RemoteJid: message.Info.RemoteJid,
						},
						Text:        "Thank you for contacting Benjamin Student Service Centre. \nDue to the significant volume of emails and requests, we will be responding to your query within 72 - 96 hours. \nUnable to enrol due to Finance hold\nIf you have a Finance hold on your account and are unable to enrol, please check the payment plan received from Finance. You are required to make the necessary payment in order to lift the hold and enrol. \nStudent ID card\nIf you have completed the enrolment for AY 2021-22 and would like to collect your Student ID card, please visit the Student Service Centre.\nEnrolment and course registration\nFor guidance on enrolment and course registration, please view the video links below:\nEnrolment guidance \nCourse registration guidance\nLetters/Transcripts request\nThe processing time for letters/transcript requests is 3 - 4 working days. \nDeposit refund\nThe processing time for deposit refund request is 6-8 weeks. \nAll enrolled students can log an query using ‘Ask HWU’ via student portal for a quicker response. \n\nKind Regards,\nBenjamin Student Service Centre",
					}
					waHandler.c.Send(msg)

					msgId, err := waHandler.c.Send(msg)

			if err != nil {
			fmt.Fprintf(os.Stderr, "error sending message: %v", err)
			//os.Exit(1)
			} else {
			fmt.Println("Message Sent -> ID : " + msgId)
			}


		 
		//	for i := 0; i < len(data.Broadcast); i++ {
		//		fmt.Println("name: ", data.Broadcast[i].Name)
		//		fmt.Println("mobile: ", data.Broadcast[i].Mobile_no)
		//
		//		msg := whatsapp.TextMessage{
		//			Info: whatsapp.MessageInfo{
		//				RemoteJid: strconv.Itoa(data.Broadcast[i].Mobile_no)+"@s.whatsapp.net",
		//			},
		//			Text:        "targeted spam",
		//		}
		//		waHandler.c.Send(msg)
		//
		//		msgId, err := waHandler.c.Send(msg)
	    //if err != nil {
		//  fmt.Fprintf(os.Stderr, "error sending message: %v", err)
		//  //os.Exit(1)
	    //} else {
		//  fmt.Println("Message Sent -> ID : " + msgId)
		//}
		//
		//	}
		}
       /* for{

		msg := whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: "916355536074@s.whatsapp.net",
			},
			Text:        "targeted spam",
		}

	    msgId, err := waHandler.c.Send(msg)
	    if err != nil {
		  fmt.Fprintf(os.Stderr, "error sending message: %v", err)
		  //os.Exit(1)
	    } else {
		  fmt.Println("Message Sent -> ID : " + msgId)
		}
		
	} */

}


func main() {
	//create new WhatsApp connection
	wac, err := whatsapp.NewConn(3 * time.Second)
	if err != nil {
		log.Fatalf("error creating connection: %v\n", err)
	}
	wac.SetClientVersion(2, 2123, 7)


	//Add handler
	wac.AddHandler(&waHandler{wac})

	//login or restore
	if err := login(wac); err != nil {
		log.Fatalf("error logging in: %v\n", err)
	}

	//verifies phone connectivity
	pong, err := wac.AdminTest()

	if !pong || err != nil {
		log.Fatalf("error pinging in: %v\n", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c


	//Disconnect safe
	fmt.Println("Shutting down now.")
	session, err := wac.Disconnect()
	if err != nil {
		log.Fatalf("error disconnecting: %v\n", err)
	}
	if err := writeSession(session); err != nil {
		log.Fatalf("error saving session: %v", err)
	}
}


func login(wac *whatsapp.Conn) error {
	//load saved session
	session, err := readSession()
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v\n", err)
		}
	} else {
		//no saved session -> regular login
		qr := make(chan string)
		go func() {
			terminal := qrcodeTerminal.New()
			terminal.Get(<-qr).Print()
		}()
		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("error during login: %v\n", err)
		}
	}

	//save session
	err = writeSession(session)
	if err != nil {
		return fmt.Errorf("error saving session: %v\n", err)
	}
	return nil
}

func readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	file, err := os.Open(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session) error {
	file, err := os.Create(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}