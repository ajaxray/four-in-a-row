package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/faiface/pixel"
	"github.com/tidwall/gjson"
)

var unsplashBaseURL, unsplashAccessKey string
var onlineBackgrounds []string

func prepareUnsplash(apiURL, accessKey string, collectionID int) bool {
	unsplashBaseURL = apiURL
	unsplashAccessKey = accessKey

	if accessKey == "" || collectionID == 0 {
		return false
	}

	onlineBackgrounds, _ = loadCollectionPhotos(collectionID, "regular")
	return (len(onlineBackgrounds) > 0)
}

func loadCollectionPhotos(id int, size string) ([]string, error) {
	response, err := http.Get(fmt.Sprintf("%s/collections/%d/photos?client_id=%s", unsplashBaseURL, id, unsplashAccessKey))
	if err == nil {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)

		if err == nil {
			var urls []string
			imgLinks := gjson.GetBytes(contents, "#.urls."+size)
			imgLinks.ForEach(func(_, value gjson.Result) bool {
				urls = append(urls, value.String())
				return true // keep iterating
			})

			return urls, nil
		}
	}

	// fmt.Printf("%s\n", err)
	return []string{}, err
}

func loadUnsplashBackground() *pixel.Sprite {
	//fmt.Printf("%+v \n", onlineBackgrounds)
	if len(onlineBackgrounds) > 0 {
		rand.Seed(time.Now().UnixNano())
		selectedBack := onlineBackgrounds[rand.Intn(len(onlineBackgrounds))]

		if back, err := loadPictureURL(selectedBack); err == nil {
			return pixel.NewSprite(back, back.Bounds())
		}
		//fmt.Printf("Error: %s \n", err)
	}
	return nil
}
