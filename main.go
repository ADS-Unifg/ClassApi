package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Ra        int                `bson:"ra"`
	Password  string             `bson:"password"`
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

	//uri := os.Getenv("urlMongoDb")
	uri := "mongodb://localhost:27017"

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

	r.POST("/upload", uploadHandler)
	r.GET("/user", serveUserDataHandler)
	r.GET("/photo/:id", servePhotoHandler)
	r.GET("/all_users", serveAllUsersHandler)

	r.GET("/edit_user", func(c *gin.Context) {
		c.HTML(http.StatusOK, "edit_user.html", nil)
	})
	r.POST("/edit_user", editUserHandler)

	r.GET("/delete_user", func(c *gin.Context) {
		c.HTML(http.StatusOK, "delete_user.html", nil)
	})

	r.POST("/delete_user", deleteUserHandler)

	r.Run(":8080")
}
func deleteUserHandler(c *gin.Context) {
	raStr := c.PostForm("ra")
	ra, err := strconv.Atoi(raStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "RA deve ser um número válido"})
		return
	}

	submittedPassword := c.PostForm("password")

	collection := client.Database("user").Collection("userData")

	var existingUser User
	err = collection.FindOne(context.Background(), bson.M{"ra": ra}).Decode(&existingUser)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar o usuário"})
		return
	}

	if existingUser.Password != submittedPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Senha incorreta"})
		return
	}

	result, err := collection.DeleteOne(context.Background(), bson.M{"ra": ra})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao tentar deletar o usuário"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuário deletado com sucesso"})
}

func editUserHandler(c *gin.Context) {

	ra, err := strconv.Atoi(c.PostForm("ra"))
	if err != nil || ra < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "RA deve ser um número válido"})
		return
	}

	submittedPassword := c.PostForm("password")

	collection := client.Database("user").Collection("userData")

	var existingUser User
	err = collection.FindOne(context.Background(), bson.M{"ra": ra}).Decode(&existingUser)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar o usuário"})
		return
	}

	if existingUser.Password != submittedPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Senha incorreta"})
		return
	}

	updatedData := bson.M{
		"$set": bson.M{
			"name":      c.PostForm("name"),
			"apelido":   c.PostForm("apelido"),
			"linkedin":  c.PostForm("linkedin"),
			"github":    c.PostForm("github"),
			"instagram": c.PostForm("instagram"),
		},
	}

	file, _, err := c.Request.FormFile("photo")
	if err == nil {
		defer file.Close()
		photo, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler o arquivo"})
			return
		}
		updatedData["$set"].(bson.M)["photo"] = photo
	}

	_, err = collection.UpdateOne(context.Background(), bson.M{"ra": ra}, updatedData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar o documento"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuário atualizado com sucesso"})
}

func uploadHandler(c *gin.Context) {
	collection := client.Database("user").Collection("userData")

	userCount, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao contar os usuários"})
		return
	}

	if userCount >= 42 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Limite de 42 usuários atingido. Não é possível adicionar mais usuários."})
		return
	}

	ra, err := strconv.Atoi(c.PostForm("ra"))

	if err != nil || ra < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "RA deve ser um número válido"})
		return
	}

	var existingUser User
	err = collection.FindOne(context.Background(), bson.M{"ra": ra}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "RA já está em uso. Escolha outro."})
		return
	} else if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar RA"})
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
		Ra:        ra,
		Password:  c.PostForm("password"),
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
	user.Ra = 0
	user.Password = "sai fora"

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
		users[i].Ra = 0
		users[i].Password = "Sai fora"
	}

	c.JSON(http.StatusOK, users)
}
