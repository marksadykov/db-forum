package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fasthttp/router"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgxpool"
	"github.com/labstack/gommon/log"
	"github.com/valyala/fasthttp"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

//var address = "127.0.0.1:3001"
var address = "0.0.0.0:5000"

//models
type (
	Error struct {
		Message string `json:"message"`
	}

	Forum struct {
		Posts int64 `json:"-"`
		Slug string `json:"slug"`
		Threads int64 `json:"-"`
		Title string `json:"title"`
		User string `json:"user"`
	}

	ForumDetails struct {
		Posts int64 `json:"posts"`
		Slug string `json:"slug"`
		Threads int64 `json:"threads"`
		Title string `json:"title"`
		User string `json:"user"`
	}

	Post struct {
		Author string `json:"author"`
		Created time.Time `json:"created"`
		Forum string `json:"forum"`
		Id uint64 `json:"id"`
		IsEdited bool `json:"isEdited"`
		Message string `json:"message"`
		Parent uint64 `json:"parent"`
		Thread uint32 `json:"thread"`
		UserId uint64 `json:"-"`
		About string `json:"-"`
		Fullname string `json:"-"`
		Email string `json:"-"`
		Children Posts `json:"-"`
	}

	//PostForDetails struct {
	//	Author string `json:"author"`
	//	Created string `json:"created"`
	//	Forum string `json:"forum"`
	//	Id uint64 `json:"id"`
	//	IsEdited bool `json:"isEdited"`
	//	Message string `json:"message"`
	//	Parent uint64 `json:"parent"`
	//	Thread uint32 `json:"thread"`
	//}
	//
	//PostForDetailsPost struct {
	//	Author string `json:"author"`
	//	Created string `json:"created"`
	//	Forum string `json:"forum"`
	//	Id uint64 `json:"id"`
	//	IsEdited bool `json:"isEdited"`
	//	Message string `json:"message"`
	//	Parent uint64 `json:"parent"`
	//	Thread uint32 `json:"thread"`
	//}

	PostPost struct {
		Post Post `json:"post"`
	}

	PostAuthor struct {
		Author User `json:"author"`
		Post Post `json:"post"`
	}

	PostThread struct {
		Thread Thread `json:"thread"`
		Post Post `json:"post"`
	}

	PostForum struct {
		Forum ForumDetails `json:"forum"`
		Post Post `json:"post"`
	}

	PostForumThread struct {
		Thread Thread `json:"thread"`
		Forum ForumDetails `json:"forum"`
		Post Post `json:"post"`
	}

	PostForumUser struct {
		Author User `json:"author"`
		Forum ForumDetails `json:"forum"`
		Post Post `json:"post"`
	}

	PostThreadUser struct {
		Author User `json:"author"`
		Thread Thread `json:"thread"`
		Post Post `json:"post"`
	}

	PostFullDetails struct {
		Forum ForumDetails `json:"forum"`
		Author User `json:"author"`
		Thread Thread `json:"thread"`
		Post Post `json:"post"`
	}

	PostUpdate struct {
		Id uint64 `json:"id"`
		Message string `json:"message"`
	}

	Posts []Post

	Status struct {
		Forum uint32 `json:"forum"`
		Post uint64 `json:"post"`
		Thread uint32 `json:"thread"`
		User uint32 `json:"user"`
	}

	Thread struct {
		Author string `json:"author"`
		Created time.Time `json:"created"`
		Forum string `json:"forum"`
		Id uint32 `json:"id"`
		Message string `json:"message"`
		Slug string `json:"slug"`
		Title string `json:"title"`
		Votes int32 `json:"votes"`
	}

	Threads []Thread

	User struct {
		Id uint64 `json:"-"`
		About string `json:"about"`
		Email string `json:"email"`
		Fullname string `json:"fullname"`
		Nickname string `json:"nickname"`
	}

	Users []User

	Vote struct {
		Nickname string `json:"nickname"`
		Voice int32 `json:"voice"`
	}

	CurrentVote struct {
		Id uint32 `json:"-"`
		Nickname string `json:"-"`
		Voice int32 `json:"-"`
		threadSlug string `json:"-"`
		threadId uint32 `json:"-"`
	}
)

func sortTree(input Posts, parentId uint64, descBool bool) (Posts){
	var output Posts
	for i := 0; i < len(input); i++ {
		if (input[i].Parent == parentId) {
			input[i].Children = sortTree(input, input[i].Id, descBool)
			if descBool {
				sort.SliceStable(input[i].Children, func(m, j int) bool {
					return input[i].Children[m].Id > input[i].Children[j].Id
				})
			} else {
				sort.SliceStable(input[i].Children, func(m, j int) bool {
					return input[i].Children[m].Id < input[i].Children[j].Id
				})
			}
			output = append(output, input[i])
		}
	}
	return output
}


func sortTreeParent(input Posts, parentId uint64) (Posts){
	var output Posts
	for i := 0; i < len(input); i++ {
		if (input[i].Parent == parentId) {
			input[i].Children = sortTreeParent(input, input[i].Id)
			sort.SliceStable(input[i].Children, func(m, j int) bool {
				return input[i].Children[m].Id < input[i].Children[j].Id
			})
			output = append(output, input[i])
		}
	}
	return output
}

func showTree(root Post) (Posts){
	var nodes Posts
	var output Posts

	nodes = append(nodes, root)

	for len(nodes) > 0 {
		current := nodes[0]
		nodes = nodes[1:]
		output  = append(output, current)
		for i := (len(current.Children) - 1); i >= 0; i-- {
			nodes = append([]Post{current.Children[i]}, nodes...)
		}
	}
	return output
}

func showTreeReverse(root Post) (Posts){
	var nodes Posts
	var output Posts

	nodes = append(nodes, root)

	for len(nodes) > 0 {
		current := nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]
		output  = append(Posts{current}, output...)
		for i := 0 ; i < len(current.Children); i++ {
			nodes = append(nodes, current.Children[i])
		}
	}
	return output
}

func showFullTree(tree Posts, descBool bool) (Posts){
	var output Posts
	var postsNodes Post
	for i := 0; i < len(tree); i++ {
		postsNodes.Children = append(postsNodes.Children, tree[i])
		if descBool {
			output = showTreeReverse(postsNodes)
		} else {
			output = showTree(postsNodes)
		}
	}
	if descBool {
		output = output[:len(output)-1]
	} else {
		if len(output) > 0 {
			output = output[1:]
		}
	}
	return output
}

func JohnySins (posts Posts, sins uint64) (Posts) {
	var output Posts
	myIndex := 0
	for i := 0; i < len(posts); i++ {
		if posts[i].Id == sins  {
			myIndex = i + 1
		}
	}
	if myIndex == 0 {
		return Posts{}
	}
	if len(posts) > 0 {
		output = posts[myIndex:]
	}
	return output
}

