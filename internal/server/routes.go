package server

import (
	"encoding/json"
	"fmt"
	"forum-go/internal/models"
	"log"
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {

	mux := http.NewServeMux()

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	mux.HandleFunc("/", s.HomePageHandler)
	mux.HandleFunc("/about", s.AboutPageHandler)

	mux.HandleFunc("/activity", s.ActivityPageHandler)

	mux.HandleFunc("GET /login", s.GetLoginHandler)
	mux.HandleFunc("POST /login", s.PostLoginHandler)

	mux.HandleFunc("POST /logout", s.LogoutHandler)

	mux.HandleFunc("GET /register", s.GetRegisterHandler)
	mux.HandleFunc("POST /register", s.PostRegisterHandler)

	mux.HandleFunc("GET /delete/users/{id}", s.DeleteUsersHandler)
	mux.HandleFunc("GET /ban/users/{id}", s.BanUserHandler)
	mux.HandleFunc("GET /promote/users/{id}", s.PromoteUserHandler)
	mux.HandleFunc("GET /demote/users/{id}", s.DemoteUserHandler)

	mux.HandleFunc("GET /posts/create", s.GetNewPostHandler)
	mux.HandleFunc("POST /posts/create", s.PostNewPostsHandler)
	mux.HandleFunc("POST /posts/delete/{id}", s.DeletePostsHandler)
	mux.HandleFunc("POST /posts/edit/{id}", s.EditPostHandler)

	mux.HandleFunc("GET /categories", s.GetCategoriesHandler)
	mux.HandleFunc("POST /categories/add", s.PostCategoriesHandler)
	mux.HandleFunc("POST /categories/delete/{id}", s.DeleteCategoriesHandler)
	mux.HandleFunc("POST /categories/edit/{id}", s.EditCategoriesHandler)

	mux.HandleFunc("GET /post/{id}", s.GetPostHandler)
	mux.HandleFunc("POST /comment/delete/{id}", s.DeleteCommentHandler)
	mux.HandleFunc("POST /comment/edit/{id}", s.EditCommentHandler)
	mux.HandleFunc("POST /post/comment", s.PostCommentHandler)

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("GET /adminPanel", s.AdminPanelHandler)
	mux.HandleFunc("GET /report", s.reportHandler)
	mux.HandleFunc("GET /adminPanel/modrequests", s.ModRequestsHandler)
	mux.HandleFunc("POST /vote", s.VoteHandler)

	mux.HandleFunc("POST /modRequest", s.PostModRequestHandler)
	mux.HandleFunc("GET /modRequest", s.GetModRequestHandler)
	mux.HandleFunc("POST /modRequest/accepted", s.AcceptRequestHandler)
	mux.HandleFunc("POST /modRequest/rejected", s.RejectRequestHandler)
	return s.authenticate(mux)
}

func (s *Server) VoteHandler(w http.ResponseWriter, r *http.Request) {
	if !s.isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	postID := r.FormValue("post_id")
	userID := r.FormValue("user_id")
	vote := r.FormValue("vote")
	commentID := r.FormValue("comment_id")
	var isLike bool
	if vote == "like" {
		isLike = true
	} else {
		isLike = false
	}

	err := s.db.Vote(postID, commentID, userID, isLike)
	if err != nil {
		s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	post, err := s.db.GetPost(postID)
	if err != nil {
		s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
	}
	if commentID == "" {
		if isLike {
			if post.UserID != s.getUser(r).UserId {
				newActivity := models.NewActivity(post.UserID, userID, string(models.GET_POST_LIKED), postID, "", post.Title)
				err = s.db.CreateActivity(newActivity)
				if err != nil {
					s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
				}
			}
			newActivity := models.NewActivity(s.getUser(r).UserId, userID, string(models.POST_LIKED), postID, "", post.Title)
			err = s.db.CreateActivity(newActivity)
			if err != nil {
				s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
			}
		} else {
			if post.UserID != s.getUser(r).UserId {
				newActivity := models.NewActivity(post.UserID, userID, string(models.GET_POST_DISLIKED), postID, "", post.Title)
				err = s.db.CreateActivity(newActivity)
				if err != nil {
					s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
				}
			}
			newActivity := models.NewActivity(s.getUser(r).UserId, userID, string(models.POST_DISLIKED), postID, "", post.Title)
			err = s.db.CreateActivity(newActivity)
			if err != nil {
				s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
			}
		}
	} else {
		ActualComment := models.Comment{}
		for _, comment := range post.Comments {
			if comment.CommentId == commentID {
				ActualComment = comment
			}
		}
		if isLike {
			if ActualComment.UserID != s.getUser(r).UserId {
				newActivity := models.NewActivity(ActualComment.UserID, userID, string(models.GET_COMMENT_LIKED), postID, ActualComment.CommentId, ActualComment.Content)
				err = s.db.CreateActivity(newActivity)
				if err != nil {
					s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
				}
			}
			newActivity := models.NewActivity(s.getUser(r).UserId, userID, string(models.COMMENT_LIKED), postID, ActualComment.CommentId, ActualComment.Content)
			err = s.db.CreateActivity(newActivity)
			if err != nil {
				s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
			}
		} else {
			if ActualComment.UserID != s.getUser(r).UserId {
				newActivity := models.NewActivity(ActualComment.UserID, userID, string(models.GET_COMMENT_DISLIKED), postID, ActualComment.CommentId, ActualComment.Content)
				err = s.db.CreateActivity(newActivity)
				if err != nil {
					s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
				}
			}
			newActivity := models.NewActivity(s.getUser(r).UserId, userID, string(models.COMMENT_DISLIKED), postID, ActualComment.CommentId, ActualComment.Content)
			err = s.db.CreateActivity(newActivity)
			if err != nil {
				s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
			}
		}
	}
	referer := r.Header.Get("Referer")
	if referer != "" {
		http.Redirect(w, r, referer, http.StatusSeeOther)
		return
	}
}

func (s *Server) reportHandler(w http.ResponseWriter, r *http.Request) {
	if !s.isLoggedIn(r) || (!IsAdmin(r) && !IsModerator(r)) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	users, err := s.db.GetUsers()
	if err != nil {
		s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	render(w, r, "report", map[string]interface{}{"users": users})

}

func (s *Server) AdminPanelHandler(w http.ResponseWriter, r *http.Request) {
	if !s.isLoggedIn(r) || !IsAdmin(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	users, err := s.db.GetUsers()
	if err != nil {
		s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	render(w, r, "admin/adminPanel", map[string]interface{}{"users": users})
}

func (s *Server) HomePageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/created" && r.URL.Path != "/liked" {
		s.errorHandler(w, r, http.StatusNotFound, "Page not found")
		return
	}
	if !s.isLoggedIn(r) && (r.URL.Path == "/created" || r.URL.Path == "/liked") {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	err := error(nil)
	s.categories, err = s.db.GetCategories()
	if err != nil {
		s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	s.posts, err = s.db.GetPosts()
	if err != nil {
		s.errorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	for i, post := range s.posts {
		s.posts[i].HasVoted = GetUserVote(post, s.getUser(r).UserId)
	}
	postsToRender := []models.Post{}
	if r.URL.Path == "/created" {
		for _, post := range s.posts {
			if post.UserID == s.getUser(r).UserId {
				postsToRender = append(postsToRender, post)
			}
		}
	} else if r.URL.Path == "/liked" {
		for _, post := range s.posts {
			if post.HasVoted == 1 {
				postsToRender = append(postsToRender, post)
			}
		}
	} else {
		postsToRender = s.posts
	}

	render(w, r, "home", map[string]interface{}{"Categories": s.categories, "Posts": postsToRender})
}
func (s *Server) AboutPageHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "about", nil)
}
func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"
	for k, v := range r.Header {
		resp[k] = fmt.Sprintf("%v", v)
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) errorHandler(w http.ResponseWriter, r *http.Request, status int, message string) {
	w.WriteHeader(status)
	error := models.Error{Message: message, StatusCode: status}
	render(w, r, "error", map[string]interface{}{"Error": error})
}
