package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/taubyte/go-sdk/event"
	http "github.com/taubyte/go-sdk/http/event"
	"github.com/taubyte/go-sdk/storage"
)

func failed(h http.Event, err error, code int) uint32 {
	h.Write([]byte(err.Error()))
	h.Return(code)
	return 1
}

func setCORSHeaders(h http.Event) {
	h.Headers().Set("Access-Control-Allow-Origin", "*")
	h.Headers().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	h.Headers().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}



type UploadReq struct {
	Filename string `json:"filename"`
	Data     string `json:"data"`
}

// POST /api/upload
//export upload
func upload(e event.Event) uint32 {
	h, err := e.HTTP()
	if err != nil {
		return 1
	}

	setCORSHeaders(h)

	// Open/Create the storage
	sto, err := storage.New("pastebin")
	if err != nil {
		return failed(h, err, 500)
	}

	// Read the request body
	reqDec := json.NewDecoder(h.Body())
	defer h.Body().Close()

	var req UploadReq
	err = reqDec.Decode(&req)
	if err != nil {
		return failed(h, err, 500)
	}

	// Select file/object
	file := sto.File(req.Filename)

	// Convert text data to bytes
	fileData := []byte(req.Data)

	// Write data to the file using Add method
	_, err = file.Add(fileData, true)
	if err != nil {
		return failed(h, err, 500)
	}

	h.Write([]byte("File uploaded successfully"))
	h.Return(200)
	return 0
}

// GET /api/download?filename={filename}
//export download
func download(e event.Event) uint32 {
	h, err := e.HTTP()
	if err != nil {
		return 1
	}

	setCORSHeaders(h)

	// Read the filename from the query string
	filename, err := h.Query().Get("filename")
	if err != nil {
		return failed(h, err, 400)
	}

	// Open/Create the storage
	sto, err := storage.New("pastebin")
	if err != nil {
		return failed(h, err, 500)
	}

	// Select file/object
	file := sto.File(filename)

	// Get the file content using GetFile method
	reader, err := file.GetFile()
	if err != nil {
		return failed(h, err, 404) // File not found
	}
	defer reader.Close()

	// Read the file content
	fileContent, err := io.ReadAll(reader)
	if err != nil {
		return failed(h, err, 500)
	}

	// Write the binary data to response
	h.Write(fileContent)
	h.Return(200)
	return 0
}

// GET /api/list
//export listFiles
func listFiles(e event.Event) uint32 {
	h, err := e.HTTP()
	if err != nil {
		return 1
	}

	setCORSHeaders(h)

	// Open/Create the storage
	sto, err := storage.New("pastebin")
	if err != nil {
		return failed(h, err, 500)
	}

	// List all files
	files, err := sto.ListFiles()
	if err != nil {
		// Return empty array on error
		emptyArray := []string{}
		filesJson, _ := json.Marshal(emptyArray)
		h.Write(filesJson)
		h.Return(200)
		return 0
	}

	// Convert file objects to strings and extract filenames
	var fileNames []string
	for _, file := range files {
		fileStr := fmt.Sprintf("%v", file)
		// Extract filename from format like "{0 Group 1.png 1}"
		// Remove the braces and split by spaces
		cleanStr := strings.Trim(fileStr, "{}")
		parts := strings.Fields(cleanStr)
		if len(parts) >= 3 {
			// Join all parts except first and last (which are metadata)
			filename := strings.Join(parts[1:len(parts)-1], " ")
			fileNames = append(fileNames, filename)
		} else {
			// Fallback to original string if parsing fails
			fileNames = append(fileNames, fileStr)
		}
	}

	// Return list of filenames
	filesJson, err := json.Marshal(fileNames)
	if err != nil {
		emptyArray := []string{}
		filesJson, _ = json.Marshal(emptyArray)
	}

	h.Write(filesJson)
	h.Return(200)
	return 0
}