func showFullParentTree(tree Posts) (Posts){
	var output Posts
	var postsNodes Post
	for i := 0; i < len(tree); i++ {
		postsNodes.Children = append(postsNodes.Children, tree[i])
		output = showTree(postsNodes)
	}
	if len(output) > 0 {
		output = output[1:]
	}
	return output
}

//API
type RequestHandler struct {
	pool *pgxpool.Pool
}

func NewRequestHandler(p *pgxpool.Pool) *RequestHandler {
	return &RequestHandler{
		pool:  p,
	}
}

// Handlers
// /forum/create
func (h *RequestHandler) createForum(ctx *fasthttp.RequestCtx) {
	log.Printf("/forum/create")

	forum := Forum{}
	err := json.Unmarshal(ctx.PostBody(), &forum)
	if err != nil {
		log.Printf("Unmarshal:", err)
	}

	status, response := InsertCreateForum(h.pool, forum)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}

// /forum/{slug}/details
func (h *RequestHandler) getForumDetails(ctx *fasthttp.RequestCtx) {
	//log.Printf("/forum/{slug}/details")

	slug := fmt.Sprintf("%v", ctx.UserValue("slug"))

	status, response := getForumDetails(h.pool, slug)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}

// /forum/{slug}/create
func (h *RequestHandler) createForumThread(ctx *fasthttp.RequestCtx) {
	//log.Printf("/forum/{slug}/create")

	thread := Thread{}
	thread.Forum = fmt.Sprintf("%v", ctx.UserValue("slug"))
	err := json.Unmarshal(ctx.PostBody(), &thread)
	if err != nil {
		log.Printf("Unmarshal:", err)
	}

	status, response := InsertCreateForumThread(h.pool, thread)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}

// /forum/{slug}/users
func (h *RequestHandler) getForumUsers(ctx *fasthttp.RequestCtx) {
	//log.Printf("/forum/{slug}/users")

	slug := fmt.Sprintf("%v", ctx.UserValue("slug"))

	limitParam := ctx.QueryArgs().Peek("limit")
	sinceParam := ctx.QueryArgs().Peek("since")
	descParam := ctx.QueryArgs().Peek("desc")

	limit := string(limitParam)
	since := string(sinceParam)
	desc := string(descParam)

	status, response := getForumUsers(h.pool, slug, limit, since, desc)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}

// /forum/{slug}/threads
func (h *RequestHandler) getForumThreads(ctx *fasthttp.RequestCtx) {
	//log.Printf("/forum/{slug}/threads")

	slug := fmt.Sprintf("%v", ctx.UserValue("slug"))

	limitParam := ctx.QueryArgs().Peek("limit")
	sinceParam := ctx.QueryArgs().Peek("since")
	descParam := ctx.QueryArgs().Peek("desc")

	limit := string(limitParam)
	since := string(sinceParam)
	desc := string(descParam)

	status, response := getForumThreads(h.pool, slug, limit, since, desc)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}

// /post/{id}/details
func (h *RequestHandler) getPostDetails(ctx *fasthttp.RequestCtx) {
	//log.Printf("/post/{id}/details")

	id := fmt.Sprintf("%v", ctx.UserValue("id"))

	relatedParam := ctx.QueryArgs().Peek("related")
	related := string(relatedParam)

	status, response := getPostDetails(h.pool, id, related)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}


// /post/{id}/details
func (h *RequestHandler) updatePost(ctx *fasthttp.RequestCtx) {
	//log.Printf("/post/{id}/details")

	postUpdate := PostUpdate{}
	id := fmt.Sprintf("%v", ctx.UserValue("id"))
	err := json.Unmarshal(ctx.PostBody(), &postUpdate)
	if err != nil {
		log.Printf("Unmarshal:", err)
	}

	status, response := updatePost(h.pool, id, postUpdate)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}

// /service/clear
func (h *RequestHandler) deleteService(ctx *fasthttp.RequestCtx) {
	//log.Printf("/service/clear")

	status, response := deleteService(h.pool)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}

// /service/status
func (h *RequestHandler) getService(ctx *fasthttp.RequestCtx) {
	//log.Printf("/service/status")

	status, response := getService(h.pool)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)

}

// /thread/{slug_or_id}/create
func (h *RequestHandler) addPostThread(ctx *fasthttp.RequestCtx) {
	//log.Printf("/thread/{slug_or_id}/create")

	posts := Posts{}
	slugOrId := fmt.Sprintf("%v", ctx.UserValue("slug_or_id"))
	err := json.Unmarshal(ctx.PostBody(), &posts)
	if err != nil {
		log.Printf("Unmarshal:", err)
	}

	status, response := addPostThread(h.pool, slugOrId, posts)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}

// /thread/{slug_or_id}/details
func (h *RequestHandler) getThreadDetails(ctx *fasthttp.RequestCtx) {
	//log.Printf("/thread/{slug_or_id}/details")

	slug := fmt.Sprintf("%v", ctx.UserValue("slug_or_id"))

	status, response := getThreadDetails(h.pool, slug)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}


// /thread/{slug_or_id}/details
func (h *RequestHandler) updateThreadDetails(ctx *fasthttp.RequestCtx) {
	//log.Printf("/thread/{slug_or_id}/details")

	slug := fmt.Sprintf("%v", ctx.UserValue("slug_or_id"))

	var thread Thread
	err := json.Unmarshal(ctx.PostBody(), &thread)
	if err != nil {
		log.Printf("Unmarshal:", err)
	}

	status, response := updateThreadDetails(h.pool, slug, thread)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}

// /thread/{slug_or_id}/posts
func (h *RequestHandler) getPostThread(ctx *fasthttp.RequestCtx) {
	//log.Printf("/thread/{slug_or_id}/posts")

	slug := fmt.Sprintf("%v", ctx.UserValue("slug_or_id"))

	limitParam := ctx.QueryArgs().Peek("limit")
	sinceParam := ctx.QueryArgs().Peek("since")
	descParam := ctx.QueryArgs().Peek("desc")
	sortParam := ctx.QueryArgs().Peek("sort")

	limit := string(limitParam)
	since := string(sinceParam)
	desc := string(descParam)
	sortP := string(sortParam )

	status, response := getPostThread(h.pool, slug, limit, since, desc, sortP)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}

// /thread/{slug_or_id}/vote
func (h *RequestHandler) addVoteThread(ctx *fasthttp.RequestCtx) {
	//log.Printf("/thread/{slug_or_id}/vote")

	slug := fmt.Sprintf("%v", ctx.UserValue("slug_or_id"))

	var vote Vote
	err := json.Unmarshal(ctx.PostBody(), &vote)
	if err != nil {
		log.Printf("Unmarshal:", err)
	}

	status, response := addVoteThread(h.pool, slug, vote)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}

