package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User Struct Model
type User struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
}

// Post Struct Model
type Post struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID          string             `json:"userid,omitempty" bson:"userid,omitempty"`
	Caption         string             `json:"caption,omitempty" bson:"caption,omitempty"`
	ImageURL        string             `json:"imageurl,omitempty" bson:"imageurl,omitempty"`
	PostedTimeStamp JSONTime           `json:"postedtimestamp,omitempty" bson:"postedtimestamp,omitempty"`
}

// Payload Struct Model [For GetPostByUserIDEndpoint]
type Payload struct {
	Posts    []Post `json:"posts,omitempty" bson:"posts,omitempty"`
	Total    int64  `json:"total,omitempty" bson:"total,omitempty"`
	Page     int    `json:"page,omitempty" bson:"page,omitempty"`
	LastPage int    `json:"lastpage,omitempty" bson:"lastpage,omitempty"`
}

// Time Struct Model [For PostedTimeStamp]
type JSONTime struct {
	time.Time
}

// Function to return Day, Month & Date
func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", t.Format("Mon Jan _2"))
	return []byte(stamp), nil
}

// Function to Check Email Validity
func valid_email(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// Function to Hash Password using MD5
func GetHashedPassword(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

var client *mongo.Client

// POST: CreateUser Function
func CreateUserEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")

	var newUser User

	json.NewDecoder(request.Body).Decode(&newUser)

	usersCollection := client.Database("DataBase").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	cursor, _ := usersCollection.Find(ctx, bson.M{})

	// Return if Email is Invalid
	if !valid_email(newUser.Email) {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(`{"message": "ERROR!!! Please Enter Proper Email ID"}`))
		defer cancel()
		return
	}

	// Check if each User is already in DB
	for cursor.Next(ctx) {
		checkDuplicateUser := User{}
		cursor.Decode(&checkDuplicateUser)

		// Return if Email Already Exists in DB
		if checkDuplicateUser.Email == newUser.Email {
			response.WriteHeader(http.StatusConflict)
			response.Write([]byte(`{"message": "ERROR!!! Email ID has been already used. Please Use Different Email"}`))
			defer cancel()
			return
		}

	}

	// Storing the Hashed Password
	newUser.Password = GetHashedPassword(newUser.Password)

	result, _ := usersCollection.InsertOne(ctx, newUser)
	response.WriteHeader(http.StatusCreated)
	json.NewEncoder(response).Encode(result)

	// Prints the Inserted User ID
	fmt.Println(result.InsertedID)
	defer cancel()
}

// POST: CreateUserPost Function
func CreatePostEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")

	var newPost Post

	json.NewDecoder(request.Body).Decode(&newPost)

	postsCollection := client.Database("DataBase").Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// Get Current Time in Custom Format and Set it to Post
	currTime := JSONTime{time.Now()}
	newPost.PostedTimeStamp = currTime

	result, _ := postsCollection.InsertOne(ctx, newPost)
	response.WriteHeader(http.StatusCreated)
	json.NewEncoder(response).Encode(result)

	// Prints the Inserted Post ID
	fmt.Println(result.InsertedID)
	defer cancel()
}

// GET: GetUserByID Function
func GetUserByIDEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")

	// Get the userID from RequestURL Path
	userID := strings.Replace(request.URL.Path, "/users/", "", 1)

	id, _ := primitive.ObjectIDFromHex(userID)

	var user User

	usersCollection := client.Database("DataBase").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// Check if User Exists with Given userID, If Not Return
	err := usersCollection.FindOne(ctx, User{ID: id}).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusNotFound)
		response.Write([]byte(`{"message": "ERROR!!! User NOT Found"}`))
		defer cancel()
		return
	}

	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(user)
	defer cancel()
}

// GET: GetPostByID Function
func GetPostByIDEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")

	// Get the postID from RequestURL Path
	postID := strings.Replace(request.URL.Path, "/posts/", "", 1)

	id, _ := primitive.ObjectIDFromHex(postID)

	var post Post

	postsCollection := client.Database("DataBase").Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// Check if Post Exists with Given postID, If Not Return
	err := postsCollection.FindOne(ctx, User{ID: id}).Decode(&post)
	if err != nil {
		response.WriteHeader(http.StatusNotFound)
		response.Write([]byte(`{"message": "ERROR!!! Post NOT Found"}`))
		defer cancel()
		return
	}

	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(post)
	defer cancel()
}

// GET: GetPostByUserID Function
func GetPostsByUserIDEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")

	var posts []Post

	postsCollection := client.Database("DataBase").Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// Get userID from the RequestURL Path
	userID := strings.Replace(request.URL.Path, "/posts/users/", "", 1)

	// Set the Filter to userID
	filter := bson.M{"userid": userID}

	// Initialize findOptions
	findOptions := options.Find()

	// API Pagination
	// Get the Page Param from RequestURL Query, and Store it as Int
	page, _ := strconv.Atoi(request.URL.Query().Get("page"))

	// If no Page Param is provided, Set Page to Default 1
	if page == 0 {
		page = 1
	}

	// Sets the Number of Results per Page,
	// For Simplicity, Its set as 2, can be larger > 2
	var perPage int64 = 2

	// Finds the Total Number of Posts that Exist for the given userID
	total, _ := postsCollection.CountDocuments(ctx, filter)

	// Finds the Last Page Number from the Total
	lastpage := int(math.Ceil(float64(total / perPage)))

	// findOptions to SetSkip value and SetLimit
	findOptions.SetSkip((int64(page) - 1) * perPage)
	findOptions.SetLimit(perPage)

	// Pass in all the Parameters to the Find,
	filterCursor, err := postsCollection.Find(ctx, filter, findOptions)
	if err != nil {
		response.WriteHeader(http.StatusNotFound)
		response.Write([]byte(`{"message": "ERROR!!! Posts NOT Found for given UserID"}`))
		defer cancel()
		return
	}

	defer filterCursor.Close(ctx)

	// Appending each Resultant Post to Posts
	for filterCursor.Next(ctx) {
		var post Post
		filterCursor.Decode(&post)
		posts = append(posts, post)
	}

	response.WriteHeader(http.StatusOK)

	// To add Addition Payload Values
	payload := Payload{posts, total, page, lastpage}
	json.NewEncoder(response).Encode(payload)
	defer cancel()
}

func main() {
	fmt.Println("Application has started")

	// Set Context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// Set Client Options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, _ = mongo.Connect(ctx, clientOptions)

	// Check the Connection
	err := client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	handler := http.NewServeMux()

	// HTTP Routing
	handler.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			handler.HandleFunc("/posts/users/", GetPostsByUserIDEndpoint)
			handler.HandleFunc("/posts/", GetPostByIDEndpoint)
			handler.HandleFunc("/users/", GetUserByIDEndpoint)
		case http.MethodPost:
			handler.HandleFunc("/users", CreateUserEndpoint)
			handler.HandleFunc("/posts", CreatePostEndpoint)
		default:
			http.Error(response, "Method Not Allowed Try Again!", http.StatusMethodNotAllowed)
		}
	})

	http.ListenAndServe(":8080", handler)
	defer cancel()
}
