package responses

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Prettifies input into json string for output
func JSON(input interface{}) string {
	res, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(res)
}

// enum for api response codes
type ResponseCode int
const (
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
)

// Defines Response structure for output
type Response struct {
	Code ResponseCode `json:"code" binding:"required"`
	Message string `json:"message" binding:"required"`
	Data interface{} `json:"data"`
}

// Returns the prettified json string of a properly structure api response given the inputs
func FormatResponse(code ResponseCode, data interface{}, messageDetail string) string {
	var message string
	// Based on code choose base message text
	switch code {
	case -1:
		message = "Unimplemented Feature. You shouldn't be able to hit this on the live build... Please contact developer"
	case 0:
		message = "Generic Failure: Contact Admin"
	case 1:
		message = "Success"
	case 2:
		message = "Token was invalid or missing from request. Did you confirm sending the token as an authorization header?"
	case 3:
		message = "Username failed validation!"
	case 4:
		message = "Failed to save to DB"
	case 5:
		message = "Username passed initial validation but could not generate token, contact Admin."
	case 6:
		message = "Could not get from world DB"
	case 7:
		message = "Could not get from user DB"
	case 8:
		message = "Error while attempting to unmarshal JSON from DB"
	case 9:
		message = "Could not get WDB context from middleware"
	case 10:
		message = "Could not get UDB context from middleware"
	case 11:
		message = "Failed to get AuthPair context from middleware"
	case 12:
		message = "User not found!"
	default:
		message = "Unexpected Error, ResponseCode not in valid enum range!"
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

	return JSON(res)
}

func SendRes(w http.ResponseWriter, code ResponseCode, data interface{}, messageDetail string) {
	fmt.Fprint(w, FormatResponse(code, data, messageDetail))
}