// /user/{nickname}/create
func (h *RequestHandler) createUser(ctx *fasthttp.RequestCtx) {
	log.Printf("/user/{nickname}/create")

	user := User{}
	user.Nickname = fmt.Sprintf("%v", ctx.UserValue("nickname"))
	err := json.Unmarshal(ctx.PostBody(), &user)
	if err != nil {
		log.Printf("Unmarshal:", err)
	}

	status, response := InsertCreateUser(h.pool, user)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}

// /user/{nickname}/profile
func (h *RequestHandler) getUser(ctx *fasthttp.RequestCtx) {
	//log.Printf("/user/{nickname}/profile")

	nickname := fmt.Sprintf("%v", ctx.UserValue("nickname"))

	status, response := getUser(h.pool, nickname)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}

// /user/{nickname}/profile
func (h *RequestHandler) updateUser(ctx *fasthttp.RequestCtx) {
	//log.Printf("/user/{nickname}/profile")

	user := User{}
	user.Nickname = fmt.Sprintf("%v", ctx.UserValue("nickname"))
	err := json.Unmarshal(ctx.PostBody(), &user)
	if err != nil {
		log.Printf("Unmarshal:", err)
	}

	status, response := updateUser(h.pool, user)
	//log.Printf("status:", status)
	//log.Printf("response:", response)

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(response)
}


//CRUD
//CreateForum
func InsertCreateForum(p *pgxpool.Pool, forum Forum) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	var id uint64
	err = conn.QueryRow(context.Background(), "SELECT id, nickname FROM public.users WHERE lower(nickname) = $1", strings.ToLower(forum.User)).Scan(&id, &forum.User)
	if err != nil {
		log.Errorf("Unable to SELECT id FROM users: %v\n", err)
		error := Error{"Can't find user," + forum.User + "," + forum.Title + "," + forum.Slug}
		response, _ = json.Marshal(error)
		return 404, response
	}

	var currentForum Forum
	var userId int64
	err = conn.QueryRow(context.Background(), "SELECT title, user_id, slug FROM public.forum WHERE lower(slug) = $1", strings.ToLower(forum.Slug)).Scan(&currentForum.Title, &userId, &currentForum.Slug)

	if (currentForum.Title != "") {
		//log.Errorf("Unable to INSERT: %v\n", err)
		err = conn.QueryRow(context.Background(), "SELECT nickname FROM public.users WHERE id = $1", userId).Scan(&currentForum.User)
		response, _ = json.Marshal(currentForum)
		return 409, response
	}

	conn.QueryRow(context.Background(),
		"INSERT INTO public.forum (slug, title, user_id, posts, threads) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		forum.Slug, forum.Title, id, 0, 0)

	response, _ = json.Marshal(forum)

	return 201, response
}

//InsertCreateForumThread
func InsertCreateForumThread(p *pgxpool.Pool, thread Thread) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	var forumId uint64
	var threads int64
	err = conn.QueryRow(context.Background(), "SELECT id, threads, slug FROM public.forum WHERE lower(slug) = $1", strings.ToLower(thread.Forum)).Scan(&forumId, &threads, &thread.Forum)
	if err != nil {
		//log.Errorf("Unable to SELECT id FROM forum: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}

	var userId uint64
	var user User
	err = conn.QueryRow(context.Background(), "SELECT id, about, email, fullname FROM public.users WHERE lower(nickname) = $1", strings.ToLower(thread.Author)).Scan(&userId, &user.About, &user.Email, &user.Fullname)
	if err != nil {
		//log.Errorf("Unable to SELECT id FROM users: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}

	var currentThread Thread
	err = conn.QueryRow(context.Background(), "SELECT id, created, message, votes, forum_id, user_id, slug, users_nickname, forum, title FROM public.thread WHERE lower(slug) = $1", strings.ToLower(thread.Slug)).Scan(&currentThread.Id, &currentThread.Created, &currentThread.Message, &currentThread.Votes, &forumId, &userId, &currentThread.Slug, &currentThread.Author, &currentThread.Forum, &currentThread.Title)
	if (currentThread.Title != "" && currentThread.Slug != "") {
		//log.Errorf("Unable to INSERT: %v\n", err)
		//currentThread.Created = currentThread.Created.Add(-3 * time.Hour)
		response, _ = json.Marshal(currentThread)
		return 409, response
	}


	var row pgx.Row
	if (thread.Slug == "") {
		row = conn.QueryRow(context.Background(),
			"INSERT INTO public.thread (forum, created, message, title, votes, forum_id, user_id, users_nickname, users_fullname, users_email, users_about) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id",
			thread.Forum, thread.Created, thread.Message, thread.Title, 0, forumId, userId, thread.Author, user.Fullname, user.Email, user.About)
	} else {
		row = conn.QueryRow(context.Background(),
			"INSERT INTO public.thread (forum, slug, created, message, title, votes, forum_id, user_id, users_nickname, users_fullname, users_email, users_about) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id",
			thread.Forum, thread.Slug, thread.Created, thread.Message, thread.Title, 0, forumId, userId, thread.Author, user.Fullname, user.Email, user.About)
	}

	err = row.Scan(&thread.Id)

	threads++
	conn.QueryRow(context.Background(), "UPDATE public.forum SET threads = $2 WHERE id = $1", forumId, threads)

	response, _ = json.Marshal(thread)

	return 201, response
}

//getForumDetails
func getForumDetails(p *pgxpool.Pool, slug string) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	var forum ForumDetails
	var userId uint64
	err = conn.QueryRow(context.Background(), "SELECT title, user_id, slug, posts, threads FROM public.forum WHERE lower(slug) = $1", strings.ToLower(slug)).Scan(&forum.Title, &userId, &forum.Slug, &forum.Posts, &forum.Threads)
	if err != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}

	err = conn.QueryRow(context.Background(), "SELECT nickname FROM public.users WHERE id = $1", userId).Scan(&forum.User)

	response, _ = json.Marshal(forum)
	return 200, response
}

