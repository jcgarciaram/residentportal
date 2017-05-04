package residentportal_api

import (
    r "github.com/jcgarciaram/general-api/routes"
    m "github.com/jcgarciaram/messenger"
)

var routes = r.Routes{
    
    // VerifyWebhook
    r.Route{
        "VerifyWebhook",
        "GET",
        "/v1/api/fbwebhook",
        VerifyWebhook,
    },
}

func GetRoutes(client *m.Messenger) r.Routes {
    
    clientRoute := r.Route{
        "FacebookWebhook",
        "POST",
        "/v1/api/fbwebhook",
        client.Handle,
    }
    
    routes = append(routes, clientRoute)
    
    return routes
}