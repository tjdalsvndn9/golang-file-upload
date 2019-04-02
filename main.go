package main


import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"mime/multipart"
	"sync"
)
type message struct{
 Message string `json:"message"`
}

func main(){
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/upload", uploadHandler).Methods("POST")
	r.HandleFunc("/upload-p", uploadPHandler).Methods("POST")
	log.Fatal(http.ListenAndServe(":7000", r))
}


func homeHandler(w http.ResponseWriter, r *http.Request){
	m := message{
		Message:"what is up brother",
	}

	json.NewEncoder(w).Encode(m)
}	


func uploadHandler(w http.ResponseWriter, r *http.Request){
	//parse input from front end (Restful API) multipart/form-data
	err := r.ParseMultipartForm(10 << 20)
	if err != nil{
		http.Error(w, "was not able to parse the form", http.StatusBadRequest)
		return
	}

	//retrieve file from posted form-data
	file, fileHeader, err := r.FormFile("image")
	if err != nil{
		http.Error(w, "was not able to retrieve file from the multipart form", http.StatusBadRequest)
		return
	}
	defer file.Close()
	fmt.Println("upload file",fileHeader.Filename)
	fmt.Println("upload file size",fileHeader.Size)
	fmt.Println("upload file type",fileHeader.Header)
	//write temporary file on our server
	tempFile, err := ioutil.TempFile("temp-images", "upload-*.png")
	if err != nil{
		http.Error(w, "was not able to save the  file from in the memory", http.StatusBadRequest)
		return
	}
	defer tempFile.Close()
	fileByte, err := ioutil.ReadAll(file)
	if err != nil{
		http.Error(w, "was not able to read", http.StatusBadRequest)
		return
	}
	tempFile.Write(fileByte)
	//return whether or not this has been succesful
	m := message{
		Message:"we saved the file",
	}
	json.NewEncoder(w).Encode(m)
}

func saveFile(file *multipart.FileHeader) string{
	a,err := file.Open()
	if err != nil{
			fmt.Println("was not able to open file")
	}
	defer a.Close()
	temFile, err := ioutil.TempFile("temp-images", "upload-*.png")
	if err != nil{
		fmt.Println("failed to save file")
	}
	defer temFile.Close()
	bs, err := ioutil.ReadAll(a)
	if err != nil{
		fmt.Println("failed to save file")
	}
	_, err = temFile.Write(bs)
	if err != nil{
		fmt.Println("failed to save file")
	}

	return "saved"
}


func saveInTemp(done chan bool, files <-chan *multipart.FileHeader) <-chan string{
	a := make(chan string)
	go func(){
		defer close(a)
		for file := range files{
			select{
			case <-done:
				return
			case a <- saveFile(file):
			}
		}
	}()

	return a
}


func merge(done chan bool, workers ...<-chan string) <-chan string{
	var wg sync.WaitGroup
	man :=make(chan string)
	wg.Add(len(workers))

	multiplex :=func(c <-chan string){
		defer wg.Done()
		for s :=range c{
			select{
			case <-done:
				return
			case man <- s:	
			}
		}
	}	

	for _, c := range workers{
		go multiplex(c)
	}

	go func(){
		wg.Wait()
		close(man)
	}()

	return man
}

func createImages(done chan bool, fhs []*multipart.FileHeader) <-chan *multipart.FileHeader{
	files := make(chan *multipart.FileHeader)
	go func(){
		defer close(files)
		for _, image:=range fhs{
			select{
			case <-done:
				return
			case files <- image:
			}
		}
	}()

	return files
}



func uploadPHandler(w http.ResponseWriter, r *http.Request){

	done := make(chan bool)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil{
		 http.Error(w, "was not able to parseMultiForm", http.StatusBadRequest)
		 return
	}
	fhs := r.MultipartForm.File["image"]
	files := createImages(done,fhs)
	workers := make([]<-chan string, len(fhs))
	for i:= 0; i < len(fhs); i++{
		workers[i] = saveInTemp(done,files)
	}
	total := 0
	for range merge(done, workers...){
		total++
	}
	m := message{
		Message:fmt.Sprintf("we save total %v files", total),
	}
 json.NewEncoder(w).Encode(m)

	// for _, image :=range fhs{
	// 	file,err := image.Open()
	// 	if err != nil{
	// 		http.Error(w, "was not able to do this", http.StatusBadRequest)
	// 		break
	// 	}
	// 	defer file.Close()
	// 	t := strings.Split(image.Header["Content-Type"][0], "/")[1]
	// 	temFile, err := ioutil.TempFile("temp-images", "upload-*." + t)
	// 	if err != nil{
	// 		http.Error(w, "was not able to do this", http.StatusBadRequest)
	// 		break
	// 	}
	// 	defer temFile.Close()
	// 	bs, err := ioutil.ReadAll(file)
	// 	if err != nil{
	// 		http.Error(w, "was not able to do this", http.StatusBadRequest)
	// 		break
	// 	}
	// 	_, err = temFile.Write(bs)
	// 	if err != nil{
	// 		http.Error(w, "was not able to do this", http.StatusBadRequest)
	// 		break
	// 	}
	// }

 }


 
 