//getForumThreads
func getForumThreads(p *pgxpool.Pool, slug string, limit string, since string, desc string) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	limitInt := 0
	var sinceDate time.Time
	descBool := false

	if limit != "" {
		limitInt, _ = strconv.Atoi(limit)
	}

	if since != "" {
		sinceDate, _ = time.Parse(time.RFC3339, since)
		//sinceDate = sinceDate.Add(3 * time.Hour)
	}

	if desc != "" {
		descBool, _ = strconv.ParseBool(desc)
	}

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	var forumId uint64
	err = conn.QueryRow(context.Background(), "SELECT id FROM public.forum WHERE lower(slug) = $1", strings.ToLower(slug)).Scan(&forumId)
	if err != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}

	var threads Threads
	var i int
	var rows pgx.Rows
	if since != "" {
		if descBool {
			rows, err = conn.Query(context.Background(), "SELECT created, users_nickname, id, message, title, votes, forum, slug FROM public.thread WHERE forum_id = $1 AND created <= $2", forumId, sinceDate)
		} else {
			rows, err = conn.Query(context.Background(), "SELECT created, users_nickname, id, message, title, votes, forum, slug FROM public.thread WHERE forum_id = $1 AND created >= $2", forumId, sinceDate)
		}
	} else {
		rows, err = conn.Query(context.Background(), "SELECT created, users_nickname, id, message, title, votes, forum, slug  FROM public.thread WHERE forum_id = $1", forumId)
	}
	for rows.Next() {
		var thread Thread
		if err = rows.Scan(&thread.Created, &thread.Author, &thread.Id, &thread.Message, &thread.Title, &thread.Votes, &thread.Forum, &thread.Slug); err == nil {
			//thread.Created = thread.Created.Add(-3 * time.Hour)
		}
		threads = append(threads, thread)
		i++
	}

	if descBool {
		sort.SliceStable(threads, func(i, j int) bool {
			return threads[i].Created.After(threads[j].Created)
		})
	} else {
		sort.SliceStable(threads, func(i, j int) bool {
			return threads[i].Created.Before(threads[j].Created)
		})
	}

	if (limitInt != 0 && limitInt <= len(threads)) {
		threads = threads[:limitInt]
	}


	if (len(threads) == 0) {
		response := []byte("[]")
		return 200, response
	}

	response, _ = json.Marshal(threads)
	return 200, response
}

//getForumUsers
func getForumUsers(p *pgxpool.Pool, slug string, limit string, since string, desc string) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	limitInt := 0
	sinceSkip := ""
	descBool := false

	if limit != "" {
		limitInt, _ = strconv.Atoi(limit)
	}

	if since != "" {
		sinceSkip = strings.ToLower(since)
	}

	if desc != "" {
		descBool, _ = strconv.ParseBool(desc)
	} else {
		descBool = false
	}

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	var forumId uint64
	err = conn.QueryRow(context.Background(), "SELECT id FROM public.forum WHERE lower(slug) = $1", strings.ToLower(slug)).Scan(&forumId)
	if err != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}

	var users Users
	var rows pgx.Rows
	rows, err = conn.Query(context.Background(), "SELECT users_nickname, users_about, users_email, users_fullname, user_id FROM public.thread WHERE forum_id = $1 UNION SELECT users_nickname, users_about, users_email, users_fullname, user_id FROM post WHERE lower(forum) = $2", forumId, strings.ToLower(slug))

	for rows.Next() {
		var user User
		err = rows.Scan(&user.Nickname, &user.About, &user.Email, &user.Fullname, &user.Id)
		users = append(users, user)
	}

	if descBool {
		sort.SliceStable(users , func(m, n int) bool {
			return strings.ToLower(users[m].Nickname) > strings.ToLower(users[n].Nickname)
		})
	} else {
		sort.SliceStable(users , func(m, n int) bool {
			return strings.ToLower(users[m].Nickname) < strings.ToLower(users[n].Nickname)
		})
	}

	if since != "" {
		skip := 0
		for i := 0; i < len(users); i++ {
			if strings.ToLower(users[i].Nickname) == sinceSkip {
				skip = i + 1
				break
			}
		}
		users = users[skip:]
	}

	if (limitInt != 0 && limitInt <= len(users)) {
		users = users[:limitInt]
	}

	if (len(users) == 0) {
		response := []byte("[]")
		return 200, response
	}

	response, _ = json.Marshal(users)
	return 200, response
}

//getThreadDetails
func getThreadDetails(p *pgxpool.Pool, slug string) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	selector, err := strconv.Atoi(slug)
	var flag bool

	if err != nil {
		flag = true
	} else {
		flag = false
	}

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	var thread Thread
	var row interface{}

	var slugInter interface{}
	if flag {
		row = conn.QueryRow(context.Background(), "SELECT users_nickname, created, id, message, title, votes, slug, forum FROM public.thread WHERE lower(slug) = $1", strings.ToLower(slug)).Scan(&thread.Author, &thread.Created, &thread.Id, &thread.Message, &thread.Title, &thread.Votes, &slugInter, &thread.Forum)
	} else {
		row = conn.QueryRow(context.Background(), "SELECT users_nickname, created, id, message, title, votes, slug, forum FROM public.thread WHERE id = $1", selector).Scan(&thread.Author, &thread.Created, &thread.Id, &thread.Message, &thread.Title, &thread.Votes, &slugInter, &thread.Forum)
	}

	if row != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find thread" + thread.Author + thread.Message + thread.Title + fmt.Sprintf("%v", slugInter) + thread.Forum}
		response, _ = json.Marshal(error)
		return 404, response
	}

	thread.Slug = fmt.Sprintf("%v", slugInter)

	//thread.Created = thread.Created.Add(-3 * time.Hour)

	response, _ = json.Marshal(thread)
	return 200, response
}


//updateThreadDetails
func updateThreadDetails(p *pgxpool.Pool, slug string, threadNew Thread) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	selector, err := strconv.Atoi(slug)
	var flag bool

	if err != nil {
		flag = true
	} else {
		flag = false
	}

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	var row interface{}
	var thread Thread

	if flag {
		row = conn.QueryRow(context.Background(), "SELECT users_nickname, created, id, message, title, votes, slug, forum FROM public.thread WHERE lower(slug) = $1", strings.ToLower(slug)).Scan(&thread.Author, &thread.Created, &thread.Id, &thread.Message, &thread.Title, &thread.Votes, &thread.Slug, &thread.Forum)
	} else {
		row = conn.QueryRow(context.Background(), "SELECT users_nickname, created, id, message, title, votes, slug, forum FROM public.thread WHERE id = $1", selector).Scan(&thread.Author, &thread.Created, &thread.Id, &thread.Message, &thread.Title, &thread.Votes, &thread.Slug, &thread.Forum)
	}

	if flag {
		if threadNew.Title != "" && threadNew.Message == "" {
			conn.QueryRow(context.Background(), "UPDATE public.thread SET title = $2 WHERE slug = $1", thread.Slug, threadNew.Title)
		}
		if threadNew.Title == "" && threadNew.Message != "" {
			conn.QueryRow(context.Background(), "UPDATE public.thread SET message = $2 WHERE slug = $1", thread.Slug, threadNew.Message)
		}
		if threadNew.Title != "" && threadNew.Message != "" {
			conn.QueryRow(context.Background(), "UPDATE public.thread SET title = $2, message = $3 WHERE slug = $1", thread.Slug, threadNew.Title, threadNew.Message)
		}
	} else {
		if threadNew.Title != "" && threadNew.Message == "" {
			conn.QueryRow(context.Background(), "UPDATE public.thread SET title = $2 WHERE id = $1", selector, threadNew.Title)
		}
		if threadNew.Title == "" && threadNew.Message != "" {
			conn.QueryRow(context.Background(), "UPDATE public.thread SET message = $2 WHERE id = $1", selector, threadNew.Message)
		}
		if threadNew.Title != "" && threadNew.Message != "" {
			conn.QueryRow(context.Background(), "UPDATE public.thread SET title = $2, message = $3 WHERE id = $1", selector, threadNew.Title, threadNew.Message)
		}
	}


	if threadNew.Title != "" {
		thread.Title = threadNew.Title
	}

	if threadNew.Message != "" {
		thread.Message = threadNew.Message
	}

	//thread.Created = thread.Created.Add(-3 * time.Hour)
	if row != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}

	response, _ = json.Marshal(thread)
	return 200, response
}


