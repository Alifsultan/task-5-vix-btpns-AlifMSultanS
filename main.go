package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kataras/go-sessions"
	"golang.org/x/crypto/bcrypt"
)

// deklarasi variable db end errors
var db *sql.DB
var err error

// var currentImage *imageupload.Image

// deklarasi variabel untuk componen users
type user struct {
	ID       int
	Username string
	Email    string
	Password string
	level    string
	Avatar   *multipart.FileHeader `form:"avatar" binding:"required"`
}

// membuat function koneksi ke database mysql
func connect_db() {
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1)/go_db")
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}
}

// membuat routes untuk alamat URL pada browser
func routes() {
	http.HandleFunc("/", home)
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
}

// fungsi utama yang akan di eksekusi
func main() {
	// conect db
	connect_db()
	// connect router untuk url
	routes()

	defer db.Close()
	//server aktif pada port :8000
	fmt.Println("Server is Actived in port :8000")
	http.ListenAndServe(":8000", nil)
}

func checkErr(w http.ResponseWriter, r *http.Request, err error) bool {
	if err != nil {

		fmt.Println(r.Host + r.URL.Path)

		http.Redirect(w, r, r.Host+r.URL.Path, 301)
		return false
	}

	return true
}

// function query user untuk mengambil user berdasarkan userame nya
func QueryUser(username string) user {
	var users = user{}
	err = db.QueryRow(`
		SELECT id, 
		username, 
		email, 
		password 
		FROM users WHERE username=?
		`, username).
		Scan(
			&users.ID,
			&users.Username,
			&users.Email,
			&users.Password,
		)
	return users
}

// function untuk mengarahkan ke halaman home
func home(w http.ResponseWriter, r *http.Request) {
	//meng-chek ketersedian session
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/login", 301)
	}

	//properti untuk di set di html
	var data = map[string]string{
		"username": session.GetString("username"),
		"message":  "Selamat datang di menu utama",
	}

	var t, err = template.ParseFiles("views/home.html")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	t.Execute(w, data)
	return

}

// fungsi registers
func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "views/register.html")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")

	users := QueryUser(username)

	// fmt.Printf("%+v\n", (user{}))

	// fmt.Printf("%+v\n", users)

	//perbandingan user yang di post (users{}) dengan user yang ada di database users
	if (user{}) == users {

		// user belum tersedia, boleh registers
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		if len(hashedPassword) != 0 && checkErr(w, r, err) {
			stmt, err := db.Prepare("INSERT users SET username=?, password=?, email=?")
			if err == nil {
				_, err := stmt.Exec(&username, &hashedPassword, &email)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		}

	} else {
		//users sudah tersedia, gagal save
		http.Redirect(w, r, "/register", 302)
		fmt.Println("gagal save")

	}
}

// fungsi eksekusi login
func login(w http.ResponseWriter, r *http.Request) {
	//jika users sudah melakukan login dan masih aktif session nya
	//maka users tsb tidak usah login kembali
	session := sessions.Start(w, r)
	if len(session.GetString("username")) != 0 && checkErr(w, r, err) {
		http.Redirect(w, r, "/", 302)
	}

	if r.Method != "POST" {
		http.ServeFile(w, r, "views/login.html")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	users := QueryUser(username)

	//deskripsi dan compare password
	var password_tes = bcrypt.CompareHashAndPassword([]byte(users.Password), []byte(password))

	if password_tes == nil {
		//login success
		session := sessions.Start(w, r)
		session.Set("username", users.Username)
		http.Redirect(w, r, "/", 302)
	} else if password_tes == nil {
		//login success
		session := sessions.Start(w, r)
		session.Set("admin", users.level)
		http.Redirect(w, r, "views/admin.html", 302)
	} else {
		//login failed
		http.Redirect(w, r, "/login", 302)
	}
}

// fungsi logout
func logout(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	session.Clear()
	sessions.Destroy(w, r)
	http.Redirect(w, r, "/", 302)
}

// func image() {
// 	router := gin.Default()

// 	router.GET("/", func(c *gin.Context) {
// 		c.File("views/home.html")
// 	})

// 	router.GET("/image", func(c *gin.Context) {
// 		if currentImage == nil {
// 			c.AbortWithStatus(http.StatusNotFound)
// 			return
// 		}
// 		currentImage.Write(c.Writer)
// 	})

// 	router.GET("/thumbnail", func(c *gin.Context) {
// 		if currentImage == nil {
// 			c.AbortWithStatus(http.StatusNotFound)
// 		}

// 		t, err := imageupload.ThumbnailJPEG(currentImage, 300, 300, 80)

// 		if err != nil {
// 			panic(err)
// 		}

// 		t.Write(c.Writer)
// 	})

// 	router.POST("/upload", func(c *gin.Context) {
// 		img, err := imageupload.Process(c.Request, "file")
// 		if err != nil {
// 			panic(err)
// 		}

// 		currentImage = img

// 		c.Redirect(http.StatusMovedPermanently, "/")
// 	})

// 	router.Run(":8000")
// }

func image() {
	r := gin.Default()
	r.PUT("/user/:id", func(c *gin.Context) {
		var userObj user
		if err := c.ShouldBind(&userObj); err != nil {
			c.String(http.StatusBadRequest, "bad request")
			return
		}

		if err := c.ShouldBindUri(&userObj); err != nil {
			c.String(http.StatusBadRequest, "bad request")
			return
		}

		err := c.SaveUploadedFile(userObj.Avatar, userObj.Avatar.Filename)
		if err != nil {
			c.String(http.StatusInternalServerError, "unknown error")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"data":   userObj,
		})
	})
	r.Static("assets", "./assets")

	r.Run("localhost:8000")
}
