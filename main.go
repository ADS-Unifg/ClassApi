package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Idaluno   int                `bson:"idaluno"`
	Name      string             `bson:"name"`
	Apelido   string             `bson:"apelido"`
	Linkedin  string             `bson:"linkedin"`
	Github    string             `bson:"github"`
	Instagram string             `bson:"instagram"`
	Photo     []byte             `bson:"photo"`
}

var client *mongo.Client

func init() {
	var err error

	uri := os.Getenv("urlMongoDb")
	//uri := "mongodb://localhost:27017"

	client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(uri))

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := gin.Default()

	// Middleware para CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	r.LoadHTMLGlob("templates/*.html")

	r.Static("/public", "./public")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "form.html", nil)
	})
	r.GET("/view_user", func(c *gin.Context) {
		c.HTML(http.StatusOK, "view_user.html", nil)
	})

	r.POST("/upload", uploadHandler)
	r.GET("/user", serveUserDataHandler)
	r.GET("/photo/:id", servePhotoHandler)
	r.GET("/all_users", serveAllUsersHandler)

	r.Run(":8080")
}
func uploadHandler(c *gin.Context) {
	idaluno, err := strconv.Atoi(c.PostForm("idaluno"))
	if err != nil || idaluno < 1 || idaluno > 40 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDAluno deve ser um número entre 1 e 45"})
		return
	}

	collection := client.Database("user").Collection("userData")
	var existingUser User
	err = collection.FindOne(context.Background(), bson.M{"idaluno": idaluno}).Decode(&existingUser)
	if err == nil {

		c.JSON(http.StatusConflict, gin.H{"error": "IDAluno já está em uso. Escolha outro."})
		return
	} else if err != mongo.ErrNoDocuments {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar IDAluno"})
		return
	}

	file, _, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao carregar o arquivo"})
		return
	}
	defer file.Close()

	photo, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler o arquivo"})
		return
	}

	user := User{
		Idaluno:   idaluno,
		Name:      c.PostForm("name"),
		Apelido:   c.PostForm("apelido"),
		Linkedin:  c.PostForm("linkedin"),
		Github:    c.PostForm("github"),
		Instagram: c.PostForm("instagram"),
		Photo:     photo,
	}

	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao inserir o documento"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Arquivo enviado com sucesso", "id": result.InsertedID})
}

func serveUserDataHandler(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	collection := client.Database("user").Collection("userData")
	var user User
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	photoURL := "/photo/" + id
	user.Photo = nil

	c.JSON(http.StatusOK, gin.H{
		"user":     user,
		"photoURL": photoURL,
	})
}

func servePhotoHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	collection := client.Database("user").Collection("userData")
	var user User
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", user.Photo)
}

func serveAllUsersHandler(c *gin.Context) {
	collection := client.Database("user").Collection("userData")

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	defer cursor.Close(context.Background())

	var users []User
	if err = cursor.All(context.Background(), &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode users"})
		return
	}

	for i := range users {
		users[i].Photo = nil
	}

	c.JSON(http.StatusOK, users)
}
