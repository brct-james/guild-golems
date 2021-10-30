package responses

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/brct-james/guild-golems/log"
)

// Prettifies input into json string for output
func JSON(input interface{}) (string, error) {
	res, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", err
	}
	return string(res), nil
}

// enum for api response codes
type ResponseCode int
const (
	CRITICAL_JSON_MARSHAL_ERROR ResponseCode = -3
	JSON_Marshal_Error ResponseCode = -2
	Unimplemented ResponseCode = -1
	Generic_Failure ResponseCode = 0
	Generic_Success ResponseCode = 1
	Auth_Failure ResponseCode = 2
	Username_Validation_Failure ResponseCode = 3
	DB_Save_Failure ResponseCode = 4
	Generate_Token_Failure ResponseCode = 5
	WDB_Get_Failure ResponseCode = 6
	UDB_Get_Failure ResponseCode = 7
	JSON_Unmarshal_Error ResponseCode = 8
	No_WDB_Context ResponseCode = 9
	No_UDB_Context ResponseCode = 10
	No_AuthPair_Context ResponseCode = 11
	User_Not_Found ResponseCode = 12
	Not_Enough_Mana ResponseCode = 13
	No_Such_Ritual ResponseCode = 14
	No_Golem_Found ResponseCode = 15
	Ritual_Not_Known ResponseCode = 16
	No_Such_Status ResponseCode = 17
	Golem_In_Blocking_Status ResponseCode = 18
	New_Status_Not_Allowed ResponseCode = 19
	Bad_Request ResponseCode = 20
	No_Available_Routes ResponseCode = 21
	Target_Route_Unavailable ResponseCode = 22
	UDB_Update_Failed ResponseCode = 23
	Leaderboard_Not_Found ResponseCode = 24
	No_Resource_Nodes_At_Location ResponseCode = 25
	Target_Resource_Node_Unavailable ResponseCode = 26
	No_Packable_Items ResponseCode = 27
	Invalid_Manifest ResponseCode = 28
	Manifest_Overflow ResponseCode = 29
	No_Storable_Items ResponseCode = 30
	Blank_Manifest_Disallowed ResponseCode = 31
	Blank_Order_Disallowed ResponseCode = 32
	Merchant_Inventory_Empty ResponseCode = 33
	Invalid_Order_Type ResponseCode = 34
	Insufficient_Resources_Held ResponseCode = 35
	Clearinghouse_Spool_Error ResponseCode = 36
	Could_Not_Decode_Order ResponseCode = 37
)

// Defines Response structure for output
type Response struct {
	Code ResponseCode `json:"code" binding:"required"`
	Message string `json:"message" binding:"required"`
	Data interface{} `json:"data"`
}

