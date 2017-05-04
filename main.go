package main

import (

    api "github.com/jcgarciaram/general-api"
    "github.com/jcgarciaram/general-api/routes"
    rp "github.com/jcgarciaram/residentportal/residentportal_api"
    "github.com/jcgarciaram/general-api/apiutils"
    
)


func main() {
	r := routes.Routes{}
    
    // Get Messenger client
    client := apiutils.CreateMessengerClient()
    
    rp.SetUpClientHandlers(client)
    
    // Append referralapp routes
    r.AppendRoutes(rp.GetRoutes(client))
    
    verifyJWT := false
    router := api.NewRouter(r, verifyJWT)


    router.Listen()
    // router.Gateway()
}
