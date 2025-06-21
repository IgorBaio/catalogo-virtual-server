package entities

// User represents a user stored in DynamoDB.
type User struct {
	User      string `json:"userId"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	LastLogin string `json:"Last_login"`
}