//getPostThread
func getPostThread(p *pgxpool.Pool, slug string, limit string, since string, desc string, sortType string) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	limitInt := 0
	sinceIntBack := 0
	descBool := false
	var sinceInt uint64

	if limit != "" {
		limitInt, _ = strconv.Atoi(limit)
	}

	if since != "" {
		sinceIntBack, _ = strconv.Atoi(since)
		sinceInt = uint64(sinceIntBack)
	}

	if desc != "" {
		descBool, _ = strconv.ParseBool(desc)
	} else {
		descBool = false
	}

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	selector, err := strconv.Atoi(slug)
	var flag bool

	if err != nil {
		flag = true
	} else {
		flag = false
	}

	var threadId uint64
	if flag {
		err = conn.QueryRow(context.Background(), "SELECT id FROM public.thread WHERE lower(slug) = $1", strings.ToLower(slug)).Scan(&threadId)
	} else {
		err = conn.QueryRow(context.Background(), "SELECT id FROM public.thread WHERE id = $1", selector).Scan(&threadId)
	}

	if err != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}

	var parentId uint64
	parentId = 100000000
	var posts Posts
	var i int
	var rows pgx.Rows
	rows, err = conn.Query(context.Background(), "SELECT created, users_nickname, id, message, forum, thread_id, parent FROM public.post WHERE thread_id = $1", threadId)

	for rows.Next() {
		var post Post
		if err = rows.Scan(&post.Created, &post.Author, &post.Id, &post.Message, &post.Forum, &post.Thread, &post.Parent); err == nil {
			//post.Created = post.Created.Add(-3 * time.Hour)
			if post.Parent < parentId {
				parentId = post.Parent
			}
		}
		posts = append(posts, post)
		i++
	}

	if sortType == "flat" {
		sort.SliceStable(posts, func(i, j int) bool {
			return posts[i].Created.Before(posts[j].Created)
		})
	}

	if sortType == "tree" {
		posts = sortTree(posts, parentId, descBool)
		if descBool {
			sort.SliceStable(posts, func(m, n int) bool {
				return posts[m].Id > posts[n].Id
			})
		} else {
			sort.SliceStable(posts, func(m, n int) bool {
				return posts[m].Id < posts[n].Id
			})
		}
		posts = showFullTree(posts, descBool)
	}

	var limitId uint64
	if sortType == "parent_tree" {
		posts = sortTreeParent(posts, parentId)
		if descBool {
			sort.SliceStable(posts, func(m, n int) bool {
				return posts[m].Id > posts[n].Id
			})
		} else {
			sort.SliceStable(posts, func(m, n int) bool {
				return posts[m].Id < posts[n].Id
			})
		}
		if !descBool {
			if (limitInt != 0 && limitInt <= len(posts)) {
				limitId = posts[limitInt].Id
			}
		}
		posts = showFullParentTree(posts)
	}

	if (sortType == "flat" || sortType == "" ) && descBool {
		sort.SliceStable(posts, func(i, j int) bool {
			return posts[i].Id > posts[j].Id
		})
	}

	if (sortType == "flat" || sortType == "" ) && !descBool{
		sort.SliceStable(posts, func(i, j int) bool {
			return posts[i].Id < posts[j].Id
		})
	}

	if since != "" {
		posts = JohnySins(posts, sinceInt)
	}

	if (limitInt != 0 && limitInt <= len(posts)) {
		if sortType == "parent_tree"{
			if descBool {
				skip := 0
				currentParentCount := 0
				for i := 0; i < len(posts); i++ {
					if posts[i].Parent == parentId {
						currentParentCount++
					}
					if currentParentCount > limitInt {
						skip = i
						break
					}
				}
				if currentParentCount > limitInt {
					posts = posts[:skip]
				}
			} else {
				var skip int
				for i := 0; i < len(posts); i++ {
					if posts[i].Id == limitId {
						skip = i
						break
					}
				}
				posts = posts[:skip]
			}
		} else {
			posts = posts[:limitInt]
		}
	}

	if (len(posts) == 0) {
		response := []byte("[]")
		return 200, response
	}

	response, _ = json.Marshal(posts)
	return 200, response
}

//addVoteThread
func addVoteThread(p *pgxpool.Pool, slug string, vote Vote) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	//tx, _ := p.Begin(context.Background())
	//defer tx.Rollback(context.Background())

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	var userId uint64
	err = conn.QueryRow(context.Background(), "SELECT id FROM public.users WHERE lower(nickname) = $1", strings.ToLower(vote.Nickname)).Scan(&userId)
	if err != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find user by nickname: " + vote.Nickname}
		response, _ = json.Marshal(error)
		return 404, response
	}

	selector, err := strconv.Atoi(slug)
	var flag bool

	if err != nil {
		flag = true
	} else {
		flag = false
	}

	var threadSlug interface{}
	var thread Thread
	if flag {
		err = conn.QueryRow(context.Background(), "SELECT id, votes, title, users_nickname, forum, slug, message, created FROM public.thread WHERE lower(slug) = $1", strings.ToLower(slug)).Scan(&thread.Id, &thread.Votes, &thread.Title, &thread.Author, &thread.Forum, &threadSlug, &thread.Message, &thread.Created)
	} else {
		err = conn.QueryRow(context.Background(), "SELECT id, votes, title, users_nickname, forum, slug, message, created FROM public.thread WHERE id = $1", selector).Scan(&thread.Id, &thread.Votes, &thread.Title, &thread.Author, &thread.Forum, &threadSlug, &thread.Message, &thread.Created)
	}
	if err != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}
	thread.Slug = fmt.Sprintf("%v", threadSlug)

	var currentVote CurrentVote
	err = conn.QueryRow(context.Background(), "SELECT id, voice FROM public.vote WHERE thread_id = $1 AND lower(nickname) = $2", &thread.Id, strings.ToLower(vote.Nickname)).Scan(&currentVote.Id, &currentVote.Voice)

	var row pgx.Rows
	if err != nil {
		row, err = conn.Query(context.Background(),
			"INSERT INTO public.vote (voice, nickname, thread_id) VALUES ($1, $2, $3) RETURNING id",
			vote.Voice, vote.Nickname, thread.Id)
		row.Close()
		thread.Votes = thread.Votes + vote.Voice
	} else {
		thread.Votes = thread.Votes - currentVote.Voice
		thread.Votes = thread.Votes + vote.Voice
		row, err =  conn.Query(context.Background(), "UPDATE public.vote SET voice = $2 WHERE id = $1", currentVote.Id, vote.Voice)
		row.Close()
	}
	//thread.Created = thread.Created.Add(-3 * time.Hour)

	row, err = conn.Query(context.Background(), "UPDATE public.thread SET votes = $2 WHERE id = $1", thread.Id, thread.Votes)

	response, _ = json.Marshal(thread)
	return 200, response
}

