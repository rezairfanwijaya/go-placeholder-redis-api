package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

func main() {
	e := echo.New()

	e.GET("/photos", GetAllPhotosWithoutCache)
	e.GET("/photos/cache", GetAllPhotosWithCache)
	if err := e.Start("localhost:9090"); err != nil {
		log.Fatalf("failed start server, err: %s", err)
	}
}

type Photo struct {
	AlbumID      int    `json:"alumniId"`
	ID           int    `json:"id"`
	Title        string `json:"title"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnailUrl"`
}

func GetAllPhotosWithoutCache(c echo.Context) error {
	client := http.DefaultClient
	req, err := http.NewRequest(http.MethodGet, "https://jsonplaceholder.typicode.com/photos", nil)
	if err != nil {
		sendErr(c, err)
	}

	res, err := client.Do(req)
	if err != nil {
		sendErr(c, err)
	}

	photos := []Photo{}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&photos)
	if err != nil {
		sendErr(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": photos,
	})
}

func GetAllPhotosWithCache(c echo.Context) error {
	ctx := context.Background()
	photos := []Photo{}
	rc := redis.NewClient(&redis.Options{
		Addr: "localhost:9876",
		DB:   0,
	})
	defer rc.Close()

	client := http.DefaultClient
	req, err := http.NewRequest(http.MethodGet, "https://jsonplaceholder.typicode.com/photos", nil)
	if err != nil {
		log.Println("failed when create new request, err: ", err)
		return sendErr(c, err)
	}

	// return data from chache when key exist
	resCache, err := rc.Get(ctx, "photos").Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Println("failed when get keys, err: ", err)
			return sendErr(c, err)
		}
	}

	if resCache != "" {
		err := json.Unmarshal([]byte(resCache), &photos)
		if err != nil {
			log.Println("failed to unmarshal body response, err: ", err)
			return sendErr(c, err)
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code": http.StatusOK,
			"data": photos,
		})
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println("failed when get http response, err: ", err)
		return sendErr(c, err)
	}

	bd, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("failed when read body response, err: ", err)
		return sendErr(c, err)
	}

	err = rc.Set(ctx, "photos", string(bd), 0).Err()
	if err != nil {
		log.Println("failed when set keys, err: ", err)
		return sendErr(c, err)
	}

	err = json.Unmarshal(bd, &photos)
	if err != nil {
		log.Println("failed to unmarshal body response, err: ", err)
		return sendErr(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": photos,
	})
}

func sendErr(c echo.Context, err error) error {
	return c.JSON(http.StatusInternalServerError, map[string]interface{}{
		"error": err,
		"code":  http.StatusInternalServerError,
	})
}
