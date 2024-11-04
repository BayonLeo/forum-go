package server

import (
	"fmt"
	"forum-go/internal/models"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Implement function : retrieve form values and call AddCommet function in database/comment.go
func (s *Server) PostCommentHandler(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := s.db.GetPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render(w, r, "../posts", map[string]interface{}{"Posts": posts})
}

func (s *Server) PostNewcommentHandler(w http.ResponseWriter, r *http.Request) {

	type FormData struct {
		Title      string
		Content    string
		Categories []string
		Errors     map[string]string
	}
	erri := r.ParseForm()
	formData := FormData{
		Title:      r.FormValue("title"),
		Content:    r.FormValue("content"),
		Categories: r.Form["categories"],
		Errors:     make(map[string]string),
	}
	if erri != nil {
		log.Println(erri)
	}

	// Validate content
	if ValidatePostChar(formData.Content) {
		formData.Errors["Content"] = "Content cannot be empty or more than 1000 characters"
	}

	// Validate Categories
	if ValidateCategory(formData.Categories) {
		formData.Errors["Categories"] = "Please select at least one category"
	}
	if len(formData.Errors) > 0 {
		render(w, r, "createPost", map[string]interface{}{"FormData": formData, "Categories": s.categories})
		return
	}

	newPost := models.Post{
		PostId:  strconv.Itoa(rand.Intn(math.MaxInt32)),
		Title:   r.FormValue("title"),
		Content: r.FormValue("content"),
		UserID:  r.FormValue("UserId"),
		//Categories:
		CreationDate:          time.Now(),
		FormattedCreationDate: time.Now().Format("Jan 02, 2006 - 15:04:05"),
	}

	// charControl := ValidatePostChar(newPost.Content)
	categories := []models.Category{}
	for _, categoryID := range formData.Categories {
		for _, category := range s.categories {
			if category.CategoryId == categoryID {
				categories = append(categories, category)
			}
		}
	}
	newPost.Categories = categories
	err := s.db.AddPost(newPost, categories)
	s.posts = append(s.posts, newPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func (s *Server) DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	PostID := r.FormValue("postId")
	fmt.Println(PostID)
	err := s.db.DeletePost(PostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func (s *Server) GetNewCommentHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := s.db.GetCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !s.isLoggedIn(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	render(w, r, "createPost", map[string]interface{}{"Categories": categories})
}

func (s *Server) GetCommentHandler(w http.ResponseWriter, r *http.Request) {
	vars := strings.Split(r.URL.Path, "/")
	if len(vars) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	postID := vars[2]
	post, err := s.db.GetPost(postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if post.PostId == "" {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	render(w, r, "detailsPost", map[string]interface{}{"Post": post})
}

const MaxCharComment = 400

func ValidateCommentChar(content string) bool {
	if len(content) > MaxCharComment || len(content) == 0 {
		return true
	}
	return false
}

func (s *Server) EditCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentID := r.FormValue("commentId")
	commentContent := r.FormValue("newCommentContent")

	err := s.db.EditComment(commentID, commentContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}
