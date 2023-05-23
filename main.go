package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Book struct {
	BookID     string `json:"book_id" bson:"book_id"`
	BookTitle  string `json:"book_title" bson:"book_title"`
	BookAuthor string `json:"book_author" bson:"book_author"`
}

type UpdateBook struct {
	BookTitle  string `json:"book_title" bson:"book_title,omitempty"`
	BookAuthor string `json:"book_author" bson:"book_author,omitempty"`
}

var Collection *mongo.Collection

func init() {

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI("mongodb+srv://ep5-course:HlT9NpyD4Vt0HtbK@cluster0.vvx397a.mongodb.net/?retryWrites=true&w=majority")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set the collection
	Collection = client.Database("ep5_course").Collection("books")
}

func main() {

	app := fiber.New()

	app.Get("/api/books", getBooks)
	app.Get("/api/books/:id", getBook)
	app.Post("/api/books", createBook)
	app.Put("/api/books/:id", updateBook)
	app.Delete("/api/books/:id", deleteBook)

	app.Use("*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
			"code":    fiber.StatusMethodNotAllowed,
			"status":  false,
			"message": "Method Not Allowed",
		})
	})

	log.Fatal(app.Listen(":3000"))
}

func getBooks(c *fiber.Ctx) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find all books in the collection
	cursor, err := Collection.Find(ctx, bson.D{})
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	result := []Book{}
	// Deserialize the results into a slice of books
	if err := cursor.All(ctx, &result); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"status":  true,
		"message": "get books success",
		"data":    result,
	})
}

func getBook(c *fiber.Ctx) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get the book ID from the URL parameter
	bookID := c.Params("id")

	result := Book{}

	// Find the book with the given ID
	filter := bson.M{"book_id": bookID}
	err := Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"status":  true,
		"message": "get book success",
		"data":    result,
	})
}

func createBook(c *fiber.Ctx) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Parse the request body into a new book instance
	book := Book{}
	if err := c.BodyParser(&book); err != nil {
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    fiber.StatusInternalServerError,
				"status":  false,
				"message": "unexpected",
			})
		}
	}

	//Find duplicate title in db
	filter := bson.M{"book_title": book.BookTitle}
	cursor, err := Collection.Find(ctx, filter)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	findBooks := []Book{}
	if err := cursor.All(ctx, &findBooks); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	if len(findBooks) > 0 {
		log.Println("Title is duplicated")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	// Generate book idwitht google uuid
	book.BookID = uuid.NewString()

	// Insert the new book into the collection
	res, err := Collection.InsertOne(ctx, book)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	result := Book{}
	// Find the book with the given ID
	filter = bson.M{"_id": res.InsertedID}
	err = Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"status":  true,
		"message": "create book success",
		"data":    result,
	})
}

func updateBook(c *fiber.Ctx) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get the book ID from the URL parameter
	bookID := c.Params("id")

	// Parse the request body into an updated book instance
	updatedBook := UpdateBook{}
	if err := c.BodyParser(&updatedBook); err != nil {
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    fiber.StatusInternalServerError,
				"status":  false,
				"message": "unexpected",
			})
		}
	}

	//Find duplicate title in db
	filter := bson.M{"book_title": updatedBook.BookTitle}
	cursor, err := Collection.Find(ctx, filter)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	findBooks := []Book{}
	if err := cursor.All(ctx, &findBooks); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	if len(findBooks) > 0 {
		log.Println("Title is duplicated")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	// Update the book with the given ID
	filter = bson.M{"book_id": bookID}
	update := bson.M{"$set": updatedBook}
	res, err := Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	if res.ModifiedCount == 0 {
		log.Println("not update document")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	result := Book{}
	// Find the book with the given ID
	filter = bson.M{"book_id": bookID}
	err = Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"status":  true,
		"message": "update book success",
		"data":    result,
	})
}

func deleteBook(c *fiber.Ctx) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get the book ID from the URL parameter
	bookID := c.Params("id")

	// Delete the book with the given ID
	filter := bson.M{"book_id": bookID}
	res, err := Collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	if res.DeletedCount == 0 {
		log.Println("not delete document")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"status":  false,
			"message": "unexpected",
		})
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"status":  true,
		"message": "delete book success",
	})
}
