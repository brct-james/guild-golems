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