package gallery

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const SLASH = "/"

const (
	THUMBS      = "thumbs"
	EMBEDS      = "embeds"
	CONCURRENCY = 5
)

var (
	rx_img_jpg *regexp.Regexp
)

func init() {
	rx_img_jpg, _ = regexp.Compile(`(?i)\.(jpg|jpeg)`)
}

func lsImages(path string) (images []string) {
	var wg sync.WaitGroup

	os.Mkdir(path+SLASH+THUMBS, 755)
	os.Mkdir(path+SLASH+EMBEDS, 755)

	files, _ := ioutil.ReadDir(path)
	for i := 0; i < len(files); i += CONCURRENCY {

		// n goroutine at a time
		for j := 0; j < CONCURRENCY; j++ {

			if i+j < len(files) && rx_img_jpg.MatchString(files[i+j].Name()) {
				images = append(images, files[i+j].Name())
				// does file exist? if not, create thumb
				if _, err := os.Stat(path + SLASH + THUMBS + SLASH + files[i+j].Name()); err != nil {
					wg.Add(1)
					go fmt.Println(generateThumb(path, files[i+j].Name(), &wg))
				}
			}

		}

		wg.Wait()
	}

	return
}

func generateThumb(path, filename string, wg *sync.WaitGroup) (err error) {
	var (
		file      *os.File
		img       image.Image
		conf      image.Config
		f_thumb   *os.File
		f_embed   *os.File
		img_thumb image.Image
		img_embed image.Image
	)

	defer wg.Done()

	// open image
	if file, err = os.Open(path + SLASH + filename); err != nil {
		return
	}
	if conf, err = jpeg.DecodeConfig(file); err != nil {
		return
	}
	file.Close()

	// decode jpeg into image.Image
	if file, err = os.Open(path + SLASH + filename); err != nil {
		return
	}
	if img, err = jpeg.Decode(file); err != nil {
		return
	}
	file.Close()

	// resize to width / hight using Lanczos resampling
	// and preserve aspect ratio
	if conf.Width > conf.Height {
		img_thumb = resize.Resize(0, 100, img, resize.NearestNeighbor)
		img_embed = resize.Resize(1024, 0, img, resize.NearestNeighbor)
	} else {
		img_thumb = resize.Resize(100, 0, img, resize.NearestNeighbor)
		img_embed = resize.Resize(0, 1024, img, resize.NearestNeighbor)
	}

	if f_thumb, err = os.Create(path + SLASH + THUMBS + SLASH + filename); err != nil {
		return
	}
	if f_embed, err = os.Create(path + SLASH + EMBEDS + SLASH + filename); err != nil {
		return
	}
	defer f_thumb.Close()
	defer f_embed.Close()

	// write new image to file
	jpeg.Encode(f_thumb, img_thumb, nil)
	jpeg.Encode(f_embed, img_embed, nil)
	return
}

func Render(gallery, thumb, path, url, id string) (out string) {
	if url[len(url)-1:] == "/" {
		url = url[:len(url)-1]
	}

	images := lsImages(path)
	if len(images) == 0 {
		return
	}

	out = strings.Replace(gallery, "<_image|url>", url+"/"+EMBEDS+"/"+images[0], -1)
	out = strings.Replace(out, "<_image|full>", url+"/"+images[0], -1)

	var (
		thumb_staging string
		thumbs_out    string
		image_new     string
		js            string
	)
	//image_new = ""
	for i := 0; i < len(images); i++ {
		thumb_staging = strings.Replace(thumb, "<_thumb|url>", url+"/"+THUMBS+"/"+images[i], -1)
		thumb_staging = strings.Replace(thumb_staging, "<_thumb|new>", image_new, -1)
		thumb_staging = strings.Replace(thumb_staging, "<_thumb|id>", strconv.Itoa(i), -1)
		thumb_staging = strings.Replace(thumb_staging, "<_image|url>", url+"/"+EMBEDS+"/"+images[i], -1)
		thumb_staging = strings.Replace(thumb_staging, "<_image|full>", url+"/"+images[i], -1)
		thumb_staging = strings.Replace(thumb_staging, "<_image|new>", image_new, -1)
		js += `"` + images[i] + `",`
		thumbs_out += thumb_staging
	}

	out = strings.Replace(out, "<_thumbs>", thumbs_out, -1)
	out = strings.Replace(out, "<_gallery|id>", id, -1)
	out = strings.Replace(out, "<_gallery|url>", url, -1)
	out += fmt.Sprintf("<script type=\"text/javascript\">var g%s_path='%s';var i_g%s=0;var g%s=[%s];</script>\n", id, url+"/"+EMBEDS+"/", id, id, js[:len(js)-1])

	return
}
