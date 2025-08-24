package handlers

import (
	"catalogo-virtual-server/internal/entities"
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandler struct {
	dbClient  *dynamodb.Client
	tableName string
}

func NewProductHandler(dbClient *dynamodb.Client, tableName string) *ProductHandler {
	return &ProductHandler{
		dbClient:  dbClient,
		tableName: tableName,
	}
}

func (h *ProductHandler) GetProductsQueryData(c *gin.Context) {
	query := c.Param("query")
	name := c.Query("name")

	// Realiza a Query com FilterExpression
	filterExpression := "contains(OwnerId, :query)"
	expressionAttributeValues := map[string]types.AttributeValue{
		":query": &types.AttributeValueMemberS{Value: query},
	}

	if name != "" {
		filterExpression += " AND contains(ProductName, :name)"
		expressionAttributeValues[":name"] = &types.AttributeValueMemberS{Value: name}
	}

	result, err := h.dbClient.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName:                 aws.String(h.tableName),
		FilterExpression:          aws.String(filterExpression),
		ExpressionAttributeValues: expressionAttributeValues,
	})
	if err != nil {
		log.Fatalf("Erro ao executar Scan: %v", err)
	}

	fmt.Println("tamanho results", len(result.Items))

	// Exibe os resultados
	products := make([]entities.Product, len(result.Items))
	fmt.Println("tamanho array", len(products))
	for idx, item := range result.Items {
		fmt.Println(item)
		product := entities.Product{}

		err := mapToStructClient(item, &product)
		fmt.Println("product", product)
		if err != nil {
			fmt.Println("Erro ao fazer unmarshal:", err)
			return
		}
		products[idx] = product
		fmt.Println("tamanho array", len(products))
	}

	c.JSON(200, gin.H{"data": products})
}

// Função para mapear manualmente o map para a struct
func mapToStructClient(item map[string]types.AttributeValue, response *entities.Product) error {
	for key, value := range item {
		switch key {
		case "id":
			if v, ok := value.(*types.AttributeValueMemberS); ok {
				response.ID = v.Value
			}

		case "ProductName":
			if v, ok := value.(*types.AttributeValueMemberS); ok {
				response.Name = v.Value
			}
		case "Description":
			if v, ok := value.(*types.AttributeValueMemberS); ok {
				response.Description = v.Value
			}
		case "Price":
			if v, ok := value.(*types.AttributeValueMemberS); ok {
				valueFloat := v.Value

				// Converte o valor de string para float64
				// Aqui você pode usar strconv.ParseFloat ou outra função de conversão
				flValue, _ := strconv.ParseFloat(valueFloat, 64)

				response.Price = flValue // Supondo que o preço seja armazenado como string
			}
		case "WhatsappMessage":
			if v, ok := value.(*types.AttributeValueMemberS); ok {
				response.WhatsappMessage = v.Value
			}
		case "Image":
			if v, ok := value.(*types.AttributeValueMemberS); ok {
				response.Image = v.Value
			}
		case "Active":
			if v, ok := value.(*types.AttributeValueMemberS); ok {
				if v.Value == "true" {
					response.IsActive = true
				} else {
					response.IsActive = false
				}
			}
		case "OwnerId":
			if v, ok := value.(*types.AttributeValueMemberS); ok {
				response.OwnerID = v.Value
			}
		}
	}
	return nil
}

// func getProduct

func (h *ProductHandler) PostProductData(c *gin.Context) {
	var inputData entities.Product // Define uma variável para armazenar os dados recebidos

	fmt.Println("PostProductData")
	fmt.Println("body", c.Request.Body)

	// Lê o corpo da requisição (JSON) e vincula os dados à variável inputData
	if err := c.ShouldBindJSON(&inputData); err != nil {
		c.JSON(400, gin.H{"error": "Erro ao analisar dados"})
		return
	}

	// Insere o item na tabela DynamoDB

	input := prepareItemToInput(h.tableName, inputData, true)

	_, err := h.dbClient.PutItem(context.TODO(), input)
	if err != nil {
		log.Printf("Erro ao inserir item no DynamoDB: %v", err)
		c.JSON(500, gin.H{"error": "Erro ao inserir item"})
		return
	}

	c.JSON(201, gin.H{"message": "Item inserido com sucesso"})
}

