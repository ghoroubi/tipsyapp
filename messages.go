package main

var Messages map[string]string

func InitMessages() {
	Messages = make(map[string]string)

	Messages["InternalServerError"] = "Internal server error occured!"
	// Messages["DuplicateData"] = "Duplicate data!"
	// Messages["NotFound"] = "Resource not found!"
	// Messages["NotValid"] = "Invalid input"
	// Messages["ResourceDeleted"] = "Resource deleted successfully!"
	// Messages["ResourceUpdated"] = "Resource updated successfully!"
	// Messages["ResourceCreated"] = "Resource created successfully!"
	// Messages["RequiredPassword"] = "Password is required!"
	// Messages["PasswordFormatFailed"] = "Password format failed!"
	// Messages["AgentIsNotActive"] = "Agent is not activated!"
	// Messages["PasswordError"] = "Password is not correct!"
	// Messages["InvalidToken"] = "Invalid token!"
	// Messages["BadRequest"] = "Bad request"
	// Messages["NoDbSpace"] = "dbspace not found in id!"
	// Messages["IndexError"] = `already exists`
	// Messages["DuplicateIndex"] = "pq: duplicate key value violates unique constraint"
	// Messages["PasswordFormat"] = "^[a-z.A-Z.0-9]{6,}$"
	Messages["ShortForm"] = "2006-01-02"
	Messages["LongForm"] = "2006-01-02T15:04:05-07:00"

	Messages["ChangePasswordSubject"] = "Reset your password"
	Messages["ChangePasswordBody"] = "Changing your password is simple. Please use the code below \r\n %s"
	// Messages["Forbidden"] = "You are not authorized for access to this resource"
	// Messages["ExpectedPointer"] = "Expected pointer value!"
	// Messages["RateExceed"] = "Rate of request exceeds server policy"
	// Messages["ManyItems"] = "Too many items!"
	// Messages["JsonFormatError"] = "Incorrect Json Format!"
	// Messages["DbNotFoundError"] = "sql: no rows in result set"
	Messages["ActiveUserSubject"] = "Confirm your email address"
	Messages["ActiveUserBody"] = "You recently entered %s email address at our application.\r\n Please confirm your email by entering the code below at the application.\r\n %s"
	Messages["NewRequest"] = "%s shows interest to join you!"
	Messages["NewMessage"] = "You received a new message from %s!"

}