//CreateUser
func InsertCreateUser(p *pgxpool.Pool, user User) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	var existUserNick User
	var existUserEmail User
	err = conn.QueryRow(context.Background(), "SELECT about, email, fullname, nickname FROM public.users WHERE lower(nickname) = $1", strings.ToLower(user.Nickname)).Scan(&existUserNick.About, &existUserNick.Email, &existUserNick.Fullname, &existUserNick.Nickname)
	err = conn.QueryRow(context.Background(), "SELECT about, nickname, fullname, email FROM public.users WHERE lower(email) = $1", strings.ToLower(user.Email)).Scan(&existUserEmail.About, &existUserEmail.Nickname, &existUserEmail.Fullname, &existUserEmail.Email)

	if (existUserNick.Nickname != "" || existUserEmail.Email != "") {
		var users Users

		if (existUserNick.Nickname == existUserEmail.Nickname) {
			users = append(users, existUserNick)
			response, _ = json.Marshal(users)
			return 409, response
		}

		if (existUserNick.Nickname != "") {
			users = append(users, existUserNick)
		}

		if (existUserEmail.Email != "") {
			users = append(users, existUserEmail)
		}

		response, _ = json.Marshal(users)
		return 409, response
	}

	conn.QueryRow(context.Background(),
		"INSERT INTO public.users (about, email, fullname, nickname) VALUES ($1, $2, $3, $4) RETURNING id",
		user.About, user.Email, user.Fullname, user.Nickname)

	response, _ = json.Marshal(user)
	return 201, response
}


//getUser
func getUser(p *pgxpool.Pool, nickname string) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	var user User
	err = conn.QueryRow(context.Background(), "SELECT about, email, fullname, nickname FROM public.users WHERE lower(nickname) = $1", strings.ToLower(nickname)).Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname)
	if err != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}

	response, _ = json.Marshal(user)
	return 200, response
}


//updateUser
func updateUser(p *pgxpool.Pool, user User) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	var currentUser User
	err = conn.QueryRow(context.Background(), "SELECT nickname, email, fullname, about FROM public.users WHERE lower(nickname) = $1", strings.ToLower(user.Nickname)).Scan(&currentUser.Nickname, &currentUser.Email, &currentUser.Fullname, &currentUser.About)
	if err != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}

	var existUserEmail User
	err = conn.QueryRow(context.Background(), "SELECT nickname FROM public.users WHERE lower(email) = $1", strings.ToLower(user.Email)).Scan(&existUserEmail.Nickname)
	if (existUserEmail.Nickname != "") {
		//log.Errorf("This email is already registered by user: " + existUserEmail.Nickname, err)

		error := Error{"This email is already registered by user: " + existUserEmail.Nickname}

		response, _ = json.Marshal(error)
		return 409, response
	}

	if (user.About != "" && user.Fullname != "" && user.Email != "") {
		conn.QueryRow(context.Background(), "UPDATE public.users SET about = $2, email = $3, fullname = $4 WHERE nickname = $1", user.Nickname, user.About, user.Email, user.Fullname)
	}

	if (user.About != "" && user.Fullname == "" && user.Email == "") {
		conn.QueryRow(context.Background(), "UPDATE public.users SET about = $2 WHERE nickname = $1", user.Nickname, user.About)
		user.Fullname = currentUser.Fullname
		user.Email = currentUser.Email
	}

	if (user.About == "" && user.Fullname != "" && user.Email == "") {
		conn.QueryRow(context.Background(), "UPDATE public.users SET fullname = $2 WHERE nickname = $1", user.Nickname, user.Fullname)
		user.About = currentUser.About
		user.Email = currentUser.Email
	}

	if (user.About == "" && user.Fullname == "" && user.Email != "") {
		conn.QueryRow(context.Background(), "UPDATE public.users SET email = $2 WHERE nickname = $1", user.Nickname, user.Email)
		user.About = currentUser.About
		user.Fullname = currentUser.Fullname
	}

	if (user.About != "" && user.Fullname != "" && user.Email == "") {
		conn.QueryRow(context.Background(), "UPDATE public.users SET about = $2, fullname = $3 WHERE nickname = $1", user.Nickname, user.About, user.Fullname)
		user.Email = currentUser.Email
	}

	if (user.About != "" && user.Fullname == "" && user.Email != "") {
		conn.QueryRow(context.Background(), "UPDATE public.users SET about = $2, email = $3 WHERE nickname = $1", user.Nickname, user.About, user.Email)
		user.Fullname = currentUser.Fullname
	}

	if (user.About == "" && user.Fullname != "" && user.Email != "") {
		conn.QueryRow(context.Background(), "UPDATE public.users SET email = $2, fullname = $3 WHERE nickname = $1", user.Nickname, user.Email, user.Fullname)
		user.About = currentUser.About
	}

	if (user.About == "" && user.Fullname != "" && user.Email != "") {
		conn.QueryRow(context.Background(), "UPDATE public.users SET email = $2, fullname = $3 WHERE nickname = $1", user.Nickname, user.Email, user.Fullname)
		user.About = currentUser.About
	}

	if (user.About == "" && user.Fullname == "" && user.Email == "") {
		user.About = currentUser.About
		user.Email = currentUser.Email
		user.Fullname = currentUser.Fullname
	}

	response, _ = json.Marshal(user)
	return 200, response
}


