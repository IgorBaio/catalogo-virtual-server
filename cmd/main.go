package main

import (
	"catalogo-virtual-server/internal/handlers"
	"catalogo-virtual-server/internal/utils"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	accessKey := utils.GetEnvVar("AccessKeyID")
	secretAccessKey := utils.GetEnvVar("SecretAccessKey")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"), // Substitua pela região do seu DynamoDB
		config.WithCredentialsProvider(aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretAccessKey, ""))))
	if err != nil {
		log.Fatalf("Erro ao carregar a configuração: %v\n", err)
	}

	// Cria o cliente do DynamoDB
	client := dynamodb.NewFromConfig(cfg)

	// Teste a conexão listando as tabelas
	output, err := client.ListTables(context.TODO(), &dynamodb.ListTablesInput{})
	if err != nil {
		log.Fatalf("Erro ao listar tabelas: %v\n", err)
	}

	fmt.Println("Tabelas disponíveis:", output.TableNames)

	// Injetar o cliente no handler
	productHandler := handlers.NewProductHandler(client, "catalog-products")
	loginHandler := handlers.NewLoginHandler(client, "users-catalogo")
	// auth := utils.NewAuth()

	// Configurar o servidor Gin
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://igorbaio.github.io"},                   // Domínios permitidos
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Métodos permitidos
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}, // Headers permitidos
		ExposeHeaders:    []string{"Content-Length", "Authorization"},         // Headers expostos
		AllowCredentials: true,                                                // Permitir cookies
		MaxAge:           12 * time.Hour,                                      // Cache da configuração de CORS
	}))

	// router.POST("/login", func(c *gin.Context) {
	// 	handlers.LoginHandler(c, auth, userHandler)
	// })

	// protected := router.Group("/protected")
	// protected.Use(middleware.AuthMiddleware(auth))
	// protected.GET("", handlers.ProtectedHandler)
	// protected.POST("/admin/create/user", func(c *gin.Context) {
	// 	userHandler.CreateUserByAdmin(c, auth)
	// })

	// Definir as rotas GET e POST
	router.POST("/login", loginHandler.Login)
	router.GET("/ping", productHandler.Ping)
	// router.GET("/api/produto/:id", productHandler.GetClientData)
	router.GET("/api/produto/query/:query", productHandler.GetProductsQueryData)
	// router.GET("/api/produtos", productHandler.GetClientsData)
	router.POST("/api/produto", productHandler.PostProductData)
	router.PUT("/api/produto/:id", productHandler.UpdateProduct)
	router.DELETE("/api/produto/:id", productHandler.DeactivateProduct)
	router.DELETE("/api/produto/delete/:id", productHandler.DeleteProduct)
	// router.GET("/validateJWT", func(c *gin.Context) {

	// 	userHandler.ValidateJWTORCreate(c, auth)
	// })

	port := utils.GetEnvVar("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Servidor rodando em http://localhost:"+port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
