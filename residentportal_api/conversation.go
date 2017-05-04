package residentportal_api

import (
    "github.com/jcgarciaram/general-api/apiutils"
    "github.com/jcgarciaram/messenger"
    
    "net/http"
    "errors"
    "fmt"
    "log"
)

// getResponse queries MySQL database to get the current state of the conversation and come up with a response.
func getResponse(senderId int64, message string) (string, []messenger.QuickReply, string, int) {
    
    var quickReplies []messenger.QuickReply
    var nodeId       int
    
    // Query to get response text
    query := 
        "SELECT " + 
            "ct.`response_text`, " +
            "ct.`node_id`, " +
            "ct.`end_of_tree` " +
        "FROM " +
            "`conversation` c " +
            "LEFT OUTER JOIN `conversation_tree` ct ON c.`current_state` = ct.`parent_node_id` " +
        "WHERE " +
            "c.`sender_id` = ? " +
            "AND (ct.`parent_branch_value` = ? OR ct.`parent_branch_value` = '*')"
    
    // Run query in MySQL
    getTotalCount := false
    schema := "residentportal"
    parameters := []interface{}{senderId, message}
    rowMapSlice, _, errStr, httpResponse := apiutils.RunSelectQuery(schema, query, parameters, getTotalCount)
    if httpResponse != 0 {
        log.Println("Error running query 1")
        return "", nil, errStr, httpResponse
    }

    // If no rows are returned, this is the beginning of the conversation send first message
    if len(rowMapSlice) == 0 {
        
        nodeId = 0
        
        query = 
            "SELECT " + 
                "ct.`response_text` " +
            "FROM " +
                "`conversation_tree` ct " +
            "WHERE " +
                "ct.`parent_node_id` = -1"
        
        
        // Run query from MySQL
        parameters = []interface{}{}
        rowMapSlice, _, errStr, httpResponse = apiutils.RunSelectQuery(schema, query, parameters, getTotalCount)
        if httpResponse != 0 {
            return "", nil, errStr, httpResponse
        }
        
        if err := insertSenderCurrentState(senderId, nodeId); err != nil {
            log.Println("Error inserting new sender")
            return "", nil, err.Error(), http.StatusInternalServerError
        }
    
    
    // If a response is returned, you already have your response. Now we have to check how the current state of the conversation will be updated in MySQL
    } else {
        
        // Verify if this is the end of the conversation tree
        endOfTree, err := apiutils.InterfaceToInt(rowMapSlice[0]["end_of_tree"])
        if err != nil {
            return "", nil, err.Error(), http.StatusInternalServerError
        }

        
        
        // If it is the end, reset the sender's entry in conversation table to beginning of conversation tree
        if endOfTree == 1 {
            
            nodeId = -1
        
            if err := updateSenderCurrentState(senderId, nodeId); err != nil {
                log.Println("Error updating sender")
                return "", nil, err.Error(), http.StatusInternalServerError
            }
            
            
            
        // If it is not the end of the conversation
        } else {

            // Find what is the current node of the response being given
            tNodeId, err := apiutils.InterfaceToInt(rowMapSlice[0]["node_id"])
            if err != nil {
                return "", nil, err.Error(), http.StatusInternalServerError
            } else {
                nodeId = tNodeId
            }
            
            // Update conversation current state with the current node
            err = updateSenderCurrentState(senderId, nodeId)
            if err != nil {
                log.Println("Error updating sender")
                return "", nil, err.Error(), http.StatusInternalServerError
            }
        }
    }
    
    quickReplies, errStr, httpResponse = getQuickReplies(nodeId)
    if httpResponse != 0 {
        return "", nil, errStr, httpResponse
    }
    
    // Return response
    return rowMapSlice[0]["response_text"].(string), quickReplies, "", 0
}

// insertCurrentState inserts a new sender to the conversation table
func insertSenderCurrentState(senderId int64, nodeId int) error {

    // Query to run
    query := fmt.Sprintf(
        "INSERT INTO `%s`.`conversation` " +
        "(`sender_id`,`current_state`) " +
        "VALUES (?,?)", 
        
        "residentportal")
        
    parameters := []interface{}{
        senderId,
        nodeId,
    }
 
    
    // Build query to run in MySQL
    upsertQueries := []apiutils.UpsertQuery{
        {
            Query: query,
            Parameters: parameters,
        },
    }
        
    // Run queries
    getLastInsertId := false
    _, _, errStr, httpResponse := apiutils.RunUpsertQueries(upsertQueries, getLastInsertId)
    if httpResponse != 0 {
        return errors.New(errStr)
    }
    
    return nil
}


// updateSenderCurrentState updates a sender's current state in the conversation table
func updateSenderCurrentState(senderId int64, nodeId int) error {

    // Query to run
    query := fmt.Sprintf("UPDATE `%s`.`conversation` SET `current_state` = ? WHERE `sender_id` = ?", "residentportal") 
    parameters := []interface{}{
        nodeId,
        senderId,
    }
    
    // Build query to run in MySQL
    upsertQueries := []apiutils.UpsertQuery{
        {
            Query: query,
            Parameters: parameters,
        },

    }
    
    // Run queries
    getLastInsertId := false
    _, _, errStr, httpResponse := apiutils.RunUpsertQueries(upsertQueries, getLastInsertId)
    if httpResponse != 0 {
        return errors.New(errStr)
    }
    
    return nil
}


func getQuickReplies(nodeId int) ([]messenger.QuickReply, string, int) {

    // Query to get all quick replies
    query := 
        "SELECT " + 
            "qr.`quick_reply_text`, " +
            "qr.`content_type`, " +
            "qr.`payload` " +
        "FROM " +
            "`quick_reply` qr " +
        "WHERE " +
            "qr.`node_id` = ?"
    
    // Run query in MySQL
    getTotalCount := false
    schema := "residentportal"
    parameters := []interface{}{nodeId}
    rowMapSlice, _, errStr, httpResponse := apiutils.RunSelectQuery(schema, query, parameters, getTotalCount)
    if httpResponse != 0 {
        return nil, errStr, httpResponse
    }

    // If no rows are returned, this is the beginning of the conversation send first message
    if len(rowMapSlice) == 0 {
        return nil, errStr, httpResponse
    }
    
    qrSlice := make([]messenger.QuickReply, len(rowMapSlice))
    
    for i, r := range rowMapSlice {
        qrSlice[i] = messenger.QuickReply{
            ContentType:    r["content_type"].(string),
            Title:          r["quick_reply_text"].(string),
            Payload:        r["payload"].(string),
        }
    }
    
    return qrSlice, "", 0
}
