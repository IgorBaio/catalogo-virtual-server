package handlers

import (
	"catalogo-virtual-server/internal/entities"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

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
	ctx := c.Request.Context()

	ownerID := c.Param("query") // supondo que seja o OwnerId exato
	if ownerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query (OwnerId) é obrigatório"})
		return
	}
	name := c.Query("name")

	fmt.Println("Name", name)

	// Monta KeyCondition: OwnerId = :oid
	exprAttrNames := map[string]string{
		"#oid": "OwnerId",
	}
	exprAttrValues := map[string]types.AttributeValue{
		":oid": &types.AttributeValueMemberS{Value: ownerID},
	}

	keyCond := "#oid = :oid"

	// Filtro opcional por Name (palavra reservada -> usar alias)
	var filterExpr *string
	if name != "" {
		exprAttrNames["#nm"] = "ProductNameLower"
		exprAttrValues[":name"] = &types.AttributeValueMemberS{Value: name}
		fe := "contains(#nm, :name)"
		filterExpr = aws.String(fe)
	}

	fmt.Println("FilterExpression", filterExpr)

	// Paginação: acumula todos os itens (ou adapte p/ limit/offset)
	var products []entities.Product
	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		out, err := h.dbClient.Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String(h.tableName),
			IndexName:                 aws.String("OwnerId"), // ajuste para o nome real do seu GSI
			KeyConditionExpression:    aws.String(keyCond),
			FilterExpression:          filterExpr, // nil se não houver filtro por name
			ExpressionAttributeNames:  exprAttrNames,
			ExpressionAttributeValues: exprAttrValues,
			ExclusiveStartKey:         lastEvaluatedKey,
			// ProjectionExpression:    aws.String("#oid, #nm, ...") // opcional
		})
		if err != nil {
			// retorna HTTP 500 em vez de derrubar o processo
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao executar Query: %v", err)})
			return
		}

		fmt.Println("tamanho results", len(out.Items))

		// Transforma Items -> struct
		for _, item := range out.Items {
			var p entities.Product
			if err := mapToStructClient(item, &p); err != nil {
				// apenas loga e segue, ou retorne 500 conforme sua política
				log.Printf("unmarshal product: %v", err)
				continue
			}
			products = append(products, p)
		}

		if out.LastEvaluatedKey == nil || len(out.LastEvaluatedKey) == 0 {
			break
		}
		lastEvaluatedKey = out.LastEvaluatedKey
	}

	c.JSON(http.StatusOK, gin.H{"data": products})
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
		case "Category":
			if v, ok := value.(*types.AttributeValueMemberS); ok {
				response.Category = v.Value
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
			"id":               &types.AttributeValueMemberS{Value: id},
			"ProductName":      &types.AttributeValueMemberS{Value: inputData.Name},
			"ProductNameLower": &types.AttributeValueMemberS{Value: strings.ToLower(inputData.Name)},
			"Description":      &types.AttributeValueMemberS{Value: inputData.Description},
			"Price":            &types.AttributeValueMemberS{Value: priceStr},
			"Category":         &types.AttributeValueMemberS{Value: inputData.Category},
			"Image":            &types.AttributeValueMemberS{Value: inputData.Image},
			"Active":           &types.AttributeValueMemberS{Value: fmt.Sprintf("%t", isActive)},
			"OwnerId":          &types.AttributeValueMemberS{Value: inputData.OwnerID},
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