// Returns the prettified json string of a properly structure api response given the inputs
func FormatResponse(code ResponseCode, data interface{}, messageDetail string) (string, error) {
	var message string
	// Based on code choose base message text
	switch code {
	case -3:
		message = "[CRITICAL_JSON_MARSHAL_ERROR] Server error in responses.JSON, could not marshal JSON_Marshal_Error response! PLEASE contact developer."
	case -2:
		message = "[JSON_Marshal_Error] Responses module encountered an error while marshaling response JSON. Please contact developer."
	case -1:
		message = "[Unimplemented] Unimplemented Feature. You shouldn't be able to hit this on the live build... Please contact developer"
	case 0:
		message = "[Generic_Failure] Contact developer"
	case 1:
		message = "[Generic_Success] Request Successful"
	case 2:
		message = "[Auth_Failure] Token was invalid or missing from request. Did you confirm sending the token as an authorization header?"
	case 3:
		message = "[Username_Validation_Failure] Please ensure username conforms to requirements and account does not already exist!"
	case 4:
		message = "[DB_Save_Failure] Failed to save to DB"
	case 5:
		message = "[Generate_Token_Failure] Username passed initial validation but could not generate token, contact Admin."
	case 6:
		message = "[WDB_Get_Failure] Could not get from world DB"
	case 7:
		message = "[UDB_Get_Failure] Could not get from user DB"
	case 8:
		message = "[JSON_Unmarshal_Error] Error while attempting to unmarshal JSON from DB"
	case 9:
		message = "[No_WDB_Context] Could not get WDB context from middleware"
	case 10:
		message = "[No_UDB_Context] Could not get UDB context from middleware"
	case 11:
		message = "[No_AuthPair_Context] Failed to get AuthPair context from middleware"
	case 12:
		message = "[User_Not_Found] User not found!"
	case 13:
		message = "[Not_Enough_Mana] Could not complete requested action due to insufficient mana"
	case 14:
		message = "[No_Such_Ritual] The specified ritual is not recognized"
	case 15:
		message = "[No_Golem_Found] Golem with the specified symbol could not be found in user data"
	case 16:
		message = "[Ritual_Not_Known] User does not know the specified ritual, so it cannot be executed"
	case 17:
		message = "[No_Such_Status] Specified status does not exist"
	case 18:
		message = "[Golem_In_Blocking_Status] Golem's current status does not allow changes to be made"
	case 19:
		message = "[New_Status_Not_Allowed] Specified status is not valid for the specified golem's archetype"
	case 20:
		message = "[Bad_Request] Invalid request payload, please validate the request body conforms to expectations"
	case 21:
		message = "[No_Available_Routes] No routes available for the specified golem, contact developer"
	case 22:
		message = "[Target_Route_Unavailable] The specified route is not available at the current location"
	case 23:
		message = "[UDB_Update_Failed] Could not complete request due to error while saving user data to udb"
	case 24:
		message = "[Leaderboard_Not_Found] Requested leaderboard not found"
	case 25:
		message = "[No_Resource_Nodes_At_Location] No resources nodes found at the location of the specified golem"
	case 26:
		message = "[Target_Resource_Node_Unavailable] The specified resource node is not available at the current location"
	case 27:
		message = "[No_Packable_Items] No packable items in location inventory at specified golem's locale"
	case 28:
		message = "[Invalid_Manifest] Invalid manifest, specified item not contained in sufficient quantity in specified inventory"
	case 29:
		message = "[Manifest_Overflow] Manifest is valid but requests more items than the golem can handle"
	case 30:
		message = "[No_Storable_Items] No storable items in the specified golem's inventory"
	case 31:
		message = "[Blank_Manifest_Disallowed] Manifest cannot be blank, please includes items to load"
	case 32:
		message = "[Blank_Order_Disallowed] Order cannot be blank, please includes order type, item symbol, quantity, target price, and force_execution"
	case 33:
		message = "[Merchant_Inventory_Empty] Merchant must be holding the items you wish to sell"
	case 34:
		message = "[Invalid_Order_Type] Type of the specified order does not match any known type"
	case 35:
		message = "[Insufficient_Resources_Held] The specified action could not be completed due to insufficient resources in golem inventory"
	case 36:
		message = "[Clearinghouse_Spool_Error] Order incorrectly spooled by clearinhouse, could not execute when asked"
	case 37:
		message = "[Could_Not_Decode_Order] Error occurred while decoding order, ensure formatting is correct"
	default:
		message = "[Unexpected_Error] ResponseCode not in valid enum range! Contact developer"
	}

	// Define response
	var res Response = Response {
		Code: code,
		Message: message,
		Data: data,
	}

	// If messageDetail provided, append it
	if messageDetail != "" {
		res.Message = message + " | " + messageDetail
	}

	responseText, jsonErr := JSON(res)
	if jsonErr != nil {
		return "", jsonErr
	}
	return responseText, nil
}

func SendRes(w http.ResponseWriter, code ResponseCode, data interface{}, messageDetail string) {
	responseObject, jsonErr := FormatResponse(code, data, messageDetail)
	if jsonErr != nil {
		jsonErrMsg := fmt.Sprintf("Could not MarshallIndent json for data %v", data)
		errResponseObject, criticalJsonError := FormatResponse(JSON_Marshal_Error, nil, jsonErrMsg)
		if criticalJsonError != nil {
			log.Error.Printf("Could not format MarshallIndent response, error: %v", criticalJsonError)
			fmt.Fprintf(w, "{\"code\":-3, \"message\": \"CRITICAL SERVER ERROR in responses.JSON, could not marshal JSON_Marshal_Error response! PLEASE contact developer. Error: %v\", \"data\":{}", criticalJsonError)
		}
		fmt.Fprint(w, errResponseObject)
	}
	fmt.Fprint(w, responseObject)
}