func prepareItemToInput(tableName string, inputData entities.Product, isActive bool) *dynamodb.PutItemInput {
	var id string = inputData.ID
	if id == "" {
		id = uuid.New().String()
	}
	fmt.Println("id", id)

	priceStr := fmt.Sprintf("%.2f", inputData.Price)

	return &dynamodb.PutItemInput{
		TableName: aws.String(tableName), // Substitua pelo nome da sua tabela
		Item: map[string]types.AttributeValue{
			"id":              &types.AttributeValueMemberS{Value: id},
			"ProductName":     &types.AttributeValueMemberS{Value: inputData.Name},
			"Description":     &types.AttributeValueMemberS{Value: inputData.Description},
			"Price":           &types.AttributeValueMemberS{Value: priceStr},
			"WhatsappMessage": &types.AttributeValueMemberS{Value: inputData.WhatsappMessage},
			"Image":           &types.AttributeValueMemberS{Value: inputData.Image},
			"Active":          &types.AttributeValueMemberS{Value: fmt.Sprintf("%t", isActive)},
			"OwnerId":         &types.AttributeValueMemberS{Value: inputData.OwnerID},
		},
	}
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {

	result, _ := getProduct(h, c)
	if result.Item == nil {
		c.JSON(400, gin.H{"error": "Erro buscar produto"})
		return
	}

	var inputData entities.Product // Define uma variável para armazenar os dados recebidos

	// Lê o corpo da requisição (JSON) e vincula os dados à variável inputData
	if err := c.ShouldBindJSON(&inputData); err != nil {
		c.JSON(400, gin.H{"error": "Erro ao analisar dados"})
		return
	}

	// Insere o item na tabela DynamoDB
	input := prepareItemToInput(h.tableName, inputData, inputData.IsActive)

	_, err := h.dbClient.PutItem(context.TODO(), input)
	if err != nil {
		log.Printf("Erro ao editar item no DynamoDB: %v", err)
		c.JSON(500, gin.H{"error": "Erro ao editar item"})
		return
	}

	c.JSON(201, gin.H{"message": "Item atualizado com sucesso"})
}

func (h *ProductHandler) DeactivateProduct(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		c.JSON(500, gin.H{"error": "Id obrigatório"})
		return
	}

	result, _ := getProduct(h, c)
	if result.Item == nil {
		c.JSON(400, gin.H{"error": "Erro buscar estabelecimento"})
		return
	}

	product := entities.Product{}

	err := mapToStructClient(result.Item, &product)
	if err != nil {
		fmt.Println("Erro ao fazer unmarshal:", err)
		return
	}

	// Insere o item na tabela DynamoDB
	input := prepareItemToInput(h.tableName, product, false)

	_, err = h.dbClient.PutItem(context.TODO(), input)
	if err != nil {
		log.Printf("Erro ao desativar item no DynamoDB: %v", err)
		c.JSON(500, gin.H{"error": "Erro ao desativar item"})
		return
	}

	c.JSON(201, gin.H{"message": "Item desativado com sucesso"})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		c.JSON(500, gin.H{"error": "Id obrigatório"})
		return
	}

	result, _ := getProduct(h, c)
	if result.Item == nil {
		c.JSON(400, gin.H{"error": "Erro buscar estabelecimento"})
		return
	}

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(h.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	}

	_, err := h.dbClient.DeleteItem(context.TODO(), input)
	if err != nil {
		log.Printf("Erro ao deletar item no DynamoDB: %v", err)
		c.JSON(500, gin.H{"error": "Erro ao deletar item"})
		return
	}

	c.JSON(200, gin.H{"message": "Item deletado com sucesso"})
}

func (h *ProductHandler) Ping(c *gin.Context) {
	c.JSON(201, gin.H{"message": "pong"})
}

func getProduct(h *ProductHandler, c *gin.Context) (*dynamodb.GetItemOutput, error) {
	id := c.Param("id")

	input := &dynamodb.GetItemInput{
		TableName: aws.String(h.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	}

	result, err := h.dbClient.GetItem(context.TODO(), input)
	if err != nil {
		log.Printf("Erro ao buscar o item no DynamoDB: %v", err)
		c.JSON(500, gin.H{"error": "Erro ao buscar o item"})
		return nil, err
	}

	if result.Item == nil {
		c.JSON(404, gin.H{"message": "Item não encontrado"})
		return nil, err
	}

	return result, nil
}
