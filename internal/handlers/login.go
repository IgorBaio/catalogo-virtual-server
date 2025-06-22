package handlers

import (
	"catalogo-virtual-server/internal/entities"
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
)

// LoginHandler handles login requests against DynamoDB.
type LoginHandler struct {
	dbClient  *dynamodb.Client
	tableName string
}

// NewLoginHandler returns a new LoginHandler.
func NewLoginHandler(dbClient *dynamodb.Client, tableName string) *LoginHandler {
	return &LoginHandler{dbClient: dbClient, tableName: tableName}
}

// Login verifies the credentials and updates Last_login on success.
func (h *LoginHandler) Login(c *gin.Context) {
	var input struct {
		UserId   string `json:"user"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Erro ao analisar dados"})
		return
	}

	// Retrieve the user item by User as key
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(h.tableName),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: input.UserId},
		},
	}

	result, err := h.dbClient.GetItem(context.TODO(), getInput)
	if err != nil {
		log.Printf("Erro ao buscar usuario: %v", err)
		c.JSON(500, gin.H{"error": "Erro ao buscar usuario"})
		return
	}

	if result.Item == nil {
		c.JSON(401, gin.H{"message": "Login invalido"})
		return
	}

	var user entities.User
	mapToUser(result.Item, &user)

	if user.Email == input.Email && user.Password == input.Password {
		// Update Last_login with current time
		_, err := h.dbClient.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
			TableName: aws.String(h.tableName),
			Key: map[string]types.AttributeValue{
				"userId": &types.AttributeValueMemberS{Value: input.UserId},
			},
			ExpressionAttributeNames: map[string]string{
				"#last": "Last_login",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":val": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
			},
			UpdateExpression: aws.String("SET #last = :val"),
		})
		if err != nil {
			log.Printf("Erro ao atualizar usuario: %v", err)
			c.JSON(500, gin.H{"error": "Erro ao atualizar usuario"})
			return
		}
		c.JSON(200, gin.H{"message": "Login realizado com sucesso", "company": user.Company})
		return
	}

	c.JSON(401, gin.H{"message": "Login invalido"})
}

// mapToUser maps DynamoDB attributes to the User struct.
func mapToUser(item map[string]types.AttributeValue, user *entities.User) {
	if v, ok := item["userId"].(*types.AttributeValueMemberS); ok {
		user.User = v.Value
	}
	if v, ok := item["Email"].(*types.AttributeValueMemberS); ok {
		user.Email = v.Value
	}
	if v, ok := item["Password"].(*types.AttributeValueMemberS); ok {
		user.Password = v.Value
	}
	if v, ok := item["Company"].(*types.AttributeValueMemberS); ok {
		user.Company = v.Value
	}
	if v, ok := item["Last_login"].(*types.AttributeValueMemberS); ok {
		user.LastLogin = v.Value
	}
}
