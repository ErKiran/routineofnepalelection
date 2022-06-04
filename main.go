package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
	gim "github.com/ozankasikci/go-image-merge"
)

type ElectionData struct {
	Candidatename      string `json:"candidateName"`
	Candidatepartyname string `json:"candidatePartyName"`
	Votenumbers        string `json:"voteNumbers"`
	Candidateimage     string `json:"candidateImage"`
	Electionicon       string `json:"electionIcon"`
	Title              string `json:"title"`
}

var localBodyMap = map[string]string{
	"kathmandu": "https://election.ekantipur.com/pradesh-3/district-kathmandu/kathmandu?lng=eng",
	"bharatpur": "https://election.ekantipur.com/pradesh-3/district-chitwan/bharatpur?lng=eng",
	"lalitpur":  "https://election.ekantipur.com/pradesh-3/district-lalitpur/lalitpur?lng=eng",
	"pokhara":   "https://election.ekantipur.com/pradesh-4/district-kaski/pokhara-lekhnath?lng=eng",
	"hetauda":   "https://election.ekantipur.com/pradesh-3/district-makwanpur/hetauda?lng=eng",
	"janakpur":  "https://election.ekantipur.com/pradesh-2/district-dhanusha/janakpurdham?lng=eng",
	"butwal":    "https://election.ekantipur.com/pradesh-5/district-rupandehi/butwal?lng=eng",
	"nepalgunj": "https://election.ekantipur.com/pradesh-5/district-banke/nepalgunj?lng=eng",
	"ghorahi":   "https://election.ekantipur.com/pradesh-5/district-dang/ghorahi?lng=eng",
	"dhangadi":  "https://election.ekantipur.com/pradesh-7/district-kailali/dhangadhi?lng=eng",
	"birgunj":   "https://election.ekantipur.com/pradesh-2/district-parsa/birgunj?lng=eng",
	"dharan":    "https://election.ekantipur.com/pradesh-1/district-sunsari/dharan?lng=eng",
	"dhulikhel": "https://election.ekantipur.com/pradesh-3/district-kavrepalanchowk/dhulikhel?lng=eng",
	"banepa":    "https://election.ekantipur.com/pradesh-3/district-kavrepalanchowk/banepa?lng=eng",
}

func ReadAndParseData(city string) ([]ElectionData, error) {
	jsonFile, err := os.Open(fmt.Sprintf("data/%s.json", city))

	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var electionData []ElectionData

	err = json.Unmarshal(byteValue, &electionData)

	if err != nil {
		return nil, err
	}

	return electionData, nil
}

func main() {
	for {
		for city, _ := range localBodyMap {
			err := exec.Command("/usr/local/bin/python3", "scraper.py", city).Run()
			if err != nil {
				fmt.Println("err", err)
			}
			data, err := ReadAndParseData(city)
			if err != nil {
				fmt.Println(err)
			}

			var grids []*gim.Grid
			var title string
			MakeFolder("city")
			MakeFolder("original")
			for _, d := range data {
				title = d.Title
				_, err := downloadFile(d.Candidateimage, fmt.Sprintf("%s/", "original")+d.Candidatename+".png")
				if err != nil {
					fmt.Println(err)
				}

				var fileName string

				err = filepath.Walk("original", func(path string, info os.FileInfo, err error) error {
					if err != nil {
						fmt.Println(err)
						return nil
					}

					split := strings.Split(path, "/")
					if len(split) > 1 {
						candidateName := strings.Split(split[1], ".")[0]
						if candidateName == d.Candidatename {
							fileName = path
						}
					}
					return nil
				})

				if err != nil {
					fmt.Println(err)
				}

				if fileName == "" {
					fmt.Println("downloading from server", d.Candidateimage, d.Candidatename, d.Title)
					fileName, err = downloadFile(d.Candidateimage, fmt.Sprintf("%s/", "city")+d.Candidatename+".png")
					if err != nil {
						log.Fatal(err)
					}
				}
				ResizeImage(fileName)
				EditImage(fileName, d)
				grids = append(grids, &gim.Grid{ImageFilePath: fileName})
			}

			rgba, err := gim.New(grids, 3, 1).Merge()

			if err != nil {
				fmt.Println(err)
			}

			// save the output to jpg or png
			file, err := os.Create(fmt.Sprintf("%s.png", "city"))
			if err != nil {
				fmt.Println(err)
			}
			err = jpeg.Encode(file, rgba, &jpeg.Options{Quality: 80})
			if err != nil {
				fmt.Println(err)
			}
			err = png.Encode(file, rgba)
			if err != nil {
				fmt.Println(err)
			}
			EditFinalImage(fmt.Sprintf("%s.png", "city"), title)

			err = exec.Command("/usr/local/bin/python3", "fbpost.py", city).Run()
			if err != nil {
				fmt.Println("fucking error", err)
			}

			if city == "kathmandu" {
				time.Sleep(time.Second * 60)
			}

			if city == "bharatpur" || city == "hetauda" || city == "pokhara" {
				time.Sleep(time.Second * 30)
			}
			time.Sleep(time.Second * 30)
		}
		time.Sleep(time.Minute * 10)
	}

}

func downloadFile(URL, fileName string) (string, error) {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", errors.New("received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}

	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

func MakeFolder(fileName string) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		err := os.MkdirAll(fileName, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
}

func EditImage(image string, d ElectionData) {
	const S = 1024

	im, err := gg.LoadImage(image)

	if err != nil {
		log.Fatal(err)
	}

	dc := gg.NewContext(S/2, S-200)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	if err := dc.LoadFontFace("GothicA1-Regular.ttf", 36); err != nil {
		fmt.Println(err)
	}
	dc.DrawStringAnchored(d.Candidatename, S/4, S-390, 0.5, 0.5)

	if err := dc.LoadFontFace("GothicA1-Regular.ttf", 24); err != nil {
		fmt.Println(err)
	}

	dc.DrawStringAnchored(fmt.Sprintf("(%s)", d.Candidatepartyname), S/4, S-350, 0.5, 0.5)

	if err := dc.LoadFontFace("GothicA1-Regular.ttf", 72); err != nil {
		fmt.Println(err)
	}

	dc.DrawStringAnchored(d.Votenumbers, S/4, S-250, 0.5, 0.5)

	dc.DrawRoundedRectangle(0, 0, 512, 512, 0)
	dc.DrawImage(im, 0, 0)
	// dc.DrawStringAnchored("Balen Shah (Independent)", S/4, S-390, 0.5, 0.5)
	dc.Clip()
	dc.SavePNG(image)
}

func EditFinalImage(image, text string) {
	const H = 824
	const W = 1536

	im, err := gg.LoadImage(image)

	if err != nil {
		log.Fatal(err)
	}

	dc := gg.NewContext(W, H+200)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	if err := dc.LoadFontFace("GothicA1-Regular.ttf", 56); err != nil {
		fmt.Println(err)
	}
	dc.DrawStringAnchored(text, 768, 888, 0.5, 0.5)

	dc.DrawRoundedRectangle(0, 0, 512, 512, 0)
	dc.DrawImage(im, 0, 0)
	dc.Clip()
	dc.SavePNG(image)
}

func ResizeImage(fileName string) {
	// open "test.jpg"
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}

	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(588, 588, img, resize.Lanczos3)

	out, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// write new image to file
	jpeg.Encode(out, m, nil)
}
