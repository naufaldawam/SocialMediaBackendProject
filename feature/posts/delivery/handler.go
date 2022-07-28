package delivery

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"socialmediabackendproject/config"
	"socialmediabackendproject/domain"
	"socialmediabackendproject/feature/common"
	"socialmediabackendproject/feature/middlewares"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type postHandler struct {
	PostUsecase domain.PostUsecase
}

func New(e *echo.Echo, ps domain.PostUsecase) {
	handler := &postHandler{
		PostUsecase: ps,
	}
	useJWT := middleware.JWTWithConfig(middlewares.UseJWT([]byte(config.SECRET)))
	e.GET("/posts", handler.GetAllPosts())
	e.POST("/myposts", handler.InsertPost(), useJWT)
	e.GET("/myposts", handler.GetAllMyPosts(), useJWT)
}

func (ph *postHandler) GetAllPosts() echo.HandlerFunc {
	return func(c echo.Context) error {
		data, username, postimages, err := ph.PostUsecase.GetAllPosts()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		var GetAllPostsResponse []GetAllPost
		for i := 0; i < len(data); i++ {
			GetAllPostsResponse = append(GetAllPostsResponse, GetAllPost{
				ID:                   data[i].ID,
				User_ID:              data[i].User_ID,
				Username:             username[i].Name,
				Profile_picture_path: username[i].Profile_picture_path,
				Caption:              data[i].Caption,
				Created_At:           data[i].Created_At,
				Updated_At:           data[i].Updated_At,
				Post_Images:          postimages[i],
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "success get all data",
			"data":    GetAllPostsResponse,
		})
	}
}

func (ph *postHandler) InsertPost() echo.HandlerFunc {
	return func(c echo.Context) error {
		var dataPost domain.Post
		caption := c.FormValue("caption")
		if caption == "" {
			return c.JSON(http.StatusBadRequest, "caption cant be empty")
		}
		dataPost.Caption = caption

		id := common.ExtractData(c)
		data, err := ph.PostUsecase.AddPost(uint(id), dataPost)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		// Multipart form
		form, err := c.MultipartForm()
		if err != nil {
			return err
		}
		files := form.File["post_images"]

		for key, file := range files {
			// Source
			src, err := file.Open()
			if err != nil {
				return err
			}
			defer src.Close()

			// Destination
			path := fmt.Sprint("uploads/postimages/", data.ID, "-", strconv.Itoa(key+1), "-", file.Filename)
			dst, err := os.Create(path)
			if err != nil {
				return err
			}
			defer dst.Close()

			// Copy
			if _, err = io.Copy(dst, src); err != nil {
				return err
			}

		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "success insert post",
			"data":    data,
		})
	}
}

func (ph *postHandler) GetAllMyPosts() echo.HandlerFunc {
	return func(c echo.Context) error {
		id := common.ExtractData(c)
		posts, userdata, postimages, err := ph.PostUsecase.GetMyPosts(uint(id))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		var GetAllMyPostsResponse []GetAllPost
		for i := 0; i < len(posts); i++ {
			GetAllMyPostsResponse = append(GetAllMyPostsResponse, GetAllPost{
				ID:                   posts[i].ID,
				User_ID:              posts[i].User_ID,
				Username:             userdata.Name,
				Profile_picture_path: userdata.Profile_picture_path,
				Caption:              posts[i].Caption,
				Created_At:           posts[i].Created_At,
				Updated_At:           posts[i].Updated_At,
				Post_Images:          postimages[i],
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "success get all my posts",
			"data":    GetAllMyPostsResponse,
		})
	}
}