//getPostDetails
func getPostDetails(p *pgxpool.Pool, id string, related string) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	var threadRoute bool
	var forumRoute bool
	var userRoute bool

	if related != "" {
		threadRoute = false
		forumRoute = false
		userRoute = false
	}

	if related == "user" {
		threadRoute = false
		forumRoute = false
		userRoute = true
	}

	if related == "thread" {
		threadRoute = true
		forumRoute = false
		userRoute = false
	}

	if related == "forum" {
		threadRoute = false
		forumRoute = true
		userRoute = false
	}

	if related == "user,thread" {
		threadRoute = true
		forumRoute = false
		userRoute = true
	}

	if related == "user,forum" {
		threadRoute = false
		forumRoute = true
		userRoute = true
	}

	if related == "thread,forum" {
		threadRoute = true
		forumRoute = true
		userRoute = false
	}

	if related == "user,thread,forum" {
		threadRoute = true
		forumRoute = true
		userRoute = true
	}

	selector, err := strconv.Atoi(id)
	if err != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}

	var post Post
	var user User
	err = conn.QueryRow(context.Background(), "SELECT id, parent, users_nickname, message, isedited, forum, thread_id, created, users_fullname, users_email, users_about FROM public.post WHERE id = $1", selector).Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created, &user.Fullname, &user.Email, &user.About)
	if err != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}

	user.Nickname = post.Author

	var thread Thread
	if threadRoute {
		var slugInter interface{}
		err = conn.QueryRow(context.Background(), "SELECT id, title, users_nickname, message, votes, created, slug, forum  FROM public.thread WHERE id = $1", post.Thread).Scan(&thread.Id, &thread.Title, &thread.Author, &thread.Message, &thread.Votes, &thread.Created, &slugInter, &thread.Forum)
		if err != nil {
			//log.Errorf("Unable to acquire a database connection: %v\n", err)
			error := Error{"Can't find thread" + thread.Author + thread.Message + thread.Title + fmt.Sprintf("%v", slugInter) + thread.Forum}
			response, _ = json.Marshal(error)
			return 404, response
		}
		//thread.Created = thread.Created.Add(-3 * time.Hour)
		thread.Slug = fmt.Sprintf("%v", slugInter)
	}

	var forum ForumDetails
	if forumRoute {
		var userId uint64
		err = conn.QueryRow(context.Background(), "SELECT title, slug, posts, threads, user_id FROM public.forum WHERE lower(slug) = $1", strings.ToLower(post.Forum)).Scan(&forum.Title, &forum.Slug, &forum.Posts, &forum.Threads, &userId)
		if err != nil {
			//log.Errorf("Unable to acquire a database connection: %v\n", err)
			error := Error{"Can't find user with id #\n"}
			response, _ = json.Marshal(error)
			return 404, response
		}
		conn.QueryRow(context.Background(), "SELECT nickname FROM public.users WHERE id = $1", userId).Scan(&forum.User)
	}

	//var postForDetails PostForDetails
	//postForDetails.Author = post.Author
	//currentCreated := post.Created.String()
	//currentCreated = currentCreated[:10] + "T" + currentCreated[11:]
	//postForDetails.Created = currentCreated[:23] + "+03:00"
	//postForDetails.Forum = post.Forum
	//postForDetails.Id = post.Id
	//postForDetails.Message = post.Message
	//postForDetails.Parent = post.Parent
	//postForDetails.Thread = post.Thread
	//postForDetails.IsEdited = post.IsEdited

	var postPost PostPost
	var postAuthor PostAuthor
	var postThread PostThread
	var postForum PostForum
	var postForumThread PostForumThread
	var postForumUser PostForumUser
	var postThreadUser PostThreadUser
	var postFullDetails PostFullDetails

	if related == "" {
		postPost.Post = post
		response, _ = json.Marshal(postPost)
	}

	if userRoute && !threadRoute && !forumRoute {
		postAuthor.Post = post
		postAuthor.Author = user
		response, _ = json.Marshal(postAuthor)
	}

	if threadRoute && !userRoute && !forumRoute {
		postThread.Post = post
		postThread.Thread = thread
		response, _ = json.Marshal(postThread)
	}

	if forumRoute && !userRoute && !threadRoute {
		postForum.Post = post
		postForum.Forum = forum
		response, _ = json.Marshal(postForum)
	}

	if forumRoute && userRoute && !threadRoute {
		postForumUser.Post = post
		postForumUser.Forum = forum
		postForumUser.Author = user
		response, _ = json.Marshal(postForumUser)
	}

	if forumRoute && !userRoute && threadRoute {
		postForumThread.Post = post
		postForumThread.Forum = forum
		postForumThread.Thread = thread
		response, _ = json.Marshal(postForumThread)
	}

	if !forumRoute && userRoute && threadRoute {
		postThreadUser.Post = post
		postThreadUser.Author = user
		postThreadUser.Thread = thread
		response, _ = json.Marshal(postThreadUser)
	}

	if forumRoute && userRoute && threadRoute {
		postFullDetails.Post = post
		postFullDetails.Author = user
		postFullDetails.Thread = thread
		postFullDetails.Forum = forum
		response, _ = json.Marshal(postFullDetails)
	}

	return 200, response
}


func updatePost(p *pgxpool.Pool, id string, postUpdate PostUpdate) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	selector, err := strconv.Atoi(id)
	if err != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}

	postUpdate.Id = uint64(selector)

	var post Post
	err = conn.QueryRow(context.Background(), "SELECT id, parent, users_nickname, message, isedited, forum, thread_id, created FROM public.post WHERE id = $1", selector).Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created)
	if err != nil {
		//log.Errorf("Unable to acquire a database connection: %v\n", err)
		error := Error{"Can't find user with id #\n"}
		response, _ = json.Marshal(error)
		return 404, response
	}

	if postUpdate.Message != "" && post.Message != postUpdate.Message {
		conn.QueryRow(context.Background(), "UPDATE public.post SET message = $2, isedited = $3 WHERE id = $1", post.Id, postUpdate.Message, true)
	}

	var postForDetailsPost Post
	postForDetailsPost.Id = post.Id
	postForDetailsPost.Author = post.Author
	postForDetailsPost.Forum = post.Forum
	postForDetailsPost.Thread = post.Thread
	postForDetailsPost.Created = post.Created

	if postUpdate.Message == "" {
		postForDetailsPost.Message = post.Message
	} else {
		if post.Message != postUpdate.Message {
			postForDetailsPost.IsEdited = true
		}
		postForDetailsPost.Message = postUpdate.Message
	}
	//currentCreated := post.Created.String()
	//currentCreated = currentCreated[:10] + "T" + currentCreated[11:]
	//postForDetailsPost.Created = currentCreated[:23] + "+03:00"

	response, _ = json.Marshal(postForDetailsPost)
	return 200, response
}


