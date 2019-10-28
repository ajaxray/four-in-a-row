package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/faiface/pixel"
	"github.com/tidwall/gjson"
)

const unsplashAccessKey = "20f850349105e34e430134108aa8afca97eb82499dbc584a554a6e6f4a372fd9"

//const unsplashCollectionID = 8823531
const unsplashBaseURL = "https://api.unsplash.com"

var onlineBackgrounds []string

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
	var err error

	if len(onlineBackgrounds) == 0 {
		onlineBackgrounds, err = loadCollectionPhotos(8823531, "regular")
	}
	//fmt.Printf("%+v \n", onlineBackgrounds)
	if err == nil && len(onlineBackgrounds) > 0 {
		rand.Seed(time.Now().UnixNano())
		selectedBack := onlineBackgrounds[rand.Intn(len(onlineBackgrounds))]

		if err == nil {
			if back, err := loadPictureURL(selectedBack); err == nil {
				return pixel.NewSprite(back, back.Bounds())
			}
		}
		//fmt.Printf("Error: %s \n", err)
	}
	return nil
}

// DownloadToTemp will download a url to a temporary local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
// Keep dir empty to generate in OS temporary directory
// returns: created filename, file closing func, error
func downloadToTemp(dir, filepathPattern, url string) (string, func() error, error) {
	file, err := ioutil.TempFile(dir, filepathPattern)
	if err != nil {
		return "", file.Close, err
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", file.Close, err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(file, resp.Body)
	return file.Name(), file.Close, err
}
