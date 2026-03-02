package todo

// สำหรับใครต้องการอ่านเพิ่มเติมเกี่ยวกับ Composition Over Inheritance สามารถอ่านได้ที่ https://dev.to/pallat/composition-over-inheritance-in-go-5b7b

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Title is required, Completed is optional and defaults to false if not provided in the JSON payload when creating a new task. This allows clients to create tasks without explicitly setting the completed status, and it will be treated as incomplete by default.
// so tag can do a value validation and set default value for the field if it's not provided in the JSON payload.
type Todo struct {
	gorm.Model
	Text      string `json:"text" binding:"required"`
	Completed bool   `json:"completed" gorm:"default:false"`
}

// TableName specifies the table name for the Todo model in the database,
// in gorm, by default, the table name is the pluralized form of the struct name (e.g., "todos" for "Todo"),
func (t *Todo) TableName() string {
	return "todoslist"
}

// set type for todohandler
type TodoHandler struct {
	DB *gorm.DB
}

// NewTodoHandler is a constructor function that initializes and returns a new instance of TodoHandler
// with the provided database connection.
func NewTodoHandler(db *gorm.DB) *TodoHandler {
	return &TodoHandler{DB: db}
}

func (h *TodoHandler) NewTask(c *gin.Context) {

	// *** if use middleware, we no need to set the protect at the handler anymore, because the middleware will handle the authentication for us, and we can just focus on the business logic of creating a new task in this handler.
	// s := c.Request.Header.Get("Authorization")
	// tokenString := s[len("Bearer "):] // Extract the token string from the Authorization header by removing the "Bearer " prefix.

	// if err := auth.Protect(tokenString); err != nil { // Call the Protect function from the auth package to validate the token.
	// 	c.AbortWithStatus(http.StatusUnauthorized) // we use a abort to stop the request from being processed further and return a 401 Unauthorized status code if the token is invalid or expired.
	// 	// c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"}) // This line can use also instead of c.AbortWithStatus, for healthier
	// 	return // use abort when you want to stop the of handler and return a response immediately, while use return when you want to stop the execution of the current function and return to the caller. In this case, we use both because we want to stop the handler from processing further and also return a response to the client.
	// }

	var todo Todo
	// 1. we can use BindJSON to bind the incoming JSON payload to the Todo struct.
	// 2. we can use ShouldBindJSON to bind the incoming JSON payload to the Todo struct.
	// 3. if use BindJSON, it will return an error if the JSON is invalid, but ShouldBindJSON will not return an error
	// and will instead set the fields to your own formatted error message.

	// 4. ShouldBindJSON modify todo direclty, so we don't need to return the modified todo struct, we can just use it directly after binding.
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// logging for easy investigation
	if todo.Text == "sleep" {
		// we can use c.Request.Header.Get to get the value and tracking the transaction id in the log, so we can easily investigate the issue if we have the transaction id in the log.
		transactionId := c.Request.Header.Get("Transaction-ID")
		// get audience from the context that we set in the Protect middleware, so we can also track the audience in the log for better investigation.
		aud, exists := c.Get("aud")
		if !exists {
			log.Println(transactionId, "audience not found in context")
		} else {
			log.Println(transactionId, "audience:", aud)
		}
		log.Println(transactionId, aud, "not allowed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "not allowed"})
		return
	}

	r := h.DB.Create(&todo)
	if err := r.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// we can use todo.Model.ID to get the ID of the newly created task, and we can return it in the response along with the task details.
	c.JSON(http.StatusCreated, gin.H{"message": "Task created successfully", "id": todo.ID, "task": todo})
}


func (h *TodoHandler) List(c *gin.Context) {
	var todos []Todo
	r := h.DB.Find(&todos)
	if err := r.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, todos)
}

func (h *TodoHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r := h.DB.Delete(&Todo{}, id) // delete from which table and which id, we can also use map to delete with condition like h.DB.Delete(&Todo{}, "text = ?", "sleep") to delete all task with text sleep
	if err := r.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}