//addPostThread
func addPostThread(p *pgxpool.Pool, slugOrId string, posts Posts) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	selector, err := strconv.Atoi(slugOrId)
	var flag bool

	if err != nil {
		flag = true
	} else {
		flag = false
	}

	var thread Thread
	if flag {
		err = conn.QueryRow(context.Background(), "SELECT id, forum FROM public.thread WHERE lower(slug) = $1", strings.ToLower(slugOrId)).Scan(&thread.Id, &thread.Forum)
	} else {
		err = conn.QueryRow(context.Background(), "SELECT id, forum FROM public.thread WHERE id = $1", selector).Scan(&thread.Id, &thread.Forum)
	}
	if err != nil {
		if flag {
			error := Error{"Can't find post thread by slug: "+ slugOrId}
			response, _ = json.Marshal(error)
			return 404, response
		} else {
			error := Error{"Can't find post thread by id: "+ slugOrId}
			response, _ = json.Marshal(error)
			return 404, response
		}
	}

	currentTime := time.Now()

	var ids []uint64
	rows, err := conn.Query(context.Background(), "SELECT id FROM public.post WHERE thread_id = $1", thread.Id)
	for rows.Next() {
		var id uint64
		if err = rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}

	for i := 0; i < len(posts); i++ {

		err = conn.QueryRow(context.Background(), "SELECT id, about, email, fullname FROM public.users WHERE lower(nickname) = $1", strings.ToLower(posts[i].Author)).Scan(&posts[i].UserId, &posts[i].About, &posts[i].Email, &posts[i].Fullname)
		if err != nil {
			//log.Errorf("Unable to acquire a database connection: %v\n", err)
			error := Error{"Can't find user with id #\n"}
			response, _ = json.Marshal(error)
			return 404, response
		}

		isContinue := false
		if posts[i].Parent != 0 {
			for j := 0; j < len(ids); j++ {
				if ids[j] == posts[i].Parent {
					isContinue  = true
					break
				}
			}
		} else {
			isContinue  = true
		}

		if !isContinue {
			error := Error{"Can't find user with id #\n"}
			response, _ = json.Marshal(error)
			return 409, response
		}

		if err != nil {
			//log.Errorf("Unable to acquire a database connection: %v\n", err)
			error := Error{"Can't find user with id #\n"}
			response, _ = json.Marshal(error)
			return 404, response
		}

		posts[i].Forum = thread.Forum
		posts[i].Thread = thread.Id
		posts[i].IsEdited = false
		posts[i].Created = currentTime

		var row pgx.Row
		row = conn.QueryRow(context.Background(),
			"INSERT INTO public.post (forum, thread_id, created, isedited, parent, users_nickname, message, users_email, users_fullname, users_about, user_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id",
			posts[i].Forum, posts[i].Thread, posts[i].Created, posts[i].IsEdited, posts[i].Parent, posts[i].Author, posts[i].Message, posts[i].Email, posts[i].Fullname, posts[i].About, posts[i].UserId)

		err = row.Scan(&posts[i].Id)
	}

	conn.QueryRow(context.Background(), "UPDATE public.forum SET posts = posts + $2 WHERE slug = $1", thread.Forum, len(posts))

	response, _ = json.Marshal(posts)
	return 201, response
}

// getService
func getService(p *pgxpool.Pool) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()

	var status Status

	conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM public.forum").Scan(&status.Forum)
	conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM public.post").Scan(&status.Post)
	conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM public.thread").Scan(&status.Thread)
	conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM public.users").Scan(&status.User)

	response, _ = json.Marshal(status)
	return 200, response
}

// deleteService
func deleteService(p *pgxpool.Pool) (int, []byte) {

	response := []byte(`{"Error":"Unable to acquire a database connection"}`)

	conn, err := p.Acquire(context.Background())
	if err != nil {
		log.Errorf("Unable to acquire a database connection: %v\n", err)
		return 500, response
	}
	defer conn.Release()


	//row, err = conn.Query(context.Background(),
	//	"INSERT INTO public.vote (voice, nickname, thread_id) VALUES ($1, $2, $3) RETURNING id",
	//	vote.Voice, vote.Nickname, thread.Id)
	//row.Close()

	var row pgx.Rows

	row, err = conn.Query(context.Background(), "TRUNCATE public.vote CASCADE")
	row.Close()

	row, err = conn.Query(context.Background(), "TRUNCATE public.post CASCADE")
	row.Close()

	row, err = conn.Query(context.Background(), "TRUNCATE public.thread CASCADE")
	row.Close()

	row, err = conn.Query(context.Background(), "TRUNCATE public.forum CASCADE")
	row.Close()

	row, err = conn.Query(context.Background(), "TRUNCATE public.users CASCADE")
	row.Close()

	response, _ = json.Marshal("Очистка базы успешно завершена")
	return 200, response
}

func main() {
	//database
	//const dsn = "user=mark host=localhost port=5432 dbname=mark pool_max_conns=30 slmode=disable"
	const dsn = "user=root host=localhost port=5432 dbname=root pool_max_conns=30 slmode=disable"
	pool, err := pgxpool.Connect(context.Background(), os.Getenv(dsn))
	if err != nil {
		log.Fatalf("Unable to connection to database: %v\n", err)
	}
	defer pool.Close()
	log.Infof("Connected!")

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Fatalf("Unable to acquire a database connection: %v\n", err)
	}
	conn.Release()

	//config requests
	config := NewRequestHandler(pool)

	// router
	router := router.New()
	router.POST("/api/forum/create", config.createForum)
	router.GET("/api/forum/{slug}/details", config.getForumDetails)
	router.POST("/api/forum/{slug}/create", config.createForumThread)
	router.GET("/api/forum/{slug}/users", config.getForumUsers)
	router.GET("/api/forum/{slug}/threads", config.getForumThreads)
	router.GET("/api/post/{id}/details", config.getPostDetails)
	router.POST("/api/post/{id}/details", config.updatePost)

	router.POST("/api/service/clear", config.deleteService)
	router.GET("/api/service/status", config.getService)
	router.POST("/api/thread/{slug_or_id}/create", config.addPostThread)
	router.GET("/api/thread/{slug_or_id}/details", config.getThreadDetails)
	router.POST("/api/thread/{slug_or_id}/details", config.updateThreadDetails)

	router.GET("/api/thread/{slug_or_id}/posts", config.getPostThread)
	router.POST("/api/thread/{slug_or_id}/vote", config.addVoteThread)
	router.POST("/api/user/{nickname}/create", config.createUser)
	router.GET("/api/user/{nickname}/profile", config.getUser)
	router.POST("/api/user/{nickname}/profile", config.updateUser)


	server := fasthttp.Server{Handler: router.Handler}
	fmt.Printf("Server in %s", address)

	err = server.ListenAndServe(address)
	if err != nil {
		fmt.Println("error in ListenAndServe: %s", err)
		fmt.Println()
	}
}