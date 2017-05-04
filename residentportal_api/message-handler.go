package residentportal_api

import (
    "net/http"
    "strconv"
	"net/url"
    "time"
    // "fmt"
    "log"
	
	"github.com/jcgarciaram/messenger"
    "github.com/tmaiaroto/aegis/lambda"
)

var client *messenger.Messenger


// VerifyWebhook returns OK
func VerifyWebhook(ctx *lambda.Context, evt *lambda.Event, res *lambda.ProxyResponse, params url.Values) {
    
    token := evt.QueryStringParameters["hub.verify_token"]
    challenge := evt.QueryStringParameters["hub.challenge"]
    
    
    if ok := client.VerifyToken(token); ok {
        res.Headers["Content-Type"] = "charset=UTF-8"
        res.StatusCode = strconv.Itoa(http.StatusOK)
        res.Body = challenge
        return
    }
    
    res.Headers["Content-Type"] = "charset=UTF-8"
    res.StatusCode = strconv.Itoa(http.StatusOK)
    res.Body = "Token didn't match"
    return
    
}


func SetUpClientHandlers(c *messenger.Messenger) {

	// Setup a handler to be triggered when a message is delivered
	c.HandleDelivery(deliveredHandler)

	// Setup a handler to be triggered when a message is read
	c.HandleRead(readHandler)
    
    // Setup a handler to be triggered when a message is received
	c.HandleMessage(receivedHandler)
    
    client = c
}


func deliveredHandler(d messenger.Delivery, r *messenger.Response) {
    log.Println("Delivered at:", d.Watermark().Format(time.UnixDate))
}

func receivedHandler(m messenger.Message, r *messenger.Response) {
    // log.Printf("%v (Sent, %v)\n", m.Text, m.Time.Format(time.UnixDate))

    log.Printf("message  ----%s----\n", m.Text)
    
    response, quickReplies, _, httpResponse := getResponse(m.Sender.ID, m.Text)
    if httpResponse != 0 {
        r.Text("Whoops, something is wrong in my brain. Can you ask me again later? I promise I'll try harder next time :) !")
        return
    }
    
    r.TextWithReplies(response, quickReplies)
    
    
}


func readHandler(m messenger.Read, r *messenger.Response) {
    log.Println("Read at:", m.Watermark().Format(time.UnixDate))
}