package responses

import "encoding/json"

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
	Generic_Failure ResponseCode = 0
	Generic_Success ResponseCode = 1
	Auth_Failure ResponseCode = 2
	Username_Validation_Failure ResponseCode = 3
	DB_Save_Failure ResponseCode = 4
	Generate_Token_Failure ResponseCode = 5
)

// Defines Response structure for output
type Response struct {
	Data interface{} `json:"data"`
	Message string `json:"message" binding:"required"`
	Code ResponseCode `json:"code" binding:"required"`
}

// Returns the prettified json string of a properly structure api response given the inputs
func FormatResponse(code ResponseCode, data interface{}, messageDetail string) string {
	var message string
	// Based on code choose base message text
	switch code {
	case 0:
		message = "Generic Failure: Contact Admin"
	case 1:
		message = "Success"
	case 2:
		message = "Token was invalid or missing from request. Did you confirm sending the token as an authorization header?"
	case 3:
		message = "Could not claim username, failed validation!"
	case 4:
		message = "Failed to save to DB"
	case 5:
		message = "Username passed validation but could not generate token, contact Admin."
	default:
		message = "Unexpected Error, ResponseCode not in valid enum range!"
	}

	// Define response
	var res Response = Response {
		Data: data,
		Message: message,
		Code: code,
	}

	// If messageDetail provided, append it
	if messageDetail != "" {
		res.Message = message + " | " + messageDetail
	}

	return JSON(